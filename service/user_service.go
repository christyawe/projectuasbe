package service

import (
	"context"
	"errors"
	"time"

	model "uas_backend/app/model/Postgresql"
	"uas_backend/app/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	// Business logic methods
	CreateUser(ctx context.Context, req model.CreateUserRequest) (*model.Users, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*model.UserResponse, error)
	GetUsers(ctx context.Context, page, limit int) (*UserListResponse, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, req model.UpdateUserRequest) (*model.Users, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	UpdateUserRole(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) (*model.Users, error)

	// Students & Lecturers methods
	GetStudents(ctx context.Context, page, limit int) (*StudentListResponse, error)
	GetStudentByID(ctx context.Context, studentID uuid.UUID) (*model.StudentWithUser, error)
	GetStudentAchievements(ctx context.Context, studentID uuid.UUID, page, limit int) (*model.AchievementListResponse, error)
	UpdateStudentAdvisor(ctx context.Context, studentID uuid.UUID, advisorID uuid.UUID) (*model.Student, error)
	GetLecturers(ctx context.Context, page, limit int) (*LecturerListResponse, error)
	GetLecturerAdvisees(ctx context.Context, lecturerID uuid.UUID, page, limit int) (*StudentListResponse, error)

	// HTTP endpoints
	GetUsersEndpoint(c *fiber.Ctx) error
	GetUserByIDEndpoint(c *fiber.Ctx) error
	CreateUserEndpoint(c *fiber.Ctx) error
	UpdateUserEndpoint(c *fiber.Ctx) error
	DeleteUserEndpoint(c *fiber.Ctx) error
	UpdateUserRoleEndpoint(c *fiber.Ctx) error

	// Students & Lecturers endpoints
	GetStudentsEndpoint(c *fiber.Ctx) error
	GetStudentByIDEndpoint(c *fiber.Ctx) error
	GetStudentAchievementsEndpoint(c *fiber.Ctx) error
	UpdateStudentAdvisorEndpoint(c *fiber.Ctx) error
	GetLecturersEndpoint(c *fiber.Ctx) error
	GetLecturerAdviseesEndpoint(c *fiber.Ctx) error
}

type userService struct {
	repo repository.UserRepository
}

type UserListResponse struct {
	Users      []model.UserResponse     `json:"users"`
	Pagination model.PaginationMetadata `json:"pagination"`
}

type StudentListResponse struct {
	Students   []model.StudentWithUser  `json:"students"`
	Pagination model.PaginationMetadata `json:"pagination"`
}

type LecturerListResponse struct {
	Lecturers  []model.LecturerWithUser `json:"lecturers"`
	Pagination model.PaginationMetadata `json:"pagination"`
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

// CreateUser membuat user baru
func (s *userService) CreateUser(ctx context.Context, req model.CreateUserRequest) (*model.Users, error) {
	// 1. Validasi username dan email belum ada
	usernameExists, err := s.repo.CheckUsernameExists(ctx, req.Username)
	if err != nil {
		return nil, errors.New("failed to check username")
	}
	if usernameExists {
		return nil, errors.New("username already exists")
	}

	emailExists, err := s.repo.CheckEmailExists(ctx, req.Email)
	if err != nil {
		return nil, errors.New("failed to check email")
	}
	if emailExists {
		return nil, errors.New("email already exists")
	}

	// 2. Validasi role exists
	_, err = s.repo.GetRoleByID(ctx, req.RoleID)
	if err != nil {
		return nil, errors.New("role not found")
	}

	// 3. Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// 4. Create user
	user := &model.Users{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FullName:     req.FullName,
		RoleID:       req.RoleID,
		ISActive:     req.IsActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = s.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, errors.New("failed to create user")
	}

	// 5. Create profile jika ada
	if req.ProfileType != "" && req.ProfileData != nil {
		if req.ProfileType == "student" {
			student := &model.Student{
				ID:            uuid.New(),
				UserID:        user.ID,
				StudentID:     req.ProfileData.StudentID,
				Program_Study: req.ProfileData.ProgramStudy,
				Academic_Year: req.ProfileData.AcademicYear,
				AdvisorID:     req.ProfileData.AdvisorID,
				Created_at:    time.Now(),
			}
			err = s.repo.CreateStudentProfile(ctx, student)
			if err != nil {
				return nil, errors.New("failed to create student profile")
			}
		} else if req.ProfileType == "lecturer" {
			lecturer := &model.Lecturers{
				ID:         uuid.New(),
				UserID:     user.ID,
				LecturerID: req.ProfileData.LecturerID,
				Department: req.ProfileData.Department,
				Created_at: time.Now(),
			}
			err = s.repo.CreateLecturerProfile(ctx, lecturer)
			if err != nil {
				return nil, errors.New("failed to create lecturer profile")
			}
		}
	}

	return user, nil
}

// GetUserByID mengambil user berdasarkan ID
func (s *userService) GetUserByID(ctx context.Context, userID uuid.UUID) (*model.UserResponse, error) {
	user, roleName, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return &model.UserResponse{
		ID:       user.ID,
		Username: user.Username,
		FullName: user.FullName,
		Role:     roleName,
	}, nil
}

// GetUsers mengambil semua users dengan pagination
func (s *userService) GetUsers(ctx context.Context, page, limit int) (*UserListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	users, roleNames, total, err := s.repo.GetAllUsers(ctx, page, limit)
	if err != nil {
		return nil, errors.New("failed to get users")
	}

	var userResponses []model.UserResponse
	for i, user := range users {
		userResponses = append(userResponses, model.UserResponse{
			ID:       user.ID,
			Username: user.Username,
			FullName: user.FullName,
			Role:     roleNames[i],
		})
	}

	totalPages := (total + limit - 1) / limit

	return &UserListResponse{
		Users: userResponses,
		Pagination: model.PaginationMetadata{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// UpdateUser mengupdate data user
func (s *userService) UpdateUser(ctx context.Context, userID uuid.UUID, req model.UpdateUserRequest) (*model.Users, error) {
	// Check if user exists
	_, _, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Check email if provided
	if req.Email != "" {
		emailExists, err := s.repo.CheckEmailExists(ctx, req.Email)
		if err != nil {
			return nil, errors.New("failed to check email")
		}
		if emailExists {
			return nil, errors.New("email already exists")
		}
	}

	err = s.repo.UpdateUser(ctx, userID, &req)
	if err != nil {
		return nil, errors.New("failed to update user")
	}

	user, _, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("failed to get updated user")
	}

	return user, nil
}

// DeleteUser menghapus user
func (s *userService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	// Check if user exists
	_, _, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	err = s.repo.DeleteUser(ctx, userID)
	if err != nil {
		return errors.New("failed to delete user")
	}

	return nil
}

// UpdateUserRole mengupdate role user
func (s *userService) UpdateUserRole(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) (*model.Users, error) {
	// Check if user exists
	_, _, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Check if role exists
	_, err = s.repo.GetRoleByID(ctx, roleID)
	if err != nil {
		return nil, errors.New("role not found")
	}

	err = s.repo.UpdateUserRole(ctx, userID, roleID)
	if err != nil {
		return nil, errors.New("failed to update user role")
	}

	user, _, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("failed to get updated user")
	}

	return user, nil
}

// HTTP Endpoints
func (s *userService) GetUsersEndpoint(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	result, err := s.GetUsers(c.Context(), page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get users"})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   result,
	})
}

func (s *userService) GetUserByIDEndpoint(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID format"})
	}

	user, err := s.GetUserByID(c.Context(), userID)
	if err != nil {
		switch err.Error() {
		case "user not found":
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(500).JSON(fiber.Map{"error": "Failed to get user"})
		}
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   user,
	})
}

