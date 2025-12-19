package test

import (
	"testing"
	"time"
	"UASBE/utils"

	"github.com/stretchr/testify/assert"
)

func TestNewTokenBlacklistManager(t *testing.T) {
	manager := utils.NewTokenBlacklistManager()
	assert.NotNil(t, manager)
	assert.Equal(t, 0, manager.GetBlacklistSize())
}

func TestAddToken(t *testing.T) {
	manager := utils.NewTokenBlacklistManager()
	token := "test.token.here"
	expiresAt := time.Now().Add(time.Hour)

	t.Run("Add single token", func(t *testing.T) {
		manager.AddToken(token, expiresAt)
		assert.Equal(t, 1, manager.GetBlacklistSize())
		assert.True(t, manager.IsBlacklisted(token))
	})

	t.Run("Add multiple tokens", func(t *testing.T) {
		manager2 := utils.NewTokenBlacklistManager()
		token1 := "token1"
		token2 := "token2"
		expiresAt := time.Now().Add(time.Hour)

		manager2.AddToken(token1, expiresAt)
		manager2.AddToken(token2, expiresAt)

		assert.Equal(t, 2, manager2.GetBlacklistSize())
		assert.True(t, manager2.IsBlacklisted(token1))
		assert.True(t, manager2.IsBlacklisted(token2))
	})
}

func TestIsBlacklisted(t *testing.T) {
	manager := utils.NewTokenBlacklistManager()

	t.Run("Token not in blacklist", func(t *testing.T) {
		result := manager.IsBlacklisted("nonexistent.token")
		assert.False(t, result)
	})

	t.Run("Token in blacklist", func(t *testing.T) {
		token := "blacklisted.token"
		expiresAt := time.Now().Add(time.Hour)
		manager.AddToken(token, expiresAt)
		assert.True(t, manager.IsBlacklisted(token))
	})

	t.Run("Expired token in blacklist", func(t *testing.T) {
		manager2 := utils.NewTokenBlacklistManager()
		token := "expired.token"
		expiresAt := time.Now().Add(-time.Hour) // Already expired
		manager2.AddToken(token, expiresAt)

		// Should return false because token is expired
		result := manager2.IsBlacklisted(token)
		assert.False(t, result)
	})
}

func TestGetBlacklistSize(t *testing.T) {
	manager := utils.NewTokenBlacklistManager()

	t.Run("Empty blacklist", func(t *testing.T) {
		assert.Equal(t, 0, manager.GetBlacklistSize())
	})

	t.Run("After adding tokens", func(t *testing.T) {
		expiresAt := time.Now().Add(time.Hour)
		manager.AddToken("token1", expiresAt)
		manager.AddToken("token2", expiresAt)
		manager.AddToken("token3", expiresAt)
		assert.Equal(t, 3, manager.GetBlacklistSize())
	})
}

func TestTokenBlacklistConcurrency(t *testing.T) {
	manager := utils.NewTokenBlacklistManager()
	expiresAt := time.Now().Add(time.Hour)

	// Test concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(index int) {
			token := "token" + string(rune(index))
			manager.AddToken(token, expiresAt)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have 10 tokens
	assert.Equal(t, 10, manager.GetBlacklistSize())
}
