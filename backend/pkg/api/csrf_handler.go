package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/csrf"
)

type CSRFResponse struct {
	CSRFToken string `json:"csrfToken"`
}

func CSRFTokenHandler(w http.ResponseWriter, r *http.Request) {
	token := csrf.Token(r)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-CSRF-Token", token)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(CSRFResponse{CSRFToken: token}); err != nil {
		log.Printf("Failed to encode CSRF response: %v", err)
	}
}
