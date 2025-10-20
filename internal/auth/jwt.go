package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Use a secret key for signing (in production use env variable)
var jwtSecret = []byte("supersecretkey_change_me")

// GenerateToken creates a JWT for a given user ID
func GenerateToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // expires in 24h
	})

	return token.SignedString(jwtSecret)
}

// ValidateToken verifies JWT and returns the userID if valid
func ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Check signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecret, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, _ := claims["user_id"].(string)
		return userID, nil
	}

	return "", jwt.ErrSignatureInvalid
}

