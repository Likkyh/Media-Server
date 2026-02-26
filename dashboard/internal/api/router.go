package api

import (
	"crypto/subtle"
	"embed"
	"io/fs"
	"net/http"
	"sync"
	"time"

	"arcticmon/internal/config"
	"arcticmon/internal/store"
)

// NewRouter creates the HTTP mux with all routes registered.
func NewRouter(s *store.Store, cfg *config.Config, webFS embed.FS) http.Handler {
	mux := http.NewServeMux()
	h := &Handlers{store: s}

	// Unauthenticated health endpoint for Docker healthcheck
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// API routes
	mux.HandleFunc("GET /api/overview", h.Overview)
	mux.HandleFunc("GET /api/services", h.Services)
	mux.HandleFunc("GET /api/streams", h.Streams)
	mux.HandleFunc("GET /api/torrents", h.Torrents)
	mux.HandleFunc("GET /api/downloads", h.Downloads)
	mux.HandleFunc("GET /api/requests", h.Requests)
	mux.HandleFunc("GET /api/host", h.Host)
	mux.HandleFunc("GET /api/transcoding", h.Transcoding)
	mux.HandleFunc("GET /api/library", h.Library)
	mux.HandleFunc("GET /api/health", h.Health)
	mux.HandleFunc("GET /api/ssh-security", h.SSHSecurity)

	// Actions (rate-limited)
	rl := newRateLimiter(30 * time.Second)
	actions := NewActions(cfg)
	mux.HandleFunc("POST /api/actions/restart-stack", rl.wrap(actions.RestartStack))
	mux.HandleFunc("POST /api/actions/restart-vm", rl.wrap(actions.RestartVM))
	mux.HandleFunc("POST /api/actions/update-stack", rl.wrap(actions.UpdateStack))
	mux.HandleFunc("POST /api/actions/update-system", rl.wrap(actions.UpdateSystem))

	// SSE
	sse := &SSEHandler{store: s}
	mux.HandleFunc("GET /api/events", sse.ServeHTTP)

	// Static files (embedded)
	webSub, _ := fs.Sub(webFS, "web")
	fileServer := http.FileServer(http.FS(webSub))
	mux.Handle("/", fileServer)

	// Chain middleware: security headers → basic auth (skips /healthz)
	var handler http.Handler = mux
	handler = basicAuthMiddleware(handler, cfg)
	handler = securityHeadersMiddleware(handler)

	return handler
}

// securityHeadersMiddleware adds security headers to all responses.
func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'")
		next.ServeHTTP(w, r)
	})
}

// basicAuthMiddleware protects all endpoints except /healthz.
func basicAuthMiddleware(next http.Handler, cfg *config.Config) http.Handler {
	user := cfg.DashboardUser
	pass := cfg.DashboardPass
	if user == "" || pass == "" {
		// No credentials configured, skip auth
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for healthcheck
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}

		u, p, ok := r.BasicAuth()
		if !ok ||
			subtle.ConstantTimeCompare([]byte(u), []byte(user)) != 1 ||
			subtle.ConstantTimeCompare([]byte(p), []byte(pass)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="Arctic Monitor"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// rateLimiter tracks last call time per endpoint.
type rateLimiter struct {
	mu       sync.Mutex
	cooldown time.Duration
	lastCall map[string]time.Time
}

func newRateLimiter(cooldown time.Duration) *rateLimiter {
	return &rateLimiter{
		cooldown: cooldown,
		lastCall: make(map[string]time.Time),
	}
}

func (rl *rateLimiter) wrap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rl.mu.Lock()
		last, exists := rl.lastCall[r.URL.Path]
		if exists && time.Since(last) < rl.cooldown {
			rl.mu.Unlock()
			http.Error(w, `{"error":"rate limited, try again later"}`, http.StatusTooManyRequests)
			return
		}
		rl.lastCall[r.URL.Path] = time.Now()
		rl.mu.Unlock()
		next(w, r)
	}
}
