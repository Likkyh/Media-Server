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

// SabnzbdCollector polls SABnzbd for queue status.
type SabnzbdCollector struct {
	cfg    *config.Config
	store  *store.Store
	client *http.Client
}

func NewSabnzbdCollector(cfg *config.Config, s *store.Store) *SabnzbdCollector {
	return &SabnzbdCollector{
		cfg:    cfg,
		store:  s,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (s *SabnzbdCollector) Name() string { return "sabnzbd" }

func (s *SabnzbdCollector) Collect(ctx context.Context) error {
	if s.cfg.SabnzbdAPIKey == "" {
		return nil
	}

	url := fmt.Sprintf("%s/api?mode=queue&output=json&apikey=%s", s.cfg.SabnzbdURL, s.cfg.SabnzbdAPIKey)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("sabnzbd queue: status %d", resp.StatusCode)
	}

	var result struct {
		Queue struct {
			Slots []struct {
				Filename  string `json:"filename"`
				Status    string `json:"status"`
				Mb        string `json:"mb"`
				Mbleft    string `json:"mbleft"`
				Timeleft  string `json:"timeleft"`
				Percentage string `json:"percentage"`
			} `json:"slots"`
		} `json:"queue"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	existing := s.store.Get().Downloads
	var others []models.DownloadItem
	for _, d := range existing {
		if d.Source != "SABnzbd" {
			others = append(others, d)
		}
	}

	var items []models.DownloadItem
	for _, slot := range result.Queue.Slots {
		progress := jsonFloat(parseFloat(slot.Percentage))
		items = append(items, models.DownloadItem{
			Title:    slot.Filename,
			Source:   "SABnzbd",
			Status:   slot.Status,
			Progress: progress,
			Timeleft: slot.Timeleft,
		})
	}

	s.store.UpdateDownloads(append(others, items...))
	return nil
}

func parseFloat(s string) any {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
