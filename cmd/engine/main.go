package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	enginehttp "github.com/betting-platform/internal/infrastructure/http"
	"github.com/gorilla/mux"
)

func main() {
	log.Println("Starting Betting Engine Service")

	// Initialize router
	r := mux.NewRouter()

	// Initialize handlers
	handler := enginehttp.NewEngineHandler(nil)

	// Register routes
	handler.RegisterRoutes(r)

	// Start odds sync from Sportradar
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go syncOddsFromProvider(ctx)

	// Start server
	port := os.Getenv("ENGINE_PORT")
	if port == "" {
		port = "8082"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Engine Service listening on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Engine Service...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Engine Service exited cleanly")
}

func syncOddsFromProvider(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Println("Starting odds sync from Sportradar...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Odds sync stopped")
			return
		case <-ticker.C:
			// In production:
			// 1. Fetch odds from Sportradar API
			// 2. Update Redis cache
			// 3. Publish odds update event via NATS
			// 4. Broadcast to WebSocket clients
			log.Println("Syncing odds... (mock)")
		}
	}
}
