package test

import (
	"context"
	"errors"
	"testing"
	"time"
	mongodb "UASBE/app/model/MongoDB"
	model "UASBE/app/model/Postgresql"
	"UASBE/app/service"
	"UASBE/test/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAchievementService_SubmitPrestasi(t *testing.T) {
	ctx := context.Background()

	t.Run("Successful submission", func(t *testing.T) {
		mockRepo := new(mocks.MockAchievementRepository)
		achievementService := service.NewAchievementService(mockRepo)

		userID := uuid.New()
		studentID := uuid.New()
		student := &model.Student{
			ID:            studentID,
			UserID:        userID,
			StudentID:     "12345",
			Program_Study: "Computer Science",
		}

		achievement := mongodb.Achievement{
			AchievementType: "competition",
			Title:           "Test Achievement",
			Description:     "Test Description",
		}

		mockRepo.On("GetStudentByUserID", ctx, userID).Return(student, nil)
		mockRepo.On("SaveAchievementMongo", ctx, mock.AnythingOfType("mongodb.Achievement")).Return("mongo_id_123", nil)
		mockRepo.On("SaveAchievementReference", ctx, mock.AnythingOfType("model.AchievementReference")).Return(nil)

		result, err := achievementService.SubmitPrestasi(ctx, userID, achievement)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "draft", result.Status)
		assert.Equal(t, studentID, result.StudentID)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Student not found", func(t *testing.T) {
		mockRepo := new(mocks.MockAchievementRepository)
		achievementService := service.NewAchievementService(mockRepo)

		userID := uuid.New()
		achievement := mongodb.Achievement{
			AchievementType: "competition",
			Title:           "Test Achievement",
		}

		mockRepo.On("GetStudentByUserID", ctx, userID).Return(nil, errors.New("student not found"))

		result, err := achievementService.SubmitPrestasi(ctx, userID, achievement)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "student data not found for this user", err.Error())

		mockRepo.AssertExpectations(t)
	})
}

func TestAchievementService_SubmitForVerification(t *testing.T) {
	ctx := context.Background()

	t.Run("Successful submission for verification", func(t *testing.T) {
		mockRepo := new(mocks.MockAchievementRepository)
		achievementService := service.NewAchievementService(mockRepo)

		userID := uuid.New()
		studentID := uuid.New()
		achievementID := uuid.New()

		student := &model.Student{
			ID:     studentID,
			UserID: userID,
		}

		ref := &model.AchievementReference{
			ID:                 achievementID,
			StudentID:          studentID,
			MongoAchievementID: "mongo_id",
			Status:             "draft",
		}

		updatedRef := &model.AchievementReference{
			ID:                 achievementID,
			StudentID:          studentID,
			MongoAchievementID: "mongo_id",
			Status:             "submitted",
		}

		mockRepo.On("GetStudentByUserID", ctx, userID).Return(student, nil)
		mockRepo.On("GetAchievementReferenceByID", ctx, achievementID).Return(ref, nil).Once()
		mockRepo.On("UpdateAchievementStatusToSubmitted", ctx, achievementID).Return(nil)
		mockRepo.On("GetAchievementReferenceByID", ctx, achievementID).Return(updatedRef, nil).Once()

		result, err := achievementService.SubmitForVerification(ctx, userID, achievementID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "submitted", result.Status)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Achievement not in draft status", func(t *testing.T) {
		mockRepo := new(mocks.MockAchievementRepository)
		achievementService := service.NewAchievementService(mockRepo)

		userID := uuid.New()
		studentID := uuid.New()
		achievementID := uuid.New()

		student := &model.Student{
			ID:     studentID,
			UserID: userID,
		}

		ref := &model.AchievementReference{
			ID:        achievementID,
			StudentID: studentID,
			Status:    "submitted", // Already submitted
		}

		mockRepo.On("GetStudentByUserID", ctx, userID).Return(student, nil)
		mockRepo.On("GetAchievementReferenceByID", ctx, achievementID).Return(ref, nil)

		result, err := achievementService.SubmitForVerification(ctx, userID, achievementID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "achievement must be in 'draft' status to submit", err.Error())

		mockRepo.AssertExpectations(t)
	})

	t.Run("Unauthorized - not student's achievement", func(t *testing.T) {
		mockRepo := new(mocks.MockAchievementRepository)
		achievementService := service.NewAchievementService(mockRepo)

		userID := uuid.New()
		studentID := uuid.New()
		otherStudentID := uuid.New()
		achievementID := uuid.New()

		student := &model.Student{
			ID:     studentID,
			UserID: userID,
		}

		ref := &model.AchievementReference{
			ID:        achievementID,
			StudentID: otherStudentID, // Different student
			Status:    "draft",
		}

		mockRepo.On("GetStudentByUserID", ctx, userID).Return(student, nil)
		mockRepo.On("GetAchievementReferenceByID", ctx, achievementID).Return(ref, nil)

		result, err := achievementService.SubmitForVerification(ctx, userID, achievementID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "unauthorized: achievement does not belong to this student", err.Error())

		mockRepo.AssertExpectations(t)
	})
}

