package mocks

import (
	"context"
	mongodb "UASBE/app/model/MongoDB"
	model "UASBE/app/model/Postgresql"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
)

type MockAchievementRepository struct {
	mock.Mock
}

// AddAttachmentToAchievement implements repository.AchievementRepository.
func (m *MockAchievementRepository) AddAttachmentToAchievement(ctx context.Context, mongoAchievementID string, fileName string, fileURL string, fileType string) error {
	panic("unimplemented")
}

// GetStudentAchievements implements repository.AchievementRepository.
func (m *MockAchievementRepository) GetStudentAchievements(ctx context.Context, studentID uuid.UUID, page int, limit int) ([]model.AchievementWithStudent, int, error) {
	panic("unimplemented")
}

// GetStudentWithUserByID implements repository.AchievementRepository.
func (m *MockAchievementRepository) GetStudentWithUserByID(ctx context.Context, studentID uuid.UUID) (*model.StudentWithUser, error) {
	panic("unimplemented")
}

// GetAchievementsByID implements repository.AchievementRepository.
func (m *MockAchievementRepository) GetAchievementsByID(ctx context.Context, userID uuid.UUID) (*model.Users, error) {
	panic("unimplemented")
}

// GetUserByID implements repository.AchievementRepository.
func (m *MockAchievementRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*model.Users, error) {
	panic("unimplemented")
}

// UpdateAchievementInMongo implements repository.AchievementRepository.
func (m *MockAchievementRepository) UpdateAchievementInMongo(ctx context.Context, achievement mongodb.Achievement) error {
	panic("unimplemented")
}

// UpdateAchievementTimestamp implements repository.AchievementRepository.
func (m *MockAchievementRepository) UpdateAchievementTimestamp(ctx context.Context, achievementID uuid.UUID) error {
	panic("unimplemented")
}

func (m *MockAchievementRepository) GetStudentByUserID(ctx context.Context, userID uuid.UUID) (*model.Student, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Student), args.Error(1)
}

func (m *MockAchievementRepository) SaveAchievementMongo(ctx context.Context, achievement mongodb.Achievement) (string, error) {
	args := m.Called(ctx, achievement)
	return args.String(0), args.Error(1)
}

func (m *MockAchievementRepository) SaveAchievementReference(ctx context.Context, ref model.AchievementReference) error {
	args := m.Called(ctx, ref)
	return args.Error(0)
}

func (m *MockAchievementRepository) GetAchievementReferenceByID(ctx context.Context, achievementID uuid.UUID) (*model.AchievementReference, error) {
	args := m.Called(ctx, achievementID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AchievementReference), args.Error(1)
}

func (m *MockAchievementRepository) UpdateAchievementStatusToSubmitted(ctx context.Context, achievementID uuid.UUID) error {
	args := m.Called(ctx, achievementID)
	return args.Error(0)
}

func (m *MockAchievementRepository) GetAdvisorIDByStudentID(ctx context.Context, studentID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(ctx, studentID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockAchievementRepository) SoftDeleteAchievementMongo(ctx context.Context, mongoAchievementID string) error {
	args := m.Called(ctx, mongoAchievementID)
	return args.Error(0)
}

func (m *MockAchievementRepository) UpdateAchievementReferenceToDeleted(ctx context.Context, achievementID uuid.UUID) error {
	args := m.Called(ctx, achievementID)
	return args.Error(0)
}

func (m *MockAchievementRepository) GetLecturerByUserID(ctx context.Context, userID uuid.UUID) (*model.Lecturers, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Lecturers), args.Error(1)
}

func (m *MockAchievementRepository) GetStudentIDsByAdvisorID(ctx context.Context, advisorID uuid.UUID) ([]uuid.UUID, error) {
	args := m.Called(ctx, advisorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuid.UUID), args.Error(1)
}

