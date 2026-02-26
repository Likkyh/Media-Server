package collector

import (
	"bufio"
	"context"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"arcticmon/internal/config"
	"arcticmon/internal/models"
	"arcticmon/internal/store"
)

var (
	// Failed password for invalid user admin from 1.2.3.4 port 22 ssh2
	// Failed password for root from 1.2.3.4 port 22 ssh2
	reFailedPass = regexp.MustCompile(`^(\w+\s+\d+\s+[\d:]+)\s+\S+\s+sshd\[\d+\]:\s+Failed password for (?:invalid user )?(\S+) from (\S+)`)
	// Invalid user admin from 1.2.3.4 port 22
	reInvalidUser = regexp.MustCompile(`^(\w+\s+\d+\s+[\d:]+)\s+\S+\s+sshd\[\d+\]:\s+Invalid user (\S+) from (\S+)`)
	// Connection closed by authenticating user root 1.2.3.4 port 22 [preauth]
	reConnClosed = regexp.MustCompile(`^(\w+\s+\d+\s+[\d:]+)\s+\S+\s+sshd\[\d+\]:\s+Connection closed by authenticating user (\S+) (\S+)`)
	// Accepted publickey for admin from 1.2.3.4 port 22 ssh2
	// Accepted password for admin from 1.2.3.4 port 22 ssh2
	reAccepted = regexp.MustCompile(`^(\w+\s+\d+\s+[\d:]+)\s+\S+\s+sshd\[\d+\]:\s+Accepted (\S+) for (\S+) from (\S+)`)
)

// SSHSecurityCollector parses auth.log for SSH authentication events.
type SSHSecurityCollector struct {
	cfg   *config.Config
	store *store.Store
}

func NewSSHSecurityCollector(cfg *config.Config, s *store.Store) *SSHSecurityCollector {
	return &SSHSecurityCollector{cfg: cfg, store: s}
}

func (c *SSHSecurityCollector) Name() string { return "ssh-security" }

type sshEvent struct {
	t       time.Time
	user    string
	ip      string
	method  string
	success bool
}

