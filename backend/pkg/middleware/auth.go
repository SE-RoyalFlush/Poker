package middleware

import (
	"net/http"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/auth"
)

const (
	// CookieName is the name of the JWT cookie
	CookieName = "session"
)

// AuthMiddleware validates JWT token from HttpOnly cookie
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the JWT token from HttpOnly cookie
		cookie, err := r.Cookie(CookieName)
		if err != nil {
			if err == http.ErrNoCookie {
				http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
				return
			}
			http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
			return
		}

		// Verify the token
		_, err = auth.VerifyToken(cookie.Value)
		if err != nil {
			http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
			return
		}

		// Token is valid, proceed to next handler
		next.ServeHTTP(w, r)
	})
}
