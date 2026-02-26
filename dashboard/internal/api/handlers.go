package api

import (
	"encoding/json"
	"net/http"
	"time"

	"arcticmon/internal/store"
)

// Handlers provides HTTP handlers for the JSON API.
type Handlers struct {
	store *store.Store
}

func (h *Handlers) respondJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	json.NewEncoder(w).Encode(data)
}

func (h *Handlers) Overview(w http.ResponseWriter, r *http.Request) {
	data := h.store.Get()
	data.UpdatedAt = time.Now()
	h.respondJSON(w, data)
}

func (h *Handlers) Services(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, h.store.Get().Services)
}

func (h *Handlers) Streams(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, h.store.Get().Streams)
}

func (h *Handlers) Torrents(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, h.store.Get().Torrents)
}

func (h *Handlers) Downloads(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, h.store.Get().Downloads)
}

func (h *Handlers) Requests(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, h.store.Get().Requests)
}

func (h *Handlers) Host(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, h.store.Get().Host)
}

func (h *Handlers) Transcoding(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, h.store.Get().Transcodes)
}

func (h *Handlers) Library(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, h.store.Get().Library)
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, h.store.Get().Health)
}

func (h *Handlers) SSHSecurity(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, h.store.Get().SSHSecurity)
}
