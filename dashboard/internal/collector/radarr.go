package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"arcticmon/internal/config"
	"arcticmon/internal/models"
	"arcticmon/internal/store"
)

// RadarrCollector polls Radarr for queue and health.
type RadarrCollector struct {
	cfg    *config.Config
	store  *store.Store
	client *http.Client
}

func NewRadarrCollector(cfg *config.Config, s *store.Store) *RadarrCollector {
	return &RadarrCollector{
		cfg:    cfg,
		store:  s,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (r *RadarrCollector) Name() string { return "radarr" }

func (r *RadarrCollector) Collect(ctx context.Context) error {
	if r.cfg.RadarrAPIKey == "" {
		return nil
	}

	downloads, err := r.getQueue(ctx)
	if err != nil {
		return err
	}

	// Merge with existing downloads (from sonarr/sabnzbd)
	existing := r.store.Get().Downloads
	var others []models.DownloadItem
	for _, d := range existing {
		if d.Source != "Radarr" {
			others = append(others, d)
		}
	}
	r.store.UpdateDownloads(append(others, downloads...))

	// Health warnings
	health, err := r.getHealth(ctx)
	if err != nil {
		return err
	}
	existingHealth := r.store.Get().Health
	var otherHealth []models.HealthWarning
	for _, h := range existingHealth {
		if h.Source != "Radarr" {
			otherHealth = append(otherHealth, h)
		}
	}
	r.store.UpdateHealth(append(otherHealth, health...))

	return nil
}

func (r *RadarrCollector) getQueue(ctx context.Context) ([]models.DownloadItem, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		r.cfg.RadarrURL+"/api/v3/queue?pageSize=50&includeMovie=true", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Api-Key", r.cfg.RadarrAPIKey)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("radarr queue: status %d", resp.StatusCode)
	}

	var result struct {
		Records []struct {
			Title   string  `json:"title"`
			Status  string  `json:"status"`
			Size    float64 `json:"size"`
			Sizeleft float64 `json:"sizeleft"`
			Timeleft string  `json:"timeleft"`
		} `json:"records"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var items []models.DownloadItem
	for _, rec := range result.Records {
		progress := 0.0
		if rec.Size > 0 {
			progress = (1 - rec.Sizeleft/rec.Size) * 100
		}
		items = append(items, models.DownloadItem{
			Title:    rec.Title,
			Source:   "Radarr",
			Status:   rec.Status,
			Progress: progress,
			Size:     uint64(rec.Size),
			Timeleft: rec.Timeleft,
		})
	}
	return items, nil
}

func (r *RadarrCollector) getHealth(ctx context.Context) ([]models.HealthWarning, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		r.cfg.RadarrURL+"/api/v3/health", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Api-Key", r.cfg.RadarrAPIKey)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("radarr health: status %d", resp.StatusCode)
	}

	var items []struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, err
	}

	var warnings []models.HealthWarning
	for _, item := range items {
		warnings = append(warnings, models.HealthWarning{
			Source:  "Radarr",
			Type:    item.Type,
			Message: item.Message,
		})
	}
	return warnings, nil
}
