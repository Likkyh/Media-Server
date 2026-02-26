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

// BazarrCollector polls Bazarr for missing subtitles.
type BazarrCollector struct {
	cfg    *config.Config
	store  *store.Store
	client *http.Client
}

func NewBazarrCollector(cfg *config.Config, s *store.Store) *BazarrCollector {
	return &BazarrCollector{
		cfg:    cfg,
		store:  s,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (b *BazarrCollector) Name() string { return "bazarr" }

func (b *BazarrCollector) Collect(ctx context.Context) error {
	if b.cfg.BazarrAPIKey == "" {
		return nil
	}

	movieCount, err := b.getWantedCount(ctx, "/api/movies/wanted")
	if err != nil {
		movieCount = 0
	}

	episodeCount, err := b.getWantedCount(ctx, "/api/episodes/wanted")
	if err != nil {
		episodeCount = 0
	}

	existing := b.store.Get().Health
	var others []models.HealthWarning
	for _, h := range existing {
		if h.Source != "Bazarr" {
			others = append(others, h)
		}
	}

	total := movieCount + episodeCount
	if total > 0 {
		others = append(others, models.HealthWarning{
			Source:  "Bazarr",
			Type:    "warning",
			Message: fmt.Sprintf("%d missing subtitles (%d movies, %d episodes)", total, movieCount, episodeCount),
		})
	}

	b.store.UpdateHealth(others)
	return nil
}

func (b *BazarrCollector) getWantedCount(ctx context.Context, path string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		b.cfg.BazarrURL+path+"?start=0&length=1", nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("X-Api-Key", b.cfg.BazarrAPIKey)

	resp, err := b.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("bazarr %s: status %d", path, resp.StatusCode)
	}

	var result struct {
		Total int `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	return result.Total, nil
}
