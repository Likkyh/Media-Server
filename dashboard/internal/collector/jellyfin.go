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

// JellyfinSessionCollector polls Jellyfin for active sessions.
type JellyfinSessionCollector struct {
	cfg    *config.Config
	store  *store.Store
	client *http.Client
}

func NewJellyfinSessionCollector(cfg *config.Config, s *store.Store) *JellyfinSessionCollector {
	return &JellyfinSessionCollector{
		cfg:    cfg,
		store:  s,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (j *JellyfinSessionCollector) Name() string { return "jellyfin" }

func (j *JellyfinSessionCollector) Collect(ctx context.Context) error {
	if j.cfg.JellyfinAPIKey == "" {
		return nil
	}

	sessions, err := j.getSessions(ctx)
	if err != nil {
		return err
	}

	var streams []models.StreamSession
	for _, s := range sessions {
		if s.NowPlayingItem.Name == "" {
			continue
		}

		stream := models.StreamSession{
			User:   s.UserName,
			Client: s.Client,
			Device: s.DeviceName,
		}

		item := s.NowPlayingItem
		if item.SeriesName != "" {
			stream.Series = item.SeriesName
			stream.Episode = fmt.Sprintf("S%02dE%02d", item.ParentIndexNumber, item.IndexNumber)
			stream.Title = item.Name
		} else {
			stream.Title = item.Name
		}

		if s.PlayState.PositionTicks > 0 && item.RunTimeTicks > 0 {
			stream.Progress = float64(s.PlayState.PositionTicks) / float64(item.RunTimeTicks) * 100
		}

		if s.TranscodingInfo.IsTranscoding {
			stream.Transcoding = true
			stream.PlayMethod = "Transcode"
		} else if s.PlayState.PlayMethod == "DirectPlay" {
			stream.PlayMethod = "Direct Play"
		} else {
			stream.PlayMethod = s.PlayState.PlayMethod
		}

		streams = append(streams, stream)
	}

	j.store.UpdateStreams(streams)
	return nil
}

func (j *JellyfinSessionCollector) getSessions(ctx context.Context) ([]jellyfinSession, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", j.cfg.JellyfinURL+"/Sessions", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Emby-Token", j.cfg.JellyfinAPIKey)

	resp, err := j.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("jellyfin sessions: status %d", resp.StatusCode)
	}

	var sessions []jellyfinSession
	return sessions, json.NewDecoder(resp.Body).Decode(&sessions)
}

type jellyfinSession struct {
	UserName   string `json:"UserName"`
	Client     string `json:"Client"`
	DeviceName string `json:"DeviceName"`
	PlayState  struct {
		PositionTicks int64  `json:"PositionTicks"`
		PlayMethod    string `json:"PlayMethod"`
	} `json:"PlayState"`
	NowPlayingItem struct {
		Name              string `json:"Name"`
		SeriesName        string `json:"SeriesName"`
		ParentIndexNumber int    `json:"ParentIndexNumber"`
		IndexNumber       int    `json:"IndexNumber"`
		RunTimeTicks      int64  `json:"RunTimeTicks"`
	} `json:"NowPlayingItem"`
	TranscodingInfo struct {
		IsTranscoding bool `json:"IsTranscoding"`
	} `json:"TranscodingInfo"`
}
