package auth

import (
	"errors"
	"net/http"
	"os"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"github.com/gorilla/securecookie"
	"gorm.io/gorm"
)

const (
	// SessionCookieName stores the authenticated username in a signed cookie.
	SessionCookieName = "session-id"
)

var (
	ErrMissingSession    = errors.New("missing session cookie")
	ErrInvalidSession    = errors.New("invalid session cookie")
	ErrUnauthenticated   = errors.New("unauthenticated request")
	ErrSessionUserLookup = errors.New("failed to resolve session user")

	sessionKey   = []byte(os.Getenv("SESSION_KEY"))
	sessionCodec *securecookie.SecureCookie
)

func init() {
	if len(sessionKey) == 0 {
		sessionKey = []byte("dev-only-32-byte-session-secret-")
	}
	sessionCodec = securecookie.New(sessionKey, nil)
}

func EncodeSessionValue(username string) (string, error) {
	return sessionCodec.Encode(SessionCookieName, username)
}

func DecodeSessionUsername(r *http.Request) (string, error) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return "", ErrMissingSession
		}
		return "", ErrInvalidSession
	}

	var username string
	if err := sessionCodec.Decode(SessionCookieName, cookie.Value, &username); err != nil {
		return "", ErrInvalidSession
	}

	if username == "" {
		return "", ErrUnauthenticated
	}

	return username, nil
}

func AuthenticatedUserFromRequest(r *http.Request) (*models.User, error) {
	username, err := DecodeSessionUsername(r)
	if err != nil {
		return nil, err
	}

	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := database.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUnauthenticated
		}
		return nil, ErrSessionUserLookup
	}

	return &user, nil
}

func SetSessionCookie(w http.ResponseWriter, username string) error {
	encoded, err := EncodeSessionValue(username)
	if err != nil {
		return err
	}

	appEnv := os.Getenv("APP_ENV")
	goEnv := os.Getenv("GO_ENV")
	secure := true
	if appEnv == "development" || goEnv == "development" {
		secure = false
	}

	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
}
