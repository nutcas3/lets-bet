package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/betting-platform/internal/settlement"
	"github.com/gorilla/mux"
)

func main() {
	log.Println("Starting Settlement Service")

	// Initialize router
	r := mux.NewRouter()

	// Initialize handlers
	handler := settlement.NewHandler()

	// Register routes
	handler.RegisterRoutes(r)

	// Start settlement processor in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go runSettlementProcessor(ctx)

	// Start server
	port := os.Getenv("SETTLEMENT_PORT")
	if port == "" {
		port = "8083"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Settlement Service listening on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Settlement Service...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	cancel()

	log.Println("Settlement Service exited cleanly")
}

func runSettlementProcessor(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Process pending settlements
			log.Println("Checking for bets to settle...")

			// In production:
			// 1. Query pending bets
			// 2. Check match results from odds provider
			// 3. Calculate winnings
			// 4. Update bet status
			// 5. Credit winners' wallets
			// 6. Deduct taxes
			// 7. Publish settlement event
		}
	}
}