func (s *userService) CreateUserEndpoint(c *fiber.Ctx) error {
	var req model.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON body"})
	}

	user, err := s.CreateUser(c.Context(), req)
	if err != nil {
		switch err.Error() {
		case "username already exists":
			return c.Status(409).JSON(fiber.Map{"error": err.Error()})
		case "email already exists":
			return c.Status(409).JSON(fiber.Map{"error": err.Error()})
		case "role not found":
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create user"})
		}
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "User created successfully",
		"data":    user,
	})
}

func (s *userService) UpdateUserEndpoint(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID format"})
	}

	var req model.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON body"})
	}

	user, err := s.UpdateUser(c.Context(), userID, req)
	if err != nil {
		switch err.Error() {
		case "user not found":
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		case "email already exists":
			return c.Status(409).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(500).JSON(fiber.Map{"error": "Failed to update user"})
		}
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "User updated successfully",
		"data":    user,
	})
}

func (s *userService) DeleteUserEndpoint(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID format"})
	}

	err = s.DeleteUser(c.Context(), userID)
	if err != nil {
		switch err.Error() {
		case "user not found":
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(500).JSON(fiber.Map{"error": "Failed to delete user"})
		}
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "User deleted successfully",
	})
}

func (s *userService) UpdateUserRoleEndpoint(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID format"})
	}

	var req model.UpdateUserRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON body"})
	}

	user, err := s.UpdateUserRole(c.Context(), userID, req.RoleID)
	if err != nil {
		switch err.Error() {
		case "user not found":
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		case "role not found":
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(500).JSON(fiber.Map{"error": "Failed to update user role"})
		}
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "User role updated successfully",
		"data":    user,
	})
}

// Students & Lecturers Business Logic Methods

