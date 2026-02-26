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

// UnmanicCollector polls Unmanic for worker status and pending tasks.
type UnmanicCollector struct {
	cfg    *config.Config
	store  *store.Store
	client *http.Client
}

func NewUnmanicCollector(cfg *config.Config, s *store.Store) *UnmanicCollector {
	return &UnmanicCollector{
		cfg:    cfg,
		store:  s,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (u *UnmanicCollector) Name() string { return "unmanic" }

func (u *UnmanicCollector) Collect(ctx context.Context) error {
	data := models.TranscodeData{}

	// Workers
	workers, err := u.getWorkers(ctx)
	if err == nil {
		data.Workers = workers
	}

	// Pending
	pending, err := u.getPending(ctx)
	if err == nil {
		data.Pending = pending
	}

	u.store.UpdateTranscodes(data)
	return nil
}

func (u *UnmanicCollector) getWorkers(ctx context.Context) ([]models.TranscodeWorker, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		u.cfg.UnmanicURL+"/unmanic/api/v2/workers/status", nil)
	if err != nil {
		return nil, err
	}

	resp, err := u.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unmanic workers: status %d", resp.StatusCode)
	}

	var result struct {
		Workers []struct {
			ID       string `json:"id"`
			Idle     bool   `json:"idle"`
			FileName string `json:"current_file"`
			Progress float64 `json:"progress"`
			FPS      float64 `json:"fps"`
			Speed    string  `json:"speed"`
		} `json:"workers"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var workers []models.TranscodeWorker
	for _, w := range result.Workers {
		status := "idle"
		if !w.Idle {
			status = "working"
		}
		workers = append(workers, models.TranscodeWorker{
			ID:       w.ID,
			FileName: w.FileName,
			Progress: w.Progress,
			FPS:      w.FPS,
			Speed:    w.Speed,
			Status:   status,
		})
	}
	return workers, nil
}

func (u *UnmanicCollector) getPending(ctx context.Context) (int, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		u.cfg.UnmanicURL+"/unmanic/api/v2/pending/tasks?start=0&length=1", nil)
	if err != nil {
		return 0, err
	}

	resp, err := u.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("unmanic pending: status %d", resp.StatusCode)
	}

	var result struct {
		RecordsTotal int `json:"recordsTotal"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	return result.RecordsTotal, nil
}
