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

// SonarrCollector polls Sonarr for queue and health.
type SonarrCollector struct {
	cfg    *config.Config
	store  *store.Store
	client *http.Client
}

func NewSonarrCollector(cfg *config.Config, s *store.Store) *SonarrCollector {
	return &SonarrCollector{
		cfg:    cfg,
		store:  s,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (s *SonarrCollector) Name() string { return "sonarr" }

func (s *SonarrCollector) Collect(ctx context.Context) error {
	if s.cfg.SonarrAPIKey == "" {
		return nil
	}

	downloads, err := s.getQueue(ctx)
	if err != nil {
		return err
	}

	existing := s.store.Get().Downloads
	var others []models.DownloadItem
	for _, d := range existing {
		if d.Source != "Sonarr" {
			others = append(others, d)
		}
	}
	s.store.UpdateDownloads(append(others, downloads...))

	health, err := s.getHealth(ctx)
	if err != nil {
		return err
	}
	existingHealth := s.store.Get().Health
	var otherHealth []models.HealthWarning
	for _, h := range existingHealth {
		if h.Source != "Sonarr" {
			otherHealth = append(otherHealth, h)
		}
	}
	s.store.UpdateHealth(append(otherHealth, health...))

	return nil
}

func (sc *SonarrCollector) getQueue(ctx context.Context) ([]models.DownloadItem, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		sc.cfg.SonarrURL+"/api/v3/queue?pageSize=50&includeSeries=true&includeEpisode=true", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Api-Key", sc.cfg.SonarrAPIKey)

	resp, err := sc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("sonarr queue: status %d", resp.StatusCode)
	}

	var result struct {
		Records []struct {
			Title    string  `json:"title"`
			Status   string  `json:"status"`
			Size     float64 `json:"size"`
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
			Source:   "Sonarr",
			Status:   rec.Status,
			Progress: progress,
			Size:     uint64(rec.Size),
			Timeleft: rec.Timeleft,
		})
	}
	return items, nil
}

func (sc *SonarrCollector) getHealth(ctx context.Context) ([]models.HealthWarning, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		sc.cfg.SonarrURL+"/api/v3/health", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Api-Key", sc.cfg.SonarrAPIKey)

	resp, err := sc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("sonarr health: status %d", resp.StatusCode)
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
			Source:  "Sonarr",
			Type:    item.Type,
			Message: item.Message,
		})
	}
	return warnings, nil
}
