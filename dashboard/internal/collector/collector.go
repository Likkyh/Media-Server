package collector

import (
	"context"
	"log"
	"time"

	"arcticmon/internal/config"
	"arcticmon/internal/store"
)

// Collector is the interface all data collectors implement.
type Collector interface {
	Name() string
	Collect(ctx context.Context) error
}

// Orchestrator manages all collectors and their polling loops.
type Orchestrator struct {
	store *store.Store
	cfg   *config.Config
}

// NewOrchestrator creates a new orchestrator.
func NewOrchestrator(s *store.Store, cfg *config.Config) *Orchestrator {
	return &Orchestrator{store: s, cfg: cfg}
}

// Start launches all collector goroutines. Call cancel on the context to stop.
func (o *Orchestrator) Start(ctx context.Context) {
	fast := 10 * time.Second
	medium := 30 * time.Second
	slow := 60 * time.Second

	// Fast polling (10s)
	o.run(ctx, NewHostCollector(o.cfg, o.store), fast)
	o.run(ctx, NewDockerCollector(o.cfg, o.store), fast)
	o.run(ctx, NewJellyfinSessionCollector(o.cfg, o.store), fast)
	o.run(ctx, NewQbitTransferCollector(o.cfg, o.store), fast)
	o.run(ctx, NewUnmanicCollector(o.cfg, o.store), fast)

	// Medium polling (30s)
	o.run(ctx, NewSeerrCollector(o.cfg, o.store), medium)
	o.run(ctx, NewRadarrCollector(o.cfg, o.store), medium)
	o.run(ctx, NewSonarrCollector(o.cfg, o.store), medium)
	o.run(ctx, NewSabnzbdCollector(o.cfg, o.store), medium)

	// Slow polling (60s)
	o.run(ctx, NewJellyfinLibraryCollector(o.cfg, o.store), slow)
	o.run(ctx, NewProwlarrCollector(o.cfg, o.store), slow)
	o.run(ctx, NewBazarrCollector(o.cfg, o.store), slow)
	piholeLookup := NewPiholeLookup(o.cfg)
	o.run(ctx, NewSSHSecurityCollector(o.cfg, o.store, piholeLookup), slow)
}

// run starts a goroutine that calls the collector on a ticker.
func (o *Orchestrator) run(ctx context.Context, c Collector, interval time.Duration) {
	go func() {
		// Initial collection
		if err := c.Collect(ctx); err != nil {
			log.Printf("[%s] initial collect error: %v", c.Name(), err)
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := c.Collect(ctx); err != nil {
					log.Printf("[%s] collect error: %v", c.Name(), err)
				}
			}
		}
	}()
}
