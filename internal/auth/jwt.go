package auth

import (
	"time"
	"errors"

	"github.com/Emmanuel326/chatserver/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

// Claims defines the payload structure for our JWT.
// We embed jwt.RegisteredClaims to include standard claims like Expiration time.
type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// JWTManager handles token creation and validation.
type JWTManager struct {
	SecretKey string
	Expiry    time.Duration
}

// NewJWTManager creates a new JWTManager instance using the application config.
func NewJWTManager(cfg *config.Config) *JWTManager {
	return &JWTManager{
		SecretKey: cfg.JWT_SECRET,
		// Convert minutes from config to time.Duration
		Expiry: time.Duration(cfg.JWT_EXPIRY) * time.Minute,
	}
}

// GenerateToken creates a new signed JWT for the given user ID.
func (m *JWTManager) GenerateToken(userID int64) (string, error) {
	// Define token expiry time
	expirationTime := time.Now().Add(m.Expiry)

	// Create the Claims (payload)
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			// Recommended claims
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "chatserver",
		},
	}

	// Create the token using the claims and the HMAC signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(m.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken parses and validates a JWT string.
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method is what we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(m.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
