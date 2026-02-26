package collector

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"arcticmon/internal/config"
	"arcticmon/internal/models"
	"arcticmon/internal/store"
)

// HostCollector collects CPU, RAM, swap, disk, GPU, SSH sessions, and uptime.
type HostCollector struct {
	cfg    *config.Config
	store  *store.Store
	docker *http.Client

	prevIdle  uint64
	prevTotal uint64
	// Per-core previous values
	prevCoreIdle  map[int]uint64
	prevCoreTotal map[int]uint64
}

func NewHostCollector(cfg *config.Config, s *store.Store) *HostCollector {
	return &HostCollector{
		cfg:   cfg,
		store: s,
		docker: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", cfg.DockerSocket)
				},
			},
		},
		prevCoreIdle:  make(map[int]uint64),
		prevCoreTotal: make(map[int]uint64),
	}
}

func (h *HostCollector) Name() string { return "host" }

func (h *HostCollector) Collect(ctx context.Context) error {
	metrics := models.HostMetrics{}

	// CPU usage (aggregate + per-core)
	if cpu, cores, err := h.readCPU(); err == nil {
		metrics.CPUPercent = cpu
		metrics.CPUCores = cores
	}

	// Memory + swap
	h.readMemory(&metrics)

	// Disks (NVMe on /, HDD on /mnt/media)
	metrics.Disks = h.readDisks()

	// GPU (via nvidia-smi in jellyfin container, requires PCI passthrough)
	if gpu, err := h.readGPU(ctx); err == nil {
		metrics.GPU = gpu
	}

	// SSH sessions
	metrics.SSHSessions = h.countSSH()

	// Uptime
	if up, err := h.readUptime(); err == nil {
		metrics.Uptime = up
	}

	h.store.UpdateHost(metrics)
	return nil
}

func (h *HostCollector) readCPU() (float64, []models.CPUCore, error) {
	data, err := os.ReadFile(h.cfg.HostProcPath + "/stat")
	if err != nil {
		return 0, nil, err
	}

	var avgPercent float64
	var cores []models.CPUCore

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()

		// Aggregate "cpu " line
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			if len(fields) < 5 {
				continue
			}
			var values []uint64
			for _, f := range fields[1:] {
				v, _ := strconv.ParseUint(f, 10, 64)
				values = append(values, v)
			}
			idle := values[3]
			var total uint64
			for _, v := range values {
				total += v
			}

			if h.prevTotal == 0 {
				h.prevIdle = idle
				h.prevTotal = total
			} else {
				deltaTotal := total - h.prevTotal
				deltaIdle := idle - h.prevIdle
				h.prevIdle = idle
				h.prevTotal = total
				if deltaTotal > 0 {
					avgPercent = float64(deltaTotal-deltaIdle) / float64(deltaTotal) * 100
				}
			}
			continue
		}

		// Per-core "cpu0", "cpu1", etc.
		if strings.HasPrefix(line, "cpu") {
			fields := strings.Fields(line)
			if len(fields) < 5 {
				continue
			}
			coreIDStr := strings.TrimPrefix(fields[0], "cpu")
			coreID, err := strconv.Atoi(coreIDStr)
			if err != nil {
				continue
			}

			var values []uint64
			for _, f := range fields[1:] {
				v, _ := strconv.ParseUint(f, 10, 64)
				values = append(values, v)
			}
			idle := values[3]
			var total uint64
			for _, v := range values {
				total += v
			}

			prevIdle := h.prevCoreIdle[coreID]
			prevTotal := h.prevCoreTotal[coreID]
			h.prevCoreIdle[coreID] = idle
			h.prevCoreTotal[coreID] = total

			var pct float64
			if prevTotal > 0 {
				deltaTotal := total - prevTotal
				deltaIdle := idle - prevIdle
				if deltaTotal > 0 {
					pct = float64(deltaTotal-deltaIdle) / float64(deltaTotal) * 100
				}
			}

			cores = append(cores, models.CPUCore{
				ID:      coreID,
				Percent: pct,
			})
		}
	}

	if len(cores) == 0 {
		return 0, nil, fmt.Errorf("cpu lines not found")
	}
	return avgPercent, cores, nil
}

func (h *HostCollector) readMemory(m *models.HostMetrics) {
	data, err := os.ReadFile(h.cfg.HostProcPath + "/meminfo")
	if err != nil {
		return
	}

	info := map[string]uint64{}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			key := strings.TrimSuffix(parts[0], ":")
			val, _ := strconv.ParseUint(parts[1], 10, 64)
			info[key] = val * 1024 // convert from kB to bytes
		}
	}

	m.MemTotal = info["MemTotal"]
	available := info["MemAvailable"]
	m.MemUsed = m.MemTotal - available
	if m.MemTotal > 0 {
		m.MemPercent = float64(m.MemUsed) / float64(m.MemTotal) * 100
	}

	m.SwapTotal = info["SwapTotal"]
	swapFree := info["SwapFree"]
	m.SwapUsed = m.SwapTotal - swapFree
	if m.SwapTotal > 0 {
		m.SwapPercent = float64(m.SwapUsed) / float64(m.SwapTotal) * 100
	}
}

