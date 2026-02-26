package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sort"
	"strings"
	"time"

	"arcticmon/internal/config"
	"arcticmon/internal/models"
	"arcticmon/internal/store"
)

// QbitTransferCollector polls qBittorrent for transfer stats and torrents.
type QbitTransferCollector struct {
	cfg    *config.Config
	store  *store.Store
	client *http.Client
	authed bool
}

func NewQbitTransferCollector(cfg *config.Config, s *store.Store) *QbitTransferCollector {
	jar, _ := cookiejar.New(nil)
	return &QbitTransferCollector{
		cfg:   cfg,
		store: s,
		client: &http.Client{
			Timeout: 5 * time.Second,
			Jar:     jar,
		},
	}
}

func (q *QbitTransferCollector) Name() string { return "qbittorrent" }

func (q *QbitTransferCollector) Collect(ctx context.Context) error {
	if q.cfg.QbitPassword == "" {
		return nil
	}

	if !q.authed {
		if err := q.login(ctx); err != nil {
			return err
		}
	}

	data := models.TorrentData{}

	// Transfer info
	transfer, err := q.getJSON(ctx, "/api/v2/transfer/info")
	if err != nil {
		q.authed = false
		return err
	}
	if m, ok := transfer.(map[string]any); ok {
		data.DLSpeed = jsonUint64(m["dl_info_speed"])
		data.UPSpeed = jsonUint64(m["up_info_speed"])
	}

	// Torrents
	torrentsRaw, err := q.getJSON(ctx, "/api/v2/torrents/info")
	if err != nil {
		q.authed = false
		return err
	}

	if torrents, ok := torrentsRaw.([]any); ok {
		data.TotalCount = len(torrents)
		var allRatio float64
		var ratioCount int
		var active []models.TorrentBrief

		for _, t := range torrents {
			m, ok := t.(map[string]any)
			if !ok {
				continue
			}

			state := jsonStr(m["state"])
			ratio := jsonFloat(m["ratio"])
			if ratio > 0 {
				allRatio += ratio
				ratioCount++
			}

			switch state {
			case "uploading", "stalledUP", "forcedUP", "queuedUP", "checkingUP":
				data.SeedingCount++
			}

			dlSpeed := jsonUint64(m["dlspeed"])
			upSpeed := jsonUint64(m["upspeed"])
			progress := jsonFloat(m["progress"])

			if dlSpeed > 0 || upSpeed > 1024 || (progress < 1 && progress > 0) {
				data.ActiveCount++
				active = append(active, models.TorrentBrief{
					Name:     jsonStr(m["name"]),
					State:    state,
					Progress: progress * 100,
					DLSpeed:  dlSpeed,
					UPSpeed:  upSpeed,
					Size:     jsonUint64(m["size"]),
					Ratio:    ratio,
				})
			}
		}

		if ratioCount > 0 {
			data.Ratio = allRatio / float64(ratioCount)
		}

		// Sort active by download speed descending, take top 10
		sort.Slice(active, func(i, j int) bool {
			return active[i].DLSpeed > active[j].DLSpeed
		})
		if len(active) > 10 {
			active = active[:10]
		}
		data.TopTorrents = active
	}

	q.store.UpdateTorrents(data)
	return nil
}

func (q *QbitTransferCollector) login(ctx context.Context) error {
	form := url.Values{
		"username": {q.cfg.QbitUsername},
		"password": {q.cfg.QbitPassword},
	}
	req, err := http.NewRequestWithContext(ctx, "POST",
		q.cfg.QbitURL+"/api/v2/auth/login",
		strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := q.client.Do(req)
	if err != nil {
		return fmt.Errorf("qbit login: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("qbit login: status %d", resp.StatusCode)
	}
	q.authed = true
	return nil
}

func (q *QbitTransferCollector) getJSON(ctx context.Context, path string) (any, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", q.cfg.QbitURL+path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := q.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 {
		q.authed = false
		if err := q.login(ctx); err != nil {
			return nil, err
		}
		return q.getJSON(ctx, path)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("qbit %s: status %d", path, resp.StatusCode)
	}

	var result any
	return result, json.NewDecoder(resp.Body).Decode(&result)
}

func jsonStr(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func jsonFloat(v any) float64 {
	if f, ok := v.(float64); ok {
		return f
	}
	return 0
}

func jsonUint64(v any) uint64 {
	if f, ok := v.(float64); ok {
		return uint64(f)
	}
	return 0
}
