package mocks

import (
	"context"
	model "UASBE/app/model/Postgresql"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

// GetAllLecturers implements repository.UserRepository.
func (m *MockUserRepository) GetAllLecturers(ctx context.Context, page int, limit int) ([]model.LecturerWithUser, int, error) {
	panic("unimplemented")
}

// GetAllStudents implements repository.UserRepository.
func (m *MockUserRepository) GetAllStudents(ctx context.Context, page int, limit int) ([]model.StudentWithUser, int, error) {
	panic("unimplemented")
}

// GetLecturerByID implements repository.UserRepository.
func (m *MockUserRepository) GetLecturerByID(ctx context.Context, lecturerID uuid.UUID) (*model.Lecturers, error) {
	panic("unimplemented")
}

// GetStudentAchievements implements repository.UserRepository.
func (m *MockUserRepository) GetStudentAchievements(ctx context.Context, studentID uuid.UUID, page int, limit int) ([]model.AchievementWithStudent, int, error) {
	panic("unimplemented")
}

// GetStudentByID implements repository.UserRepository.
func (m *MockUserRepository) GetStudentByID(ctx context.Context, studentID uuid.UUID) (*model.Student, error) {
	panic("unimplemented")
}

// GetStudentWithUserByID implements repository.UserRepository.
func (m *MockUserRepository) GetStudentWithUserByID(ctx context.Context, studentID uuid.UUID) (*model.StudentWithUser, error) {
	panic("unimplemented")
}

// GetStudentsByAdvisorID implements repository.UserRepository.
func (m *MockUserRepository) GetStudentsByAdvisorID(ctx context.Context, advisorID uuid.UUID, page int, limit int) ([]model.StudentWithUser, int, error) {
	panic("unimplemented")
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *model.Users) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*model.Users, string, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*model.Users), args.String(1), args.Error(2)
}

func (m *MockUserRepository) GetAllUsers(ctx context.Context, page, limit int) ([]model.Users, []string, int, error) {
	args := m.Called(ctx, page, limit)
	if args.Get(0) == nil {
		return nil, nil, 0, args.Error(3)
	}
	return args.Get(0).([]model.Users), args.Get(1).([]string), args.Int(2), args.Error(3)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, userID uuid.UUID, req *model.UpdateUserRequest) error {
	args := m.Called(ctx, userID, req)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateUserRole(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockUserRepository) CreateStudentProfile(ctx context.Context, student *model.Student) error {
	args := m.Called(ctx, student)
	return args.Error(0)
}

func (m *MockUserRepository) CreateLecturerProfile(ctx context.Context, lecturer *model.Lecturers) error {
	args := m.Called(ctx, lecturer)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateStudentAdvisor(ctx context.Context, studentID uuid.UUID, advisorID uuid.UUID) error {
	args := m.Called(ctx, studentID, advisorID)
	return args.Error(0)
}

func (m *MockUserRepository) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) GetRoleByID(ctx context.Context, roleID uuid.UUID) (*model.Roles, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Roles), args.Error(1)
}
