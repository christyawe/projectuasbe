package test

import (
	"context"
	"errors"
	"testing"
	model "UASBE/app/model/Postgresql"
	"UASBE/app/service"
	"UASBE/test/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserService_CreateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("Successful user creation", func(t *testing.T) {
		mockRepo := new(mocks.MockUserRepository)
		userService := service.NewUserService(mockRepo)

		roleID := uuid.New()
		role := &model.Roles{ID: roleID, Name: "student"}
		req := model.CreateUserRequest{
			Username:    "newuser",
			Email:       "new@example.com",
			Password:    "password123",
			FullName:    "New User",
			RoleID:      roleID,
			IsActive:    true,
			ProfileType: "",
		}

		mockRepo.On("CheckUsernameExists", ctx, "newuser").Return(false, nil)
		mockRepo.On("CheckEmailExists", ctx, "new@example.com").Return(false, nil)
		mockRepo.On("GetRoleByID", ctx, roleID).Return(role, nil)
		mockRepo.On("CreateUser", ctx, mock.AnythingOfType("*model.Users")).Return(nil)

		user, err := userService.CreateUser(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "newuser", user.Username)
		assert.Equal(t, "new@example.com", user.Email)
		assert.Equal(t, "New User", user.FullName)
		assert.Equal(t, roleID, user.RoleID)
		assert.True(t, user.ISActive)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Username already exists", func(t *testing.T) {
		mockRepo := new(mocks.MockUserRepository)
		userService := service.NewUserService(mockRepo)

		roleID := uuid.New()
		req := model.CreateUserRequest{
			Username:    "existinguser",
			Email:       "new@example.com",
			Password:    "password123",
			FullName:    "New User",
			RoleID:      roleID,
			IsActive:    true,
			ProfileType: "",
		}

		mockRepo.On("CheckUsernameExists", ctx, "existinguser").Return(true, nil)

		user, err := userService.CreateUser(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, "username already exists", err.Error())

		mockRepo.AssertExpectations(t)
	})

	t.Run("Email already exists", func(t *testing.T) {
		mockRepo := new(mocks.MockUserRepository)
		userService := service.NewUserService(mockRepo)

		roleID := uuid.New()
		req := model.CreateUserRequest{
			Username:    "newuser",
			Email:       "existing@example.com",
			Password:    "password123",
			FullName:    "New User",
			RoleID:      roleID,
			IsActive:    true,
			ProfileType: "",
		}

		mockRepo.On("CheckUsernameExists", ctx, "newuser").Return(false, nil)
		mockRepo.On("CheckEmailExists", ctx, "existing@example.com").Return(true, nil)

		user, err := userService.CreateUser(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, "email already exists", err.Error())

		mockRepo.AssertExpectations(t)
	})

	t.Run("Role not found", func(t *testing.T) {
		mockRepo := new(mocks.MockUserRepository)
		userService := service.NewUserService(mockRepo)

		roleID := uuid.New()
		req := model.CreateUserRequest{
			Username:    "newuser",
			Email:       "new@example.com",
			Password:    "password123",
			FullName:    "New User",
			RoleID:      roleID,
			IsActive:    true,
			ProfileType: "",
		}

		mockRepo.On("CheckUsernameExists", ctx, "newuser").Return(false, nil)
		mockRepo.On("CheckEmailExists", ctx, "new@example.com").Return(false, nil)
		mockRepo.On("GetRoleByID", ctx, roleID).Return(nil, errors.New("role not found"))

		user, err := userService.CreateUser(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, "role not found", err.Error())

		mockRepo.AssertExpectations(t)
	})

	t.Run("Successful user creation with student profile", func(t *testing.T) {
		mockRepo := new(mocks.MockUserRepository)
		userService := service.NewUserService(mockRepo)

		roleID := uuid.New()
		advisorID := uuid.New()
		role := &model.Roles{ID: roleID, Name: "student"}
		req := model.CreateUserRequest{
			Username:    "newstudent",
			Email:       "student@example.com",
			Password:    "password123",
			FullName:    "New Student",
			RoleID:      roleID,
			IsActive:    true,
			ProfileType: "student",
			ProfileData: &model.ProfileData{
				StudentID:    "12345",
				ProgramStudy: "Computer Science",
				AcademicYear: "2023",
				AdvisorID:    &advisorID,
			},
		}

		mockRepo.On("CheckUsernameExists", ctx, "newstudent").Return(false, nil)
		mockRepo.On("CheckEmailExists", ctx, "student@example.com").Return(false, nil)
		mockRepo.On("GetRoleByID", ctx, roleID).Return(role, nil)
		mockRepo.On("CreateUser", ctx, mock.AnythingOfType("*model.Users")).Return(nil)
		mockRepo.On("CreateStudentProfile", ctx, mock.AnythingOfType("*model.Student")).Return(nil)

		user, err := userService.CreateUser(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "newstudent", user.Username)

		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_GetUsers(t *testing.T) {
	ctx := context.Background()

	t.Run("Successful get users", func(t *testing.T) {
		mockRepo := new(mocks.MockUserRepository)
		userService := service.NewUserService(mockRepo)

		users := []*model.Users{
			{ID: uuid.New(), Username: "user1", FullName: "User One"},
			{ID: uuid.New(), Username: "user2", FullName: "User Two"},
		}
		roleNames := []string{"admin", "student"}

		mockRepo.On("GetAllUsers", ctx, 1, 10).Return(users, roleNames, 2, nil)

		result, err := userService.GetUsers(ctx, 1, 10)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Users, 2)
		assert.Equal(t, 2, result.Pagination.Total)
		assert.Equal(t, 1, result.Pagination.Page)
		assert.Equal(t, 10, result.Pagination.Limit)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Empty users list", func(t *testing.T) {
		mockRepo := new(mocks.MockUserRepository)
		userService := service.NewUserService(mockRepo)

		users := []*model.Users{}
		roleNames := []string{}

		mockRepo.On("GetAllUsers", ctx, 1, 10).Return(users, roleNames, 0, nil)

		result, err := userService.GetUsers(ctx, 1, 10)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Users, 0)
		assert.Equal(t, 0, result.Pagination.Total)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository error", func(t *testing.T) {
		mockRepo := new(mocks.MockUserRepository)
		userService := service.NewUserService(mockRepo)

		mockRepo.On("GetAllUsers", ctx, 1, 10).Return(nil, nil, 0, errors.New("database error"))

		result, err := userService.GetUsers(ctx, 1, 10)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "failed to get users", err.Error())

		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("Successful user update", func(t *testing.T) {
		mockRepo := new(mocks.MockUserRepository)
		userService := service.NewUserService(mockRepo)

		userID := uuid.New()
		existingUser := &model.Users{ID: userID, Username: "testuser", Email: "old@example.com"}
		req := model.UpdateUserRequest{
			Email:    "new@example.com",
			FullName: "Updated Name",
		}

		mockRepo.On("GetUserByID", ctx, userID).Return(existingUser, "student", nil).Twice()
		mockRepo.On("CheckEmailExists", ctx, "new@example.com").Return(false, nil)
		mockRepo.On("UpdateUser", ctx, userID, &req).Return(nil)

		user, err := userService.UpdateUser(ctx, userID, req)

		assert.NoError(t, err)
		assert.NotNil(t, user)

		mockRepo.AssertExpectations(t)
	})

	t.Run("User not found", func(t *testing.T) {
		mockRepo := new(mocks.MockUserRepository)
		userService := service.NewUserService(mockRepo)

		userID := uuid.New()
		req := model.UpdateUserRequest{
			Email:    "new@example.com",
			FullName: "Updated Name",
		}

		mockRepo.On("GetUserByID", ctx, userID).Return(nil, "", errors.New("user not found"))

		user, err := userService.UpdateUser(ctx, userID, req)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, "user not found", err.Error())

		mockRepo.AssertExpectations(t)
	})
}
