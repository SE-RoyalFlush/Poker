package auth

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

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

	sessionKey   []byte
	sessionCodec *securecookie.SecureCookie
)

func init() {
	var err error
	sessionKey, err = loadSessionKey()
	if err != nil {
		panic(err)
	}
	sessionCodec = securecookie.New(sessionKey, nil)
}

func loadSessionKey() ([]byte, error) {
	configuredKey := os.Getenv("SESSION_KEY")
	if configuredKey == "" {
		if allowsInsecureDevSessionKey() {
			return []byte("dev-only-32-byte-session-secret-"), nil
		}
		return nil, errors.New("SESSION_KEY must be set outside development/test")
	}

	key := []byte(configuredKey)
	if err := validateSessionKey(key); err != nil {
		return nil, err
	}

	return key, nil
}

func validateSessionKey(key []byte) error {
	if length := len(key); length != 32 && length != 64 {
		return fmt.Errorf("SESSION_KEY must be 32 or 64 bytes, got %d", length)
	}
	return nil
}

func allowsInsecureDevSessionKey() bool {
	appEnv := strings.ToLower(os.Getenv("APP_ENV"))
	goEnv := strings.ToLower(os.Getenv("GO_ENV"))

	return appEnv == "development" || appEnv == "test" || goEnv == "development" || goEnv == "test" || strings.HasSuffix(os.Args[0], ".test")
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
