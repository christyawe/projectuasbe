package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GetTokenExpiration mengambil waktu expired dari JWT
func GetTokenExpiration(tokenString string) (time.Time, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return time.Time{}, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return time.Time{}, errors.New("invalid token claims")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return time.Time{}, errors.New("exp not found in token")
	}

	return time.Unix(int64(exp), 0), nil
}

// GetTokenIssuedAt mengambil waktu issued dari JWT
func GetTokenIssuedAt(tokenString string) (time.Time, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return time.Time{}, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return time.Time{}, errors.New("invalid token claims")
	}

	iat, ok := claims["iat"].(float64)
	if !ok {
		return time.Time{}, errors.New("iat not found in token")
	}

	return time.Unix(int64(iat), 0), nil
}
