package main

import (
	"context"
	"embed"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"arcticmon/internal/api"
	"arcticmon/internal/collector"
	"arcticmon/internal/config"
	"arcticmon/internal/store"
)

//go:embed web/*
var webFS embed.FS

func main() {
	cfg := config.Load()
	st := store.New()

	// Start collectors
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	orch := collector.NewOrchestrator(st, cfg)
	orch.Start(ctx)

	// HTTP server
	router := api.NewRouter(st, cfg, webFS)
	srv := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 0, // SSE needs no write timeout
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Arctic Monitor listening on %s", cfg.ListenAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	srv.Shutdown(shutdownCtx)
}