func TestAchievementService_DeleteDraftAchievement(t *testing.T) {
	ctx := context.Background()

	t.Run("Successful deletion", func(t *testing.T) {
		mockRepo := new(mocks.MockAchievementRepository)
		achievementService := service.NewAchievementService(mockRepo)

		userID := uuid.New()
		studentID := uuid.New()
		achievementID := uuid.New()

		student := &model.Student{
			ID:     studentID,
			UserID: userID,
		}

		ref := &model.AchievementReference{
			ID:                 achievementID,
			StudentID:          studentID,
			MongoAchievementID: "mongo_id",
			Status:             "draft",
		}

		mockRepo.On("GetStudentByUserID", ctx, userID).Return(student, nil)
		mockRepo.On("GetAchievementReferenceByID", ctx, achievementID).Return(ref, nil)
		mockRepo.On("SoftDeleteAchievementMongo", ctx, "mongo_id").Return(nil)
		mockRepo.On("UpdateAchievementReferenceToDeleted", ctx, achievementID).Return(nil)

		err := achievementService.DeleteDraftAchievement(ctx, userID, achievementID)

		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Cannot delete non-draft achievement", func(t *testing.T) {
		mockRepo := new(mocks.MockAchievementRepository)
		achievementService := service.NewAchievementService(mockRepo)

		userID := uuid.New()
		studentID := uuid.New()
		achievementID := uuid.New()

		student := &model.Student{
			ID:     studentID,
			UserID: userID,
		}

		ref := &model.AchievementReference{
			ID:        achievementID,
			StudentID: studentID,
			Status:    "verified", // Not draft
		}

		mockRepo.On("GetStudentByUserID", ctx, userID).Return(student, nil)
		mockRepo.On("GetAchievementReferenceByID", ctx, achievementID).Return(ref, nil)

		err := achievementService.DeleteDraftAchievement(ctx, userID, achievementID)

		assert.Error(t, err)
		assert.Equal(t, "only draft achievements can be deleted", err.Error())

		mockRepo.AssertExpectations(t)
	})
}

