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

// SeerrCollector polls Seerr for recent requests.
type SeerrCollector struct {
	cfg    *config.Config
	store  *store.Store
	client *http.Client
}

func NewSeerrCollector(cfg *config.Config, s *store.Store) *SeerrCollector {
	return &SeerrCollector{
		cfg:    cfg,
		store:  s,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (s *SeerrCollector) Name() string { return "seerr" }

func (s *SeerrCollector) Collect(ctx context.Context) error {
	if s.cfg.SeerrAPIKey == "" {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET",
		s.cfg.SeerrURL+"/api/v1/request?take=5&sort=added&sortDirection=desc", nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Api-Key", s.cfg.SeerrAPIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("seerr requests: status %d", resp.StatusCode)
	}

	var result struct {
		Results []struct {
			Type      string `json:"type"`
			Status    int    `json:"status"`
			CreatedAt string `json:"createdAt"`
			Media     struct {
				MediaType string `json:"mediaType"`
			} `json:"media"`
			RequestedBy struct {
				DisplayName string `json:"displayName"`
				Username    string `json:"username"`
			} `json:"requestedBy"`
			// For movies/tv, the title might be in the media info
			// We need to get it from the embedded media details
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	var requests []models.MediaRequest
	for _, r := range result.Results {
		user := r.RequestedBy.DisplayName
		if user == "" {
			user = r.RequestedBy.Username
		}

		status := "Unknown"
		switch r.Status {
		case 1:
			status = "Pending"
		case 2:
			status = "Approved"
		case 3:
			status = "Declined"
		case 4, 5:
			status = "Available"
		}

		mediaType := r.Media.MediaType
		if mediaType == "" {
			mediaType = r.Type
		}

		createdAt, _ := time.Parse(time.RFC3339, r.CreatedAt)

		requests = append(requests, models.MediaRequest{
			Title:       fmt.Sprintf("%s request", mediaType),
			Type:        mediaType,
			Status:      status,
			User:        user,
			RequestedAt: createdAt,
		})
	}

	s.store.UpdateRequests(requests)
	return nil
}
