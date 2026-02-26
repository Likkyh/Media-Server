package store

import (
	"encoding/json"
	"sync"

	"arcticmon/internal/models"
)

// Store holds the dashboard state with thread-safe access and SSE fan-out.
type Store struct {
	mu   sync.RWMutex
	data models.DashboardData

	subsMu  sync.Mutex
	subs    map[chan []byte]struct{}
}

// New creates a new Store.
func New() *Store {
	return &Store{
		subs: make(map[chan []byte]struct{}),
	}
}

// Get returns a snapshot of the current dashboard data.
func (s *Store) Get() models.DashboardData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data
}

// UpdateHost updates host metrics and notifies SSE subscribers.
func (s *Store) UpdateHost(h models.HostMetrics) {
	s.mu.Lock()
	s.data.Host = h
	s.mu.Unlock()
	s.notify("host", h)
}

// UpdateServices updates service statuses.
func (s *Store) UpdateServices(svcs []models.ServiceStatus) {
	s.mu.Lock()
	s.data.Services = svcs
	s.mu.Unlock()
	s.notify("services", svcs)
}

// UpdateStreams updates Jellyfin stream sessions.
func (s *Store) UpdateStreams(ss []models.StreamSession) {
	s.mu.Lock()
	s.data.Streams = ss
	s.mu.Unlock()
	s.notify("streams", ss)
}

// UpdateTorrents updates torrent data.
func (s *Store) UpdateTorrents(t models.TorrentData) {
	s.mu.Lock()
	s.data.Torrents = t
	s.mu.Unlock()
	s.notify("torrents", t)
}

// UpdateDownloads updates download queue items.
func (s *Store) UpdateDownloads(d []models.DownloadItem) {
	s.mu.Lock()
	s.data.Downloads = d
	s.mu.Unlock()
	s.notify("downloads", d)
}

// UpdateRequests updates media requests.
func (s *Store) UpdateRequests(r []models.MediaRequest) {
	s.mu.Lock()
	s.data.Requests = r
	s.mu.Unlock()
	s.notify("requests", r)
}

// UpdateTranscodes updates transcode worker data.
func (s *Store) UpdateTranscodes(t models.TranscodeData) {
	s.mu.Lock()
	s.data.Transcodes = t
	s.mu.Unlock()
	s.notify("transcodes", t)
}

// UpdateLibrary updates library counts.
func (s *Store) UpdateLibrary(l models.LibraryCounts) {
	s.mu.Lock()
	s.data.Library = l
	s.mu.Unlock()
	s.notify("library", l)
}

// UpdateSSHSecurity updates SSH security data.
func (s *Store) UpdateSSHSecurity(d models.SSHSecurityData) {
	s.mu.Lock()
	s.data.SSHSecurity = d
	s.mu.Unlock()
	s.notify("sshSecurity", d)
}

// UpdateHealth updates health warnings.
func (s *Store) UpdateHealth(h []models.HealthWarning) {
	s.mu.Lock()
	s.data.Health = h
	s.mu.Unlock()
	s.notify("health", h)
}

const maxSubscribers = 20

// Subscribe returns a channel that receives SSE event payloads.
// Returns nil if the max subscriber limit is reached.
func (s *Store) Subscribe() chan []byte {
	s.subsMu.Lock()
	defer s.subsMu.Unlock()
	if len(s.subs) >= maxSubscribers {
		return nil
	}
	ch := make(chan []byte, 64)
	s.subs[ch] = struct{}{}
	return ch
}

// Unsubscribe removes a subscriber channel.
func (s *Store) Unsubscribe(ch chan []byte) {
	s.subsMu.Lock()
	delete(s.subs, ch)
	s.subsMu.Unlock()
	close(ch)
}

// notify sends a JSON event to all SSE subscribers.
func (s *Store) notify(event string, data any) {
	msg, err := json.Marshal(map[string]any{
		"event": event,
		"data":  data,
	})
	if err != nil {
		return
	}
	s.subsMu.Lock()
	defer s.subsMu.Unlock()
	for ch := range s.subs {
		select {
		case ch <- msg:
		default:
			// Slow consumer, drop message
		}
	}
}
