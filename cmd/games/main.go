package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/betting-platform/internal/core/usecase"
	"github.com/betting-platform/internal/games"
	"github.com/betting-platform/internal/repository/mock"
	"github.com/gorilla/mux"
)

func main() {
	log.Println("Starting Crash Games Service")

	// Initialize router
	r := mux.NewRouter()

	// Initialize handlers
	handler := games.NewHandler()

	// Register routes
	handler.RegisterRoutes(r)

	// Initialize context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize WebSocket hub
	hub := games.NewHub()
	go hub.Run()
	log.Println("WebSocket Hub started")

	// Initialize Provably Fair service
	fairService := usecase.NewProvablyFairService()

	// Initialize repositories (mock for now)
	// In production, connect to PostgreSQL
	gameRepo := mock.NewMockGameRepository()
	betRepo := mock.NewMockGameBetRepository()

	// Initialize Crash Game Engine
	engine := games.NewCrashGameEngine(hub, fairService, gameRepo, betRepo)

	// Start the game loop
	go engine.Start(ctx)

	log.Println("Crash Game Engine running")
	log.Println("Players can connect via WebSocket to receive real-time updates")

	// Start server
	port := os.Getenv("GAMES_PORT")
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
		log.Printf("Games Service listening on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Games Service...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	cancel()

	log.Println("Games Service exited cleanly")
}
