package collector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/netip"
	"sync"
	"time"

	"arcticmon/internal/config"
)

// PiholeLookup resolves local IPs to hostnames using Pi-hole v6 DHCP leases.
type PiholeLookup struct {
	cfg *config.Config

	mu    sync.RWMutex
	cache map[string]string // IP → hostname
	lastRefresh time.Time

	sid string // Pi-hole session ID
}

// NewPiholeLookup creates a new Pi-hole DHCP lookup instance.
func NewPiholeLookup(cfg *config.Config) *PiholeLookup {
	return &PiholeLookup{
		cfg:   cfg,
		cache: make(map[string]string),
	}
}

// Lookup returns the hostname for a private IP, or empty string.
func (p *PiholeLookup) Lookup(ip string) string {
	if p.cfg.PiholePassword == "" {
		return ""
	}

	if !isPrivateIP(ip) {
		return ""
	}

	p.maybeRefresh()

	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.cache[ip]
}

func isPrivateIP(s string) bool {
	addr, err := netip.ParseAddr(s)
	if err != nil {
		return false
	}
	return addr.IsPrivate()
}

func (p *PiholeLookup) maybeRefresh() {
	p.mu.RLock()
	needsRefresh := time.Since(p.lastRefresh) > 5*time.Minute
	p.mu.RUnlock()

	if !needsRefresh {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock
	if time.Since(p.lastRefresh) <= 5*time.Minute {
		return
	}

	leases, err := p.fetchLeases()
	if err != nil {
		log.Printf("[pihole] failed to fetch DHCP leases: %v", err)
		// Set lastRefresh so we don't retry on every Lookup() call
		p.lastRefresh = time.Now()
		return
	}

	newCache := make(map[string]string, len(leases))
	for _, l := range leases {
		if l.IP != "" && l.Name != "" {
			newCache[l.IP] = l.Name
		}
	}
	p.cache = newCache
	p.lastRefresh = time.Now()
	log.Printf("[pihole] refreshed DHCP cache: %d leases", len(newCache))
}

type piholeAuthResponse struct {
	Session struct {
		SID string `json:"sid"`
	} `json:"session"`
}

type piholeLease struct {
	IP   string `json:"ip"`
	Name string `json:"name"`
}

type piholeLeaseResponse struct {
	Leases []piholeLease `json:"leases"`
}

type piholeErrorResponse struct {
	Error struct {
		Key     string `json:"key"`
		Message string `json:"message"`
	} `json:"error"`
}

func (p *PiholeLookup) authenticate() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body, _ := json.Marshal(map[string]string{"password": p.cfg.PiholePassword})
	req, err := http.NewRequestWithContext(ctx, "POST", p.cfg.PiholeURL+"/api/auth", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth returned status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Check for error response (Pi-hole v6 returns 200 with error JSON)
	var errResp piholeErrorResponse
	if json.Unmarshal(data, &errResp) == nil && errResp.Error.Key != "" {
		return fmt.Errorf("%s: %s", errResp.Error.Key, errResp.Error.Message)
	}

	var authResp piholeAuthResponse
	if err := json.Unmarshal(data, &authResp); err != nil {
		return err
	}

	p.sid = authResp.Session.SID
	if p.sid == "" {
		return fmt.Errorf("empty SID in auth response")
	}
	return nil
}

func (p *PiholeLookup) fetchLeases() ([]piholeLease, error) {
	// Try with existing SID first, then re-auth on 401
	if p.sid != "" {
		leases, err := p.doFetchLeases()
		if err == nil {
			return leases, nil
		}
	}

	// Authenticate and retry
	if err := p.authenticate(); err != nil {
		return nil, fmt.Errorf("auth: %w", err)
	}
	return p.doFetchLeases()
}

func (p *PiholeLookup) doFetchLeases() ([]piholeLease, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", p.cfg.PiholeURL+"/api/dhcp/leases", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Token "+p.sid)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		p.sid = ""
		return nil, fmt.Errorf("unauthorized")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("leases returned status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var leaseResp piholeLeaseResponse
	if err := json.Unmarshal(data, &leaseResp); err != nil {
		return nil, err
	}
	return leaseResp.Leases, nil
}
