package auth

import (
	"errors"
	"os"
	"time"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims for a user session
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrExpiredToken  = errors.New("token has expired")
	ErrMissingSecret = errors.New("JWT_SECRET not set in environment")
)

// GetJWTSecret retrieves the JWT secret from environment
func GetJWTSecret() (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", ErrMissingSecret
	}
	return secret, nil
}

// GenerateToken creates a new JWT token for a user
func GenerateToken(user *models.User, expirationHours int) (string, error) {
	secret, err := GetJWTSecret()
	if err != nil {
		return "", err
	}

	if expirationHours == 0 {
		expirationHours = 24 // Default 24 hours
	}

	expirationTime := time.Now().Add(time.Duration(expirationHours) * time.Hour)

	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// VerifyToken validates a JWT token string and returns the claims
func VerifyToken(tokenString string) (*Claims, error) {
	secret, err := GetJWTSecret()
	if err != nil {
		return nil, err
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// Check if token has expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, ErrExpiredToken
	}

	return claims, nil
}
