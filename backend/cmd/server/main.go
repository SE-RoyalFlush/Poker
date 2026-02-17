package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/api"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	envLoaded := false
	for _, envPath := range []string{".env", "../.env"} {
		if err := godotenv.Load(envPath); err == nil {
			envLoaded = true
			break
		}
	}
	if !envLoaded {
		log.Println("No .env file found; using system environment variables")
	}

	router := mux.NewRouter()

	// Register routes
	router.HandleFunc("/health", api.HealthHandler).Methods("GET")
	router.HandleFunc("/api/csrf", api.CSRFTokenHandler).Methods("GET")

	router.NotFoundHandler = http.HandlerFunc(api.NotFoundHandler)

	csrfAuthKey := []byte(os.Getenv("CSRF_AUTH_KEY"))
	if len(csrfAuthKey) != 32 {
		log.Println("CSRF_AUTH_KEY must be 32 bytes; using insecure development key")
		csrfAuthKey = []byte("dev-only-32-byte-csrf-secret-key")
	}

	csrfMiddleware := csrf.Protect(
		csrfAuthKey,
		csrf.RequestHeader("X-CSRF-Token"),
		csrf.Path("/"),
		csrf.Secure(false),
		csrf.HttpOnly(true),
		csrf.SameSite(csrf.SameSiteLaxMode),
	)

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:4200"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"X-CSRF-Token"},
		AllowCredentials: true,
	})

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      corsMiddleware.Handler(csrfMiddleware(router)),
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

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
