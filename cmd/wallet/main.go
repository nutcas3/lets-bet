package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	log.Println("Starting Wallet Service")

	// Initialize router
	r := mux.NewRouter()

	// Health check
	r.HandleFunc("/health", healthCheckHandler).Methods("GET")

	// Internal API (called by other services)
	api := r.PathPrefix("/internal/v1").Subrouter()
	api.HandleFunc("/wallets/{userId}", getWalletHandler).Methods("GET")
	api.HandleFunc("/wallets/{userId}/debit", debitWalletHandler).Methods("POST")
	api.HandleFunc("/wallets/{userId}/credit", creditWalletHandler).Methods("POST")
	api.HandleFunc("/transactions", createTransactionHandler).Methods("POST")

	// Start server
	port := os.Getenv("WALLET_PORT")
	if port == "" {
		port = "8081"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Wallet Service listening on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Wallet Service...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Wallet Service exited cleanly")
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"wallet"}`))
}

func getWalletHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"balance":5000.00,"currency":"KES","version":42}`))
}

func debitWalletHandler(w http.ResponseWriter, r *http.Request) {
	// Implement atomic wallet debit with optimistic locking
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success":true,"new_balance":4900.00}`))
}

func creditWalletHandler(w http.ResponseWriter, r *http.Request) {
	// Implement atomic wallet credit
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success":true,"new_balance":5100.00}`))
}

func createTransactionHandler(w http.ResponseWriter, r *http.Request) {
	// Log transaction to immutable ledger
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"transaction_id":"txn_123","status":"completed"}`))
}
