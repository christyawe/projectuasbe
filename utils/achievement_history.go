package utils

import (
	"sync"
	model "UASBE/app/model/Postgresql"

	"github.com/google/uuid"
)

// AchievementHistoryManager manages achievement history in memory
type AchievementHistoryManager struct {
	history map[uuid.UUID][]model.AchievementStatusLog // achievement_id -> history entries
	mu      sync.RWMutex
}

var (
	// Global instance
	HistoryManager *AchievementHistoryManager
)

func init() {
	HistoryManager = NewAchievementHistoryManager()
}

// NewAchievementHistoryManager creates a new history manager
func NewAchievementHistoryManager() *AchievementHistoryManager {
	return &AchievementHistoryManager{
		history: make(map[uuid.UUID][]model.AchievementStatusLog),
	}
}

// AddEntry adds a history entry
func (m *AchievementHistoryManager) AddEntry(entry model.AchievementStatusLog) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.history[entry.AchievementID]; !exists {
		m.history[entry.AchievementID] = []model.AchievementStatusLog{}
	}

	m.history[entry.AchievementID] = append(m.history[entry.AchievementID], entry)
}

// GetHistory gets history for an achievement
func (m *AchievementHistoryManager) GetHistory(achievementID uuid.UUID) []model.AchievementStatusLog {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entries, exists := m.history[achievementID]
	if !exists {
		return []model.AchievementStatusLog{}
	}

	// Return copy to prevent external modification
	result := make([]model.AchievementStatusLog, len(entries))
	copy(result, entries)
	return result
}

// ClearHistory clears history for an achievement (optional, for cleanup)
func (m *AchievementHistoryManager) ClearHistory(achievementID uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.history, achievementID)
}

// GetTotalEntries returns total number of history entries
func (m *AchievementHistoryManager) GetTotalEntries() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total := 0
	for _, entries := range m.history {
		total += len(entries)
	}
	return total
}