// GetStudents mengambil semua students dengan pagination
func (s *userService) GetStudents(ctx context.Context, page, limit int) (*StudentListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	students, total, err := s.repo.GetAllStudents(ctx, page, limit)
	if err != nil {
		return nil, errors.New("failed to get students")
	}

	totalPages := (total + limit - 1) / limit

	return &StudentListResponse{
		Students: students,
		Pagination: model.PaginationMetadata{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// GetStudentByID mengambil student berdasarkan ID
func (s *userService) GetStudentByID(ctx context.Context, studentID uuid.UUID) (*model.StudentWithUser, error) {
	student, err := s.repo.GetStudentWithUserByID(ctx, studentID)
	if err != nil {
		return nil, errors.New("student not found")
	}

	return student, nil
}

// GetStudentAchievements mengambil achievements student
func (s *userService) GetStudentAchievements(ctx context.Context, studentID uuid.UUID, page, limit int) (*model.AchievementListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	achievements, total, err := s.repo.GetStudentAchievements(ctx, studentID, page, limit)
	if err != nil {
		return nil, errors.New("failed to get student achievements")
	}

	totalPages := (total + limit - 1) / limit

	return &model.AchievementListResponse{
		Achievements: achievements,
		Pagination: model.PaginationMetadata{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// UpdateStudentAdvisor mengupdate advisor student
func (s *userService) UpdateStudentAdvisor(ctx context.Context, studentID uuid.UUID, advisorID uuid.UUID) (*model.Student, error) {
	// Check if student exists
	_, err := s.repo.GetStudentWithUserByID(ctx, studentID)
	if err != nil {
		return nil, errors.New("student not found")
	}

	// Check if advisor exists and is a lecturer
	_, err = s.repo.GetLecturerByID(ctx, advisorID)
	if err != nil {
		return nil, errors.New("advisor not found or not a lecturer")
	}

	err = s.repo.UpdateStudentAdvisor(ctx, studentID, advisorID)
	if err != nil {
		return nil, errors.New("failed to update student advisor")
	}

	student, err := s.repo.GetStudentByID(ctx, studentID)
	if err != nil {
		return nil, errors.New("failed to get updated student")
	}

	return student, nil
}

// GetLecturers mengambil semua lecturers dengan pagination
func (s *userService) GetLecturers(ctx context.Context, page, limit int) (*LecturerListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	lecturers, total, err := s.repo.GetAllLecturers(ctx, page, limit)
	if err != nil {
		return nil, errors.New("failed to get lecturers")
	}

	totalPages := (total + limit - 1) / limit

	return &LecturerListResponse{
		Lecturers: lecturers,
		Pagination: model.PaginationMetadata{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// GetLecturerAdvisees mengambil mahasiswa bimbingan lecturer
func (s *userService) GetLecturerAdvisees(ctx context.Context, lecturerID uuid.UUID, page, limit int) (*StudentListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Check if lecturer exists
	_, err := s.repo.GetLecturerByID(ctx, lecturerID)
	if err != nil {
		return nil, errors.New("lecturer not found")
	}

	students, total, err := s.repo.GetStudentsByAdvisorID(ctx, lecturerID, page, limit)
	if err != nil {
		return nil, errors.New("failed to get lecturer advisees")
	}

	totalPages := (total + limit - 1) / limit

	return &StudentListResponse{
		Students: students,
		Pagination: model.PaginationMetadata{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// Students & Lecturers HTTP Endpoints

func (s *userService) GetStudentsEndpoint(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	result, err := s.GetStudents(c.Context(), page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get students"})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   result,
	})
}

func (s *userService) GetStudentByIDEndpoint(c *fiber.Ctx) error {
	studentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid student ID format"})
	}

	student, err := s.GetStudentByID(c.Context(), studentID)
	if err != nil {
		switch err.Error() {
		case "student not found":
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(500).JSON(fiber.Map{"error": "Failed to get student"})
		}
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   student,
	})
}

func (s *userService) GetStudentAchievementsEndpoint(c *fiber.Ctx) error {
	studentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid student ID format"})
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	result, err := s.GetStudentAchievements(c.Context(), studentID, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get student achievements"})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   result,
	})
}

func (s *userService) UpdateStudentAdvisorEndpoint(c *fiber.Ctx) error {
	studentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid student ID format"})
	}

	var req model.UpdateAdvisorRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON body"})
	}

	student, err := s.UpdateStudentAdvisor(c.Context(), studentID, req.AdvisorID)
	if err != nil {
		switch err.Error() {
		case "student not found":
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		case "advisor not found or not a lecturer":
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(500).JSON(fiber.Map{"error": "Failed to update student advisor"})
		}
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Student advisor updated successfully",
		"data":    student,
	})
}

func (s *userService) GetLecturersEndpoint(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	result, err := s.GetLecturers(c.Context(), page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get lecturers"})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   result,
	})
}

func (s *userService) GetLecturerAdviseesEndpoint(c *fiber.Ctx) error {
	lecturerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid lecturer ID format"})
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	result, err := s.GetLecturerAdvisees(c.Context(), lecturerID, page, limit)
	if err != nil {
		switch err.Error() {
		case "lecturer not found":
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(500).JSON(fiber.Map{"error": "Failed to get lecturer advisees"})
		}
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   result,
	})
}
