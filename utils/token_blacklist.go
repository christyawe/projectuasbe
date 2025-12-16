package utils

import (
	"sync"
	"time"
)

// TokenBlacklistManager manages blacklisted tokens in memory
type TokenBlacklistManager struct {
	blacklist map[string]time.Time // token -> expiration time
	mu        sync.RWMutex
}

var (
	// Global instance
	BlacklistManager *TokenBlacklistManager
)

func init() {
	BlacklistManager = NewTokenBlacklistManager()
	// Start cleanup goroutine
	go BlacklistManager.StartCleanup()
}

// NewTokenBlacklistManager creates a new blacklist manager
func NewTokenBlacklistManager() *TokenBlacklistManager {
	return &TokenBlacklistManager{
		blacklist: make(map[string]time.Time),
	}
}

// AddToken adds a token to blacklist
func (m *TokenBlacklistManager) AddToken(token string, expiresAt time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.blacklist[token] = expiresAt
}

// IsBlacklisted checks if token is blacklisted
func (m *TokenBlacklistManager) IsBlacklisted(token string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	expiresAt, exists := m.blacklist[token]
	if !exists {
		return false
	}

	// Check if token is still in blacklist (not expired)
	if time.Now().After(expiresAt) {
		// Token expired, remove from blacklist
		go m.removeToken(token)
		return false
	}

	return true
}

// removeToken removes a token from blacklist (internal use)
func (m *TokenBlacklistManager) removeToken(token string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.blacklist, token)
}

// StartCleanup periodically removes expired tokens
func (m *TokenBlacklistManager) StartCleanup() {
	ticker := time.NewTicker(1 * time.Hour) // Cleanup every hour
	defer ticker.Stop()

	for range ticker.C {
		m.cleanup()
	}
}

// cleanup removes all expired tokens
func (m *TokenBlacklistManager) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for token, expiresAt := range m.blacklist {
		if now.After(expiresAt) {
			delete(m.blacklist, token)
		}
	}
}

// GetBlacklistSize returns the number of blacklisted tokens
func (m *TokenBlacklistManager) GetBlacklistSize() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.blacklist)
}
