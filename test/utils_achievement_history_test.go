package test

import (
	model "UASBE/app/model/Postgresql"
	"UASBE/utils"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewAchievementHistoryManager(t *testing.T) {
	manager := utils.NewAchievementHistoryManager()
	assert.NotNil(t, manager)
	assert.Equal(t, 0, manager.GetTotalEntries())
}

func TestAddEntry(t *testing.T) {
	manager := utils.NewAchievementHistoryManager()
	achievementID := uuid.New()
	changedBy := uuid.New()

	t.Run("Add single entry", func(t *testing.T) {
		entry := model.AchievementStatusLog{
			ID:            uuid.New(),
			AchievementID: achievementID,
			Status:        "draft",
			ChangedBy:     &changedBy,
			CreatedAt:     time.Now(),
		}

		manager.AddEntry(entry)
		assert.Equal(t, 1, manager.GetTotalEntries())

		history := manager.GetHistory(achievementID)
		assert.Len(t, history, 1)
		assert.Equal(t, "draft", history[0].Status)
	})

	t.Run("Add multiple entries for same achievement", func(t *testing.T) {
		manager2 := utils.NewAchievementHistoryManager()
		achievementID := uuid.New()

		entry1 := model.AchievementStatusLog{
			ID:            uuid.New(),
			AchievementID: achievementID,
			Status:        "draft",
			CreatedAt:     time.Now(),
		}
		entry2 := model.AchievementStatusLog{
			ID:            uuid.New(),
			AchievementID: achievementID,
			Status:        "submitted",
			CreatedAt:     time.Now(),
		}

		manager2.AddEntry(entry1)
		manager2.AddEntry(entry2)

		history := manager2.GetHistory(achievementID)
		assert.Len(t, history, 2)
		assert.Equal(t, 2, manager2.GetTotalEntries())
	})

	t.Run("Add entries for different achievements", func(t *testing.T) {
		manager3 := utils.NewAchievementHistoryManager()
		achievementID1 := uuid.New()
		achievementID2 := uuid.New()

		entry1 := model.AchievementStatusLog{
			ID:            uuid.New(),
			AchievementID: achievementID1,
			Status:        "draft",
			CreatedAt:     time.Now(),
		}
		entry2 := model.AchievementStatusLog{
			ID:            uuid.New(),
			AchievementID: achievementID2,
			Status:        "submitted",
			CreatedAt:     time.Now(),
		}

		manager3.AddEntry(entry1)
		manager3.AddEntry(entry2)

		assert.Equal(t, 2, manager3.GetTotalEntries())
		assert.Len(t, manager3.GetHistory(achievementID1), 1)
		assert.Len(t, manager3.GetHistory(achievementID2), 1)
	})
}

func TestGetHistory(t *testing.T) {
	manager := utils.NewAchievementHistoryManager()

	t.Run("Get history for non-existent achievement", func(t *testing.T) {
		history := manager.GetHistory(uuid.New())
		assert.NotNil(t, history)
		assert.Len(t, history, 0)
	})

	t.Run("Get history for existing achievement", func(t *testing.T) {
		achievementID := uuid.New()
		entry := model.AchievementStatusLog{
			ID:            uuid.New(),
			AchievementID: achievementID,
			Status:        "verified",
			CreatedAt:     time.Now(),
		}

		manager.AddEntry(entry)
		history := manager.GetHistory(achievementID)
		assert.Len(t, history, 1)
		assert.Equal(t, "verified", history[0].Status)
	})

	t.Run("History returns copy not reference", func(t *testing.T) {
		manager2 := utils.NewAchievementHistoryManager()
		achievementID := uuid.New()
		entry := model.AchievementStatusLog{
			ID:            uuid.New(),
			AchievementID: achievementID,
			Status:        "draft",
			CreatedAt:     time.Now(),
		}

		manager2.AddEntry(entry)
		history1 := manager2.GetHistory(achievementID)
		history2 := manager2.GetHistory(achievementID)

		// Modifying one should not affect the other
		history1[0].Status = "modified"
		assert.Equal(t, "draft", history2[0].Status)
	})
}

func TestClearHistory(t *testing.T) {
	manager := utils.NewAchievementHistoryManager()
	achievementID := uuid.New()

	entry := model.AchievementStatusLog{
		ID:            uuid.New(),
		AchievementID: achievementID,
		Status:        "draft",
		CreatedAt:     time.Now(),
	}

	manager.AddEntry(entry)
	assert.Equal(t, 1, manager.GetTotalEntries())

	manager.ClearHistory(achievementID)
	assert.Equal(t, 0, manager.GetTotalEntries())
	assert.Len(t, manager.GetHistory(achievementID), 0)
}

func TestGetTotalEntries(t *testing.T) {
	manager := utils.NewAchievementHistoryManager()

	t.Run("Empty manager", func(t *testing.T) {
		assert.Equal(t, 0, manager.GetTotalEntries())
	})

	t.Run("After adding entries", func(t *testing.T) {
		achievementID1 := uuid.New()
		achievementID2 := uuid.New()

		manager.AddEntry(model.AchievementStatusLog{
			ID:            uuid.New(),
			AchievementID: achievementID1,
			Status:        "draft",
			CreatedAt:     time.Now(),
		})
		manager.AddEntry(model.AchievementStatusLog{
			ID:            uuid.New(),
			AchievementID: achievementID1,
			Status:        "submitted",
			CreatedAt:     time.Now(),
		})
		manager.AddEntry(model.AchievementStatusLog{
			ID:            uuid.New(),
			AchievementID: achievementID2,
			Status:        "verified",
			CreatedAt:     time.Now(),
		})

		assert.Equal(t, 3, manager.GetTotalEntries())
	})
}

func TestAchievementHistoryConcurrency(t *testing.T) {
	manager := utils.NewAchievementHistoryManager()
	achievementID := uuid.New()

	// Test concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(index int) {
			entry := model.AchievementStatusLog{
				ID:            uuid.New(),
				AchievementID: achievementID,
				Status:        "draft",
				CreatedAt:     time.Now(),
			}
			manager.AddEntry(entry)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have 10 entries
	assert.Equal(t, 10, manager.GetTotalEntries())
	assert.Len(t, manager.GetHistory(achievementID), 10)
}