func (h *HostCollector) readDisks() []models.DiskInfo {
	type diskMount struct {
		mount string
		label string
	}
	targets := []diskMount{
		{mount: "/", label: "NVMe (/)"},
		{mount: "/mnt/media", label: "HDD (/mnt/media)"},
	}

	var disks []models.DiskInfo
	for _, t := range targets {
		var stat statfsResult
		// In container, / is the container root. Use /host/proc/1/root to access host FS
		path := t.mount
		if t.mount == "/" {
			// Try host root via proc
			hostRoot := h.cfg.HostProcPath + "/1/root"
			if _, err := os.Stat(hostRoot); err == nil {
				path = hostRoot
			}
		} else {
			// For /mnt/media, try via host proc first
			hostPath := h.cfg.HostProcPath + "/1/root" + t.mount
			if _, err := os.Stat(hostPath); err == nil {
				path = hostPath
			}
		}

		if err := statfs(path, &stat); err == nil && stat.Total > 0 {
			disks = append(disks, models.DiskInfo{
				Mount:   t.mount,
				Label:   t.label,
				Total:   stat.Total,
				Used:    stat.Used,
				Percent: float64(stat.Used) / float64(stat.Total) * 100,
			})
		}
	}
	return disks
}

func (h *HostCollector) readGPU(ctx context.Context) (models.GPUMetrics, error) {
	gpu := models.GPUMetrics{}

	// Create exec instance in jellyfin container
	execBody := `{"AttachStdout":true,"AttachStderr":true,"Cmd":["nvidia-smi","--query-gpu=name,temperature.gpu,utilization.gpu,memory.used,memory.total","--format=csv,noheader,nounits"]}`
	req, err := http.NewRequestWithContext(ctx, "POST",
		"http://docker/containers/jellyfin/exec",
		strings.NewReader(execBody))
	if err != nil {
		return gpu, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.docker.Do(req)
	if err != nil {
		return gpu, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return gpu, fmt.Errorf("exec create: status %d", resp.StatusCode)
	}

	var execResp struct {
		ID string `json:"Id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&execResp); err != nil {
		return gpu, err
	}

	// Start exec
	startBody := `{"Detach":false,"Tty":false}`
	req2, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("http://docker/exec/%s/start", execResp.ID),
		strings.NewReader(startBody))
	if err != nil {
		return gpu, err
	}
	req2.Header.Set("Content-Type", "application/json")

	resp2, err := h.docker.Do(req2)
	if err != nil {
		return gpu, err
	}
	defer resp2.Body.Close()

	// Read output (Docker multiplexed stream: 8-byte header + payload)
	buf := make([]byte, 4096)
	n, _ := resp2.Body.Read(buf)
	output := string(buf[:n])

	// Strip Docker stream header bytes if present
	if len(output) > 8 && (output[0] == 1 || output[0] == 2) {
		output = output[8:]
	}

	output = strings.TrimSpace(output)
	parts := strings.Split(output, ", ")
	if len(parts) >= 5 {
		gpu.Available = true
		gpu.Name = strings.TrimSpace(parts[0])
		gpu.TempC, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
		gpu.UtilPercent, _ = strconv.Atoi(strings.TrimSpace(parts[2]))
		memUsed, _ := strconv.ParseUint(strings.TrimSpace(parts[3]), 10, 64)
		memTotal, _ := strconv.ParseUint(strings.TrimSpace(parts[4]), 10, 64)
		gpu.MemUsed = memUsed * 1024 * 1024
		gpu.MemTotal = memTotal * 1024 * 1024
		if gpu.MemTotal > 0 {
			gpu.MemPercent = float64(gpu.MemUsed) / float64(gpu.MemTotal) * 100
		}
	}

	return gpu, nil
}

func (h *HostCollector) countSSH() int {
	data, err := os.ReadFile(h.cfg.HostUtmpPath)
	if err != nil {
		return 0
	}

	count := 0
	recordSize := 384
	for i := 0; i+recordSize <= len(data); i += recordSize {
		record := data[i : i+recordSize]
		utType := binary.LittleEndian.Uint32(record[0:4])
		if utType == 7 {
			host := string(bytes.TrimRight(record[76:332], "\x00"))
			if host != "" && host != ":0" && !strings.HasPrefix(host, ":") {
				count++
			}
		}
	}
	return count
}

func (h *HostCollector) readUptime() (string, error) {
	data, err := os.ReadFile(h.cfg.HostProcPath + "/uptime")
	if err != nil {
		return "", err
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return "", fmt.Errorf("empty uptime")
	}
	secs, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return "", err
	}
	dur := time.Duration(secs) * time.Second
	days := int(dur.Hours() / 24)
	hours := int(dur.Hours()) % 24
	mins := int(dur.Minutes()) % 60
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins), nil
	}
	return fmt.Sprintf("%dh %dm", hours, mins), nil
}