func TestAchievementService_VerifyAchievement(t *testing.T) {
	ctx := context.Background()

	t.Run("Successful verification", func(t *testing.T) {
		mockRepo := new(mocks.MockAchievementRepository)
		achievementService := service.NewAchievementService(mockRepo)

		userID := uuid.New()
		lecturerID := uuid.New()
		studentID := uuid.New()
		achievementID := uuid.New()

		lecturer := &model.Lecturers{
			ID:     lecturerID,
			UserID: userID,
		}

		ref := &model.AchievementReference{
			ID:        achievementID,
			StudentID: studentID,
			Status:    "submitted",
		}

		student := &model.Student{
			ID:        studentID,
			AdvisorID: lecturerID,
		}

		mockRepo.On("GetLecturerByUserID", ctx, userID).Return(lecturer, nil)
		mockRepo.On("GetAchievementReferenceByID", ctx, achievementID).Return(ref, nil).Once()
		mockRepo.On("GetStudentByID", ctx, studentID).Return(student, nil)
		mockRepo.On("UpdateAchievementStatusToVerified", ctx, achievementID, lecturerID).Return(nil)

		updatedRef := &model.AchievementReference{
			ID:        achievementID,
			StudentID: studentID,
			Status:    "verified",
		}
		mockRepo.On("GetAchievementReferenceByID", ctx, achievementID).Return(updatedRef, nil).Once()

		result, err := achievementService.VerifyAchievement(ctx, userID, achievementID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "verified", result.Status)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Achievement not in submitted status", func(t *testing.T) {
		mockRepo := new(mocks.MockAchievementRepository)
		achievementService := service.NewAchievementService(mockRepo)

		userID := uuid.New()
		lecturerID := uuid.New()
		achievementID := uuid.New()

		lecturer := &model.Lecturers{
			ID:     lecturerID,
			UserID: userID,
		}

		ref := &model.AchievementReference{
			ID:     achievementID,
			Status: "draft", // Not submitted
		}

		mockRepo.On("GetLecturerByUserID", ctx, userID).Return(lecturer, nil)
		mockRepo.On("GetAchievementReferenceByID", ctx, achievementID).Return(ref, nil)

		result, err := achievementService.VerifyAchievement(ctx, userID, achievementID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "achievement must be in 'submitted' status to verify", err.Error())

		mockRepo.AssertExpectations(t)
	})

	t.Run("Unauthorized - not advisor", func(t *testing.T) {
		mockRepo := new(mocks.MockAchievementRepository)
		achievementService := service.NewAchievementService(mockRepo)

		userID := uuid.New()
		lecturerID := uuid.New()
		otherLecturerID := uuid.New()
		studentID := uuid.New()
		achievementID := uuid.New()

		lecturer := &model.Lecturers{
			ID:     lecturerID,
			UserID: userID,
		}

		ref := &model.AchievementReference{
			ID:        achievementID,
			StudentID: studentID,
			Status:    "submitted",
		}

		student := &model.Student{
			ID:        studentID,
			AdvisorID: otherLecturerID, // Different advisor
		}

		mockRepo.On("GetLecturerByUserID", ctx, userID).Return(lecturer, nil)
		mockRepo.On("GetAchievementReferenceByID", ctx, achievementID).Return(ref, nil)
		mockRepo.On("GetStudentByID", ctx, studentID).Return(student, nil)

		result, err := achievementService.VerifyAchievement(ctx, userID, achievementID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "unauthorized: you can only verify achievements of your advisees", err.Error())

		mockRepo.AssertExpectations(t)
	})
}

func TestAchievementService_RejectAchievement(t *testing.T) {
	ctx := context.Background()

	t.Run("Successful rejection", func(t *testing.T) {
		mockRepo := new(mocks.MockAchievementRepository)
		achievementService := service.NewAchievementService(mockRepo)

		userID := uuid.New()
		lecturerID := uuid.New()
		studentID := uuid.New()
		achievementID := uuid.New()
		rejectionNote := "Needs more documentation"

		lecturer := &model.Lecturers{
			ID:     lecturerID,
			UserID: userID,
		}

		ref := &model.AchievementReference{
			ID:        achievementID,
			StudentID: studentID,
			Status:    "submitted",
		}

		student := &model.Student{
			ID:        studentID,
			AdvisorID: lecturerID,
		}

		mockRepo.On("GetLecturerByUserID", ctx, userID).Return(lecturer, nil)
		mockRepo.On("GetAchievementReferenceByID", ctx, achievementID).Return(ref, nil).Once()
		mockRepo.On("GetStudentByID", ctx, studentID).Return(student, nil)
		mockRepo.On("UpdateAchievementStatusToRejected", ctx, achievementID, rejectionNote).Return(nil)

		updatedRef := &model.AchievementReference{
			ID:            achievementID,
			StudentID:     studentID,
			Status:        "rejected",
			RejectionNote: &rejectionNote,
		}
		mockRepo.On("GetAchievementReferenceByID", ctx, achievementID).Return(updatedRef, nil).Once()

		result, err := achievementService.RejectAchievement(ctx, userID, achievementID, rejectionNote)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "rejected", result.Status)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Empty rejection note", func(t *testing.T) {
		mockRepo := new(mocks.MockAchievementRepository)
		achievementService := service.NewAchievementService(mockRepo)

		userID := uuid.New()
		achievementID := uuid.New()

		result, err := achievementService.RejectAchievement(ctx, userID, achievementID, "")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "rejection note is required", err.Error())

		mockRepo.AssertExpectations(t)
	})
}

