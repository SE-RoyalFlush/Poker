package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/api"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
	"github.com/gorilla/mux"
)

func main() {
	// Initialize database connection
	dbCfg := db.DefaultConfig()

	_, err := db.Connect(dbCfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error during database shutdown: %v", err)
		}
	}()

	// Initialize router
	router := mux.NewRouter()

	router.HandleFunc("/health", api.HealthHandler).Methods("GET")

	router.NotFoundHandler = http.HandlerFunc(api.NotFoundHandler)

	// Start server with timeouts and graceful shutdown
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Println("Starting server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