func (m *MockAchievementRepository) GetAchievementsWithStudentInfo(ctx context.Context, studentIDs []uuid.UUID, status string, page, limit int) ([]model.AchievementWithStudent, int, error) {
	args := m.Called(ctx, studentIDs, status, page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.AchievementWithStudent), args.Int(1), args.Error(2)
}

func (m *MockAchievementRepository) GetAchievementDetailFromMongo(ctx context.Context, mongoAchievementID string) (*mongodb.Achievement, error) {
	args := m.Called(ctx, mongoAchievementID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mongodb.Achievement), args.Error(1)
}

func (m *MockAchievementRepository) UpdateAchievementStatusToVerified(ctx context.Context, achievementID uuid.UUID, lecturerID uuid.UUID) error {
	args := m.Called(ctx, achievementID, lecturerID)
	return args.Error(0)
}

func (m *MockAchievementRepository) GetStudentByID(ctx context.Context, studentID uuid.UUID) (*model.Student, error) {
	args := m.Called(ctx, studentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Student), args.Error(1)
}

func (m *MockAchievementRepository) UpdateAchievementStatusToRejected(ctx context.Context, achievementID uuid.UUID, rejectionNote string) error {
	args := m.Called(ctx, achievementID, rejectionNote)
	return args.Error(0)
}

func (m *MockAchievementRepository) GetAchievementStatusHistory(ctx context.Context, achievementID uuid.UUID) ([]model.AchievementStatusLog, error) {
	args := m.Called(ctx, achievementID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.AchievementStatusLog), args.Error(1)
}

func (m *MockAchievementRepository) LogAchievementStatusChange(ctx context.Context, log model.AchievementStatusLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockAchievementRepository) GetAllAchievementsForAdmin(ctx context.Context, filters model.AdminAchievementFilters, page, limit int) ([]model.AchievementWithStudent, int, error) {
	args := m.Called(ctx, filters, page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.AchievementWithStudent), args.Int(1), args.Error(2)
}

func (m *MockAchievementRepository) GetStatisticsByType(ctx context.Context, studentIDs []uuid.UUID, filters model.StatisticsFilters, mongoColl *mongo.Collection) ([]model.StatsByType, error) {
	args := m.Called(ctx, studentIDs, filters, mongoColl)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.StatsByType), args.Error(1)
}

func (m *MockAchievementRepository) GetStatisticsByPeriod(ctx context.Context, studentIDs []uuid.UUID, filters model.StatisticsFilters) ([]model.StatsByPeriod, error) {
	args := m.Called(ctx, studentIDs, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.StatsByPeriod), args.Error(1)
}

func (m *MockAchievementRepository) GetTopStudents(ctx context.Context, studentIDs []uuid.UUID, filters model.StatisticsFilters, limit int) ([]model.TopStudent, error) {
	args := m.Called(ctx, studentIDs, filters, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.TopStudent), args.Error(1)
}

func (m *MockAchievementRepository) GetLevelDistribution(ctx context.Context, studentIDs []uuid.UUID, filters model.StatisticsFilters, mongoColl *mongo.Collection) ([]model.LevelDistribution, error) {
	args := m.Called(ctx, studentIDs, filters, mongoColl)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.LevelDistribution), args.Error(1)
}

func (m *MockAchievementRepository) GetStatusDistribution(ctx context.Context, studentIDs []uuid.UUID, filters model.StatisticsFilters) ([]model.StatusDistribution, error) {
	args := m.Called(ctx, studentIDs, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.StatusDistribution), args.Error(1)
}

func (m *MockAchievementRepository) GetTotalAchievements(ctx context.Context, studentIDs []uuid.UUID, filters model.StatisticsFilters) (int, error) {
	args := m.Called(ctx, studentIDs, filters)
	return args.Int(0), args.Error(1)
}

func (m *MockAchievementRepository) GetAllStudentIDs(ctx context.Context) ([]uuid.UUID, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuid.UUID), args.Error(1)
}