func TestAchievementService_GetStudentAchievements(t *testing.T) {
	ctx := context.Background()

	t.Run("Successful retrieval", func(t *testing.T) {
		mockRepo := new(mocks.MockAchievementRepository)
		achievementService := service.NewAchievementService(mockRepo)

		userID := uuid.New()
		lecturerID := uuid.New()
		studentID1 := uuid.New()
		studentID2 := uuid.New()

		lecturer := &model.Lecturers{
			ID:     lecturerID,
			UserID: userID,
		}

		studentIDs := []uuid.UUID{studentID1, studentID2}

		achievements := []model.AchievementWithStudent{
			{
				ID:                 uuid.New(),
				StudentID:          studentID1,
				MongoAchievementID: "mongo_id_1",
				Status:             "submitted",
			},
		}

		mockRepo.On("GetLecturerByUserID", ctx, userID).Return(lecturer, nil)
		mockRepo.On("GetStudentIDsByAdvisorID", ctx, lecturerID).Return(studentIDs, nil)
		mockRepo.On("GetAchievementsWithStudentInfo", ctx, studentIDs, "", 1, 10).Return(achievements, 1, nil)
		mockRepo.On("GetAchievementDetailFromMongo", ctx, "mongo_id_1").Return(&mongodb.Achievement{}, nil)

		result, err := achievementService.GetStudentAchievements(ctx, userID, "", 1, 10)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Achievements, 1)
		assert.Equal(t, 1, result.Pagination.Total)

		mockRepo.AssertExpectations(t)
	})

	t.Run("No students found", func(t *testing.T) {
		mockRepo := new(mocks.MockAchievementRepository)
		achievementService := service.NewAchievementService(mockRepo)

		userID := uuid.New()
		lecturerID := uuid.New()

		lecturer := &model.Lecturers{
			ID:     lecturerID,
			UserID: userID,
		}

		mockRepo.On("GetLecturerByUserID", ctx, userID).Return(lecturer, nil)
		mockRepo.On("GetStudentIDsByAdvisorID", ctx, lecturerID).Return([]uuid.UUID{}, nil)

		result, err := achievementService.GetStudentAchievements(ctx, userID, "", 1, 10)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Achievements, 0)
		assert.Equal(t, 0, result.Pagination.Total)

		mockRepo.AssertExpectations(t)
	})
}

func TestAchievementService_GetAchievementHistory(t *testing.T) {
	ctx := context.Background()

	t.Run("Student can view own achievement history", func(t *testing.T) {
		mockRepo := new(mocks.MockAchievementRepository)
		achievementService := service.NewAchievementService(mockRepo)

		userID := uuid.New()
		studentID := uuid.New()
		achievementID := uuid.New()

		student := &model.Student{
			ID:     studentID,
			UserID: userID,
		}

		ref := &model.AchievementReference{
			ID:        achievementID,
			StudentID: studentID,
		}

		timeline := []model.AchievementStatusLog{
			{
				ID:            uuid.New(),
				AchievementID: achievementID,
				Status:        "draft",
				CreatedAt:     time.Now(),
			},
		}

		mockRepo.On("GetAchievementReferenceByID", ctx, achievementID).Return(ref, nil)
		mockRepo.On("GetStudentByUserID", ctx, userID).Return(student, nil)
		mockRepo.On("GetAchievementStatusHistory", ctx, achievementID).Return(timeline, nil)

		result, err := achievementService.GetAchievementHistory(ctx, userID, achievementID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Timeline, 1)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Unauthorized user", func(t *testing.T) {
		mockRepo := new(mocks.MockAchievementRepository)
		achievementService := service.NewAchievementService(mockRepo)

		userID := uuid.New()
		otherStudentID := uuid.New()
		achievementID := uuid.New()

		ref := &model.AchievementReference{
			ID:        achievementID,
			StudentID: otherStudentID,
		}

		mockRepo.On("GetAchievementReferenceByID", ctx, achievementID).Return(ref, nil)
		mockRepo.On("GetStudentByUserID", ctx, userID).Return(nil, errors.New("not a student"))
		mockRepo.On("GetLecturerByUserID", ctx, userID).Return(nil, errors.New("not a lecturer"))

		result, err := achievementService.GetAchievementHistory(ctx, userID, achievementID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "unauthorized")

		mockRepo.AssertExpectations(t)
	})
}
