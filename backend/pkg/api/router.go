package api

import (
	"net/http"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/middleware"
	"github.com/gorilla/mux"
)

// NewRouter builds the application's route tree with separate public and protected branches.
func NewRouter() *mux.Router {
	router := mux.NewRouter()

	publicAPI := router.PathPrefix("/api").Subrouter()
	publicAPI.HandleFunc("/health", HealthHandler).Methods(http.MethodGet)
	publicAPI.HandleFunc("/register", RegisterHandler).Methods(http.MethodPost)
	publicAPI.HandleFunc("/login", LoginHandler).Methods(http.MethodPost)
	publicAPI.HandleFunc("/logout", LogoutHandler).Methods(http.MethodPost)
	publicAPI.HandleFunc("/csrf", CSRFTokenHandler).Methods(http.MethodGet)
	publicAPI.HandleFunc("/admin/users", AdminListUsersHandler).Methods(http.MethodGet)
	publicAPI.HandleFunc("/admin/users/{id:[0-9]+}", AdminDeleteUserHandler).Methods(http.MethodDelete)

	protected := router.NewRoute().Subrouter()
	protected.Use(middleware.AuthMiddleware)

	protectedAPI := protected.PathPrefix("/api").Subrouter()
	protectedAPI.HandleFunc("/me", MeHandler).Methods(http.MethodGet)
	protectedAPI.HandleFunc("/rooms", RoomsHandler).Methods(http.MethodGet, http.MethodPost)
	protectedAPI.HandleFunc("/rooms/join", JoinRoomHandler).Methods(http.MethodPost)
	protectedAPI.HandleFunc("/rooms/{code}", RoomHandler).Methods(http.MethodGet)

	protected.HandleFunc("/ws", WebSocketHandler)

	router.NotFoundHandler = http.HandlerFunc(NotFoundHandler)

	return router
}
