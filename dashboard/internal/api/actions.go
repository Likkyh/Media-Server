package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"arcticmon/internal/config"
)

// Actions provides API handlers for server management actions.
type Actions struct {
	cfg    *config.Config
	docker *http.Client
}

func NewActions(cfg *config.Config) *Actions {
	return &Actions{
		cfg: cfg,
		docker: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", cfg.DockerSocket)
				},
			},
		},
	}
}

// RestartStack restarts all containers on the medianet network.
func (a *Actions) RestartStack(w http.ResponseWriter, r *http.Request) {
	containers, err := a.listContainers()
	if err != nil {
		writeError(w, 500, "Failed to list containers: "+err.Error())
		return
	}

	var restarted []string
	var errors []string

	for _, c := range containers {
		name := strings.TrimPrefix(c.Names[0], "/")
		// Don't restart ourselves or autoheal
		if name == "arcticmon" || name == "autoheal" {
			continue
		}

		req, _ := http.NewRequestWithContext(r.Context(), "POST",
			fmt.Sprintf("http://docker/containers/%s/restart?t=10", c.ID), nil)
		resp, err := a.docker.Do(req)
		if err != nil {
			errors = append(errors, name+": "+err.Error())
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == 204 {
			restarted = append(restarted, name)
		} else {
			errors = append(errors, fmt.Sprintf("%s: status %d", name, resp.StatusCode))
		}
	}

	writeJSON(w, map[string]any{
		"restarted": restarted,
		"errors":    errors,
	})
}

// RestartVM reboots the host machine via Docker exec with nsenter.
func (a *Actions) RestartVM(w http.ResponseWriter, r *http.Request) {
	// Use nsenter through a temporary container to reboot the host
	createBody := `{
		"Image": "alpine:3.19",
		"Cmd": ["nsenter", "-t", "1", "-m", "-u", "-i", "-n", "--", "reboot"],
		"HostConfig": {
			"PidMode": "host",
			"Privileged": true,
			"AutoRemove": true
		}
	}`

	req, _ := http.NewRequestWithContext(r.Context(), "POST",
		"http://docker/containers/create?name=arcticmon-reboot",
		strings.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.docker.Do(req)
	if err != nil {
		writeError(w, 500, "Failed to create reboot container: "+err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		writeError(w, 500, fmt.Sprintf("Create container: status %d", resp.StatusCode))
		return
	}

	var created struct {
		ID string `json:"Id"`
	}
	json.NewDecoder(resp.Body).Decode(&created)

	// Start the container
	startReq, _ := http.NewRequestWithContext(r.Context(), "POST",
		fmt.Sprintf("http://docker/containers/%s/start", created.ID), nil)
	startResp, err := a.docker.Do(startReq)
	if err != nil {
		writeError(w, 500, "Failed to start reboot container: "+err.Error())
		return
	}
	startResp.Body.Close()

	writeJSON(w, map[string]any{"status": "rebooting"})
}

// UpdateStack pulls latest images, recreates containers, and prunes unused images.
func (a *Actions) UpdateStack(w http.ResponseWriter, r *http.Request) {
	containers, err := a.listContainers()
	if err != nil {
		writeError(w, 500, "Failed to list containers: "+err.Error())
		return
	}

	var pulled []string
	var errors []string

	for _, c := range containers {
		name := strings.TrimPrefix(c.Names[0], "/")
		image := c.Image

		// Pull latest image
		pullReq, _ := http.NewRequestWithContext(r.Context(), "POST",
			fmt.Sprintf("http://docker/images/create?fromImage=%s", image), nil)
		pullResp, err := a.docker.Do(pullReq)
		if err != nil {
			errors = append(errors, name+": pull failed: "+err.Error())
			continue
		}
		// Read body to completion (Docker streams progress)
		buf := make([]byte, 4096)
		for {
			_, readErr := pullResp.Body.Read(buf)
			if readErr != nil {
				break
			}
		}
		pullResp.Body.Close()

		if pullResp.StatusCode == 200 {
			pulled = append(pulled, name)
		} else {
			errors = append(errors, fmt.Sprintf("%s: pull status %d", name, pullResp.StatusCode))
		}
	}

	// Prune unused images
	pruneReq, _ := http.NewRequestWithContext(r.Context(), "POST",
		"http://docker/images/prune?filters={\"dangling\":[\"false\"]}", nil)
	pruneResp, err := a.docker.Do(pruneReq)
	pruned := ""
	if err == nil {
		defer pruneResp.Body.Close()
		var pruneResult struct {
			SpaceReclaimed uint64 `json:"SpaceReclaimed"`
		}
		json.NewDecoder(pruneResp.Body).Decode(&pruneResult)
		if pruneResult.SpaceReclaimed > 0 {
			mb := pruneResult.SpaceReclaimed / (1024 * 1024)
			pruned = fmt.Sprintf("Pruned %d MB of unused images", mb)
		} else {
			pruned = "No unused images to prune"
		}
	}

	writeJSON(w, map[string]any{
		"pulled": pulled,
		"errors": errors,
		"pruned": pruned,
		"note":   "Images pulled and pruned. Restart the stack to apply updates.",
	})
}

// UpdateSystem runs apt update && apt full-upgrade -y on the host via nsenter.
func (a *Actions) UpdateSystem(w http.ResponseWriter, r *http.Request) {
	createBody := `{
		"Image": "alpine:3.19",
		"Cmd": ["nsenter", "-t", "1", "-m", "-u", "-i", "-n", "--", "sh", "-c", "apt-get update && apt-get full-upgrade -y && apt-get autoremove -y"],
		"HostConfig": {
			"PidMode": "host",
			"Privileged": true,
			"AutoRemove": true
		}
	}`

	req, _ := http.NewRequestWithContext(r.Context(), "POST",
		"http://docker/containers/create?name=arcticmon-sysupdate",
		strings.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.docker.Do(req)
	if err != nil {
		writeError(w, 500, "Failed to create update container: "+err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		writeError(w, 500, fmt.Sprintf("Create container: status %d", resp.StatusCode))
		return
	}

	var created struct {
		ID string `json:"Id"`
	}
	json.NewDecoder(resp.Body).Decode(&created)

	// Start the container
	startReq, _ := http.NewRequestWithContext(r.Context(), "POST",
		fmt.Sprintf("http://docker/containers/%s/start", created.ID), nil)
	startResp, err := a.docker.Do(startReq)
	if err != nil {
		writeError(w, 500, "Failed to start update container: "+err.Error())
		return
	}
	startResp.Body.Close()

	// Wait for container to finish (longer timeout for apt upgrade)
	waitClient := &http.Client{
		Timeout: 10 * time.Minute,
		Transport: a.docker.Transport,
	}
	waitReq, _ := http.NewRequestWithContext(r.Context(), "POST",
		fmt.Sprintf("http://docker/containers/%s/wait", created.ID), nil)
	waitResp, err := waitClient.Do(waitReq)
	if err != nil {
		writeJSON(w, map[string]any{
			"status": "running",
			"note":   "System update started but timed out waiting for completion.",
		})
		return
	}
	defer waitResp.Body.Close()

	var waitResult struct {
		StatusCode int `json:"StatusCode"`
	}
	json.NewDecoder(waitResp.Body).Decode(&waitResult)

	if waitResult.StatusCode == 0 {
		writeJSON(w, map[string]any{"status": "completed", "note": "System packages updated successfully."})
	} else {
		writeJSON(w, map[string]any{
			"status": "failed",
			"note":   fmt.Sprintf("apt upgrade exited with code %d", waitResult.StatusCode),
		})
	}
}

func (a *Actions) listContainers() ([]struct {
	ID    string   `json:"Id"`
	Names []string `json:"Names"`
	Image string   `json:"Image"`
}, error) {
	req, _ := http.NewRequest("GET",
		"http://docker/containers/json?all=false&filters={\"network\":[\"mediaserver_medianet\"]}", nil)
	resp, err := a.docker.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var containers []struct {
		ID    string   `json:"Id"`
		Names []string `json:"Names"`
		Image string   `json:"Image"`
	}
	return containers, json.NewDecoder(resp.Body).Decode(&containers)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
