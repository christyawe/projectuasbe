package test

import (
	"testing"
	"time"
	"UASBE/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestCheckPasswordHash(t *testing.T) {
	password := "testpassword123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	t.Run("Valid password", func(t *testing.T) {
		result := utils.CheckPasswordHash(password, string(hash))
		assert.True(t, result)
	})

	t.Run("Invalid password", func(t *testing.T) {
		result := utils.CheckPasswordHash("wrongpassword", string(hash))
		assert.False(t, result)
	})

	t.Run("Empty password", func(t *testing.T) {
		result := utils.CheckPasswordHash("", string(hash))
		assert.False(t, result)
	})
}

func TestGenerateJWT(t *testing.T) {
	userID := "123e4567-e89b-12d3-a456-426614174000"
	username := "testuser"
	role := "student"
	permissions := []string{"read", "write"}

	t.Run("Generate valid JWT", func(t *testing.T) {
		token, err := utils.GenerateJWT(userID, username, role, permissions)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Verify token can be parsed
		parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			return utils.JWTSecretKey, nil
		})
		assert.NoError(t, err)
		assert.True(t, parsedToken.Valid)

		// Verify claims
		claims := parsedToken.Claims.(jwt.MapClaims)
		assert.Equal(t, userID, claims["user_id"])
		assert.Equal(t, username, claims["username"])
		assert.Equal(t, role, claims["role"])
	})

	t.Run("Generate JWT with empty permissions", func(t *testing.T) {
		token, err := utils.GenerateJWT(userID, username, role, []string{})
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})
}

func TestExtractToken(t *testing.T) {
	t.Run("Valid Bearer token", func(t *testing.T) {
		authHeader := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
		token := utils.ExtractToken(authHeader)
		assert.Equal(t, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", token)
	})

	t.Run("Invalid format - no Bearer prefix", func(t *testing.T) {
		authHeader := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
		token := utils.ExtractToken(authHeader)
		assert.Empty(t, token)
	})

	t.Run("Empty header", func(t *testing.T) {
		token := utils.ExtractToken("")
		assert.Empty(t, token)
	})
}

func TestValidateToken(t *testing.T) {
	userID := "123e4567-e89b-12d3-a456-426614174000"
	username := "testuser"
	role := "student"
	permissions := []string{"read", "write"}

	t.Run("Valid token", func(t *testing.T) {
		token, _ := utils.GenerateJWT(userID, username, role, permissions)
		claims, err := utils.ValidateToken(token)
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, userID, claims["user_id"])
		assert.Equal(t, username, claims["username"])
	})

	t.Run("Invalid token", func(t *testing.T) {
		claims, err := utils.ValidateToken("invalid.token.here")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("Expired token", func(t *testing.T) {
		// Create expired token
		now := time.Now().Add(-25 * time.Hour)
		claims := jwt.MapClaims{
			"user_id":     userID,
			"username":    username,
			"role":        role,
			"permissions": permissions,
			"iat":         now.Unix(),
			"exp":         now.Add(time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString(utils.JWTSecretKey)

		result, err := utils.ValidateToken(tokenString)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("Empty token", func(t *testing.T) {
		claims, err := utils.ValidateToken("")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

func TestGetTokenExpiration(t *testing.T) {
	userID := "123e4567-e89b-12d3-a456-426614174000"
	username := "testuser"
	role := "student"
	permissions := []string{"read"}

	t.Run("Valid token", func(t *testing.T) {
		token, _ := utils.GenerateJWT(userID, username, role, permissions)
		expiresAt, err := utils.GetTokenExpiration(token)
		assert.NoError(t, err)
		assert.True(t, expiresAt.After(time.Now()))
	})

	t.Run("Invalid token", func(t *testing.T) {
		_, err := utils.GetTokenExpiration("invalid.token")
		assert.Error(t, err)
	})
}

func TestGetTokenIssuedAt(t *testing.T) {
	userID := "123e4567-e89b-12d3-a456-426614174000"
	username := "testuser"
	role := "student"
	permissions := []string{"read"}

	t.Run("Valid token", func(t *testing.T) {
		beforeGenerate := time.Now()
		token, _ := utils.GenerateJWT(userID, username, role, permissions)
		issuedAt, err := utils.GetTokenIssuedAt(token)
		assert.NoError(t, err)
		assert.True(t, issuedAt.After(beforeGenerate.Add(-time.Second)))
		assert.True(t, issuedAt.Before(time.Now().Add(time.Second)))
	})

	t.Run("Invalid token", func(t *testing.T) {
		_, err := utils.GetTokenIssuedAt("invalid.token")
		assert.Error(t, err)
	})
}
