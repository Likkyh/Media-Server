package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"arcticmon/internal/config"
	"arcticmon/internal/models"
	"arcticmon/internal/store"
)

// DockerCollector collects container health/status via the Docker socket.
type DockerCollector struct {
	cfg    *config.Config
	store  *store.Store
	client *http.Client
}

func NewDockerCollector(cfg *config.Config, s *store.Store) *DockerCollector {
	return &DockerCollector{
		cfg:   cfg,
		store: s,
		client: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", cfg.DockerSocket)
				},
			},
		},
	}
}

func (d *DockerCollector) Name() string { return "docker" }

func (d *DockerCollector) Collect(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://docker/containers/json?all=true", nil)
	if err != nil {
		return err
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("docker socket: %w", err)
	}
	defer resp.Body.Close()

	var containers []dockerContainer
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return err
	}

	// Map of known services to their external URLs
	urlMap := map[string]string{
		"jellyfin":   "https://jellyfin.local.example.com",
		"radarr":     "https://radarr.local.example.com",
		"sonarr":     "https://sonarr.local.example.com",
		"seerr":      "https://seerr.local.example.com",
		"prowlarr":   "https://prowlarr.local.example.com",
		"bazarr":     "https://bazarr.local.example.com",
		"sabnzbd":    "https://sabnzbd.local.example.com",
		"unmanic":    "https://unmanic.local.example.com",
		"npm":        "https://npm.local.example.com",
		"arcticmon":  "https://dashboard.local.example.com",
	}

	var svcs []models.ServiceStatus
	for _, c := range containers {
		name := strings.TrimPrefix(c.Names[0], "/")
		health := "none"
		if c.State == "running" && c.Status != "" {
			if strings.Contains(c.Status, "(healthy)") {
				health = "healthy"
			} else if strings.Contains(c.Status, "(unhealthy)") {
				health = "unhealthy"
			} else if strings.Contains(c.Status, "(health: starting)") {
				health = "starting"
			} else {
				health = "running"
			}
		}

		ip := ""
		if c.NetworkSettings.Networks != nil {
			for _, net := range c.NetworkSettings.Networks {
				if net.IPAddress != "" {
					ip = net.IPAddress
					break
				}
			}
		}

		// Parse image name for display
		image := c.Image
		if parts := strings.Split(image, "/"); len(parts) > 0 {
			image = parts[len(parts)-1]
		}

		uptime := formatContainerUptime(c.Created)

		svcs = append(svcs, models.ServiceStatus{
			Name:        name,
			Status:      c.State,
			Health:      health,
			Image:       image,
			IP:          ip,
			Uptime:      uptime,
			ExternalURL: urlMap[name],
		})
	}

	d.store.UpdateServices(svcs)
	return nil
}

func formatContainerUptime(created int64) string {
	t := time.Unix(created, 0)
	dur := time.Since(t)
	switch {
	case dur < time.Hour:
		return fmt.Sprintf("%dm", int(dur.Minutes()))
	case dur < 24*time.Hour:
		return fmt.Sprintf("%dh", int(dur.Hours()))
	default:
		return fmt.Sprintf("%dd", int(dur.Hours()/24))
	}
}

type dockerContainer struct {
	ID      string   `json:"Id"`
	Names   []string `json:"Names"`
	Image   string   `json:"Image"`
	State   string   `json:"State"`
	Status  string   `json:"Status"`
	Created int64    `json:"Created"`
	NetworkSettings struct {
		Networks map[string]struct {
			IPAddress string `json:"IPAddress"`
		} `json:"Networks"`
	} `json:"NetworkSettings"`
}