func (c *SSHSecurityCollector) Collect(ctx context.Context) error {
	logPaths := []string{"/host/log/auth.log", "/host/log/auth.log.1", "/host/log/secure", "/host/log/secure.1"}

	now := time.Now()
	cutoff30d := now.Add(-30 * 24 * time.Hour)
	cutoff7d := now.Add(-7 * 24 * time.Hour)
	cutoff24h := now.Add(-24 * time.Hour)
	year := now.Year()

	var allFailed []sshEvent
	var allAccepted []sshEvent

	for _, logPath := range logPaths {
		f, err := os.Open(logPath)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 0, 256*1024), 1024*1024)

		for scanner.Scan() {
			line := scanner.Text()
			if !strings.Contains(line, "sshd[") {
				continue
			}

			// Try accepted logins
			if m := reAccepted.FindStringSubmatch(line); m != nil {
				t := parseAuthLogTime(m[1], year, now)
				if t.Before(cutoff30d) {
					continue
				}
				allAccepted = append(allAccepted, sshEvent{
					t: t, user: m[3], ip: m[4], method: m[2], success: true,
				})
				continue
			}

			// Try failed password
			if m := reFailedPass.FindStringSubmatch(line); m != nil {
				t := parseAuthLogTime(m[1], year, now)
				if t.Before(cutoff30d) {
					continue
				}
				allFailed = append(allFailed, sshEvent{
					t: t, user: m[2], ip: m[3], method: "password", success: false,
				})
				continue
			}

			// Try invalid user
			if m := reInvalidUser.FindStringSubmatch(line); m != nil {
				t := parseAuthLogTime(m[1], year, now)
				if t.Before(cutoff30d) {
					continue
				}
				allFailed = append(allFailed, sshEvent{
					t: t, user: m[2], ip: m[3], method: "invalid-user", success: false,
				})
				continue
			}

			// Connection closed during auth
			if m := reConnClosed.FindStringSubmatch(line); m != nil {
				t := parseAuthLogTime(m[1], year, now)
				if t.Before(cutoff30d) {
					continue
				}
				allFailed = append(allFailed, sshEvent{
					t: t, user: m[2], ip: m[3], method: "preauth-closed", success: false,
				})
				continue
			}
		}
		f.Close()
	}

	// Count per period
	var failed24h, failed7d, failed30d int
	var accepted24h, accepted7d, accepted30d int
	failedByIP := map[string]*ipTracker{}

	for _, e := range allFailed {
		failed30d++
		if !e.t.Before(cutoff7d) {
			failed7d++
		}
		if !e.t.Before(cutoff24h) {
			failed24h++
		}
		trackIP(failedByIP, e.ip, e.t)
	}

	for _, e := range allAccepted {
		accepted30d++
		if !e.t.Before(cutoff7d) {
			accepted7d++
		}
		if !e.t.Before(cutoff24h) {
			accepted24h++
		}
	}

	// Build top offenders sorted by attempt count (30d window)
	offenders := make([]models.SSHOffender, 0, len(failedByIP))
	for ip, tr := range failedByIP {
		offenders = append(offenders, models.SSHOffender{
			IP:       ip,
			Attempts: tr.count,
			LastSeen: tr.lastSeen.Format(time.RFC3339),
		})
	}
	sort.Slice(offenders, func(i, j int) bool {
		return offenders[i].Attempts > offenders[j].Attempts
	})
	if len(offenders) > 15 {
		offenders = offenders[:15]
	}

	// Convert to API events (most recent first, keep last 20)
	recentFailed := toAuthEvents(allFailed, 20)
	recentAccepted := toAuthEvents(allAccepted, 20)

	data := models.SSHSecurityData{
		Failed24h:      failed24h,
		Failed7d:       failed7d,
		Failed30d:      failed30d,
		Accepted24h:    accepted24h,
		Accepted7d:     accepted7d,
		Accepted30d:    accepted30d,
		TopOffenders:   offenders,
		RecentFailed:   recentFailed,
		RecentAccepted: recentAccepted,
	}

	c.store.UpdateSSHSecurity(data)
	return nil
}

func toAuthEvents(events []sshEvent, limit int) []models.SSHAuthEvent {
	// Sort by time descending (most recent first)
	sort.Slice(events, func(i, j int) bool {
		return events[i].t.After(events[j].t)
	})
	if len(events) > limit {
		events = events[:limit]
	}
	result := make([]models.SSHAuthEvent, len(events))
	for i, e := range events {
		result[i] = models.SSHAuthEvent{
			Time:    e.t.Format(time.RFC3339),
			User:    e.user,
			IP:      e.ip,
			Method:  e.method,
			Success: e.success,
		}
	}
	return result
}

type ipTracker struct {
	count    int
	lastSeen time.Time
}

func trackIP(m map[string]*ipTracker, ip string, t time.Time) {
	tr, ok := m[ip]
	if !ok {
		tr = &ipTracker{}
		m[ip] = tr
	}
	tr.count++
	if t.After(tr.lastSeen) {
		tr.lastSeen = t
	}
}

// parseAuthLogTime parses syslog-style timestamp "Jan  2 15:04:05".
func parseAuthLogTime(s string, year int, now time.Time) time.Time {
	// Normalize double spaces
	s = strings.Join(strings.Fields(s), " ")
	t, err := time.Parse("Jan 2 15:04:05", s)
	if err != nil {
		return time.Time{}
	}
	t = t.AddDate(year, 0, 0)
	// If parsed time is in the future (e.g., Dec logs read in Jan), subtract a year
	if t.After(now.Add(24 * time.Hour)) {
		t = t.AddDate(-1, 0, 0)
	}
	return t
}
