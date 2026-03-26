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
	"github.com/SE-RoyalFlush/Poker/backend/pkg/migrations"
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

	// Initialize database connection
	dbCfg := db.DefaultConfig()

	_, err := db.Connect(dbCfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}

	// Run migrations
	database, err := db.GetDB()
	if err != nil {
		log.Fatalf("Failed to get database: %v", err)
	}
	if err := migrations.RunMigrations(database); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error during database shutdown: %v", err)
		}
	}()

	// Initialize router
	router := mux.NewRouter()

	// API Routes
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/health", api.HealthHandler).Methods("GET")
	apiRouter.HandleFunc("/register", api.RegisterHandler).Methods("POST")
	apiRouter.HandleFunc("/csrf", api.CSRFTokenHandler).Methods("GET")
	apiRouter.HandleFunc("/me", api.MeHandler).Methods("GET")
	apiRouter.HandleFunc("/admin/users", api.AdminListUsersHandler).Methods("GET")
	apiRouter.HandleFunc("/admin/users/{id:[0-9]+}", api.AdminDeleteUserHandler).Methods("DELETE")

	// Authentication routes
	router.HandleFunc("/api/login", api.LoginHandler).Methods("POST")
	router.HandleFunc("/api/logout", api.LogoutHandler).Methods("POST")
	router.HandleFunc("/api/me", api.MeHandler).Methods("GET")

	router.NotFoundHandler = http.HandlerFunc(api.NotFoundHandler)

	csrfAuthKey := []byte(os.Getenv("CSRF_AUTH_KEY"))
	if len(csrfAuthKey) != 32 {
		log.Println("CSRF_AUTH_KEY must be 32 bytes; using insecure development key")
		csrfAuthKey = []byte("dev-only-32-byte-csrf-secret-key")
	}

	// Configure whether the CSRF cookie should be marked Secure.
	// Default to Secure=true, and only disable it in explicit development environments.
	appEnv := os.Getenv("APP_ENV")
	goEnv := os.Getenv("GO_ENV")
	csrfSecure := true
	if appEnv == "development" || goEnv == "development" {
		csrfSecure = false
	}

	csrfMiddleware := csrf.Protect(
		csrfAuthKey,
		csrf.RequestHeader("X-CSRF-Token"),
		csrf.Path("/"),
		csrf.Secure(csrfSecure),
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

	// Start server with timeouts and graceful shutdown
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

	log.Println("Shutting down server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
