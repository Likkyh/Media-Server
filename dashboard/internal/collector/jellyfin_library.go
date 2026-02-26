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

// JellyfinLibraryCollector polls Jellyfin for library item counts.
type JellyfinLibraryCollector struct {
	cfg    *config.Config
	store  *store.Store
	client *http.Client
}

func NewJellyfinLibraryCollector(cfg *config.Config, s *store.Store) *JellyfinLibraryCollector {
	return &JellyfinLibraryCollector{
		cfg:    cfg,
		store:  s,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (j *JellyfinLibraryCollector) Name() string { return "jellyfin-library" }

func (j *JellyfinLibraryCollector) Collect(ctx context.Context) error {
	if j.cfg.JellyfinAPIKey == "" {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET",
		j.cfg.JellyfinURL+"/Items/Counts", nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Emby-Token", j.cfg.JellyfinAPIKey)

	resp, err := j.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("jellyfin counts: status %d", resp.StatusCode)
	}

	var counts struct {
		MovieCount   int `json:"MovieCount"`
		SeriesCount  int `json:"SeriesCount"`
		EpisodeCount int `json:"EpisodeCount"`
		MusicCount   int `json:"SongCount"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&counts); err != nil {
		return err
	}

	j.store.UpdateLibrary(models.LibraryCounts{
		Movies:   counts.MovieCount,
		Series:   counts.SeriesCount,
		Episodes: counts.EpisodeCount,
		Music:    counts.MusicCount,
	})
	return nil
}
