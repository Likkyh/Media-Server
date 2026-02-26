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

// ProwlarrCollector polls Prowlarr for indexer health.
type ProwlarrCollector struct {
	cfg    *config.Config
	store  *store.Store
	client *http.Client
}

func NewProwlarrCollector(cfg *config.Config, s *store.Store) *ProwlarrCollector {
	return &ProwlarrCollector{
		cfg:    cfg,
		store:  s,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (p *ProwlarrCollector) Name() string { return "prowlarr" }

func (p *ProwlarrCollector) Collect(ctx context.Context) error {
	if p.cfg.ProwlarrAPIKey == "" {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET",
		p.cfg.ProwlarrURL+"/api/v1/health", nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Api-Key", p.cfg.ProwlarrAPIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("prowlarr health: status %d", resp.StatusCode)
	}

	var items []struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return err
	}

	existing := p.store.Get().Health
	var others []models.HealthWarning
	for _, h := range existing {
		if h.Source != "Prowlarr" {
			others = append(others, h)
		}
	}

	var warnings []models.HealthWarning
	for _, item := range items {
		warnings = append(warnings, models.HealthWarning{
			Source:  "Prowlarr",
			Type:    item.Type,
			Message: item.Message,
		})
	}

	p.store.UpdateHealth(append(others, warnings...))
	return nil
}
