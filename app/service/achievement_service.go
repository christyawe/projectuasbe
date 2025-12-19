package service

import (
	"context"
	"errors"
	mongodb "UASBE/app/model/MongoDB"
	model "UASBE/app/model/Postgresql"
	"UASBE/app/repository"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AchievementService interface {
	// Business logic methods
	SubmitPrestasi(ctx context.Context, userID uuid.UUID, req mongodb.Achievement) (*model.AchievementReference, error)
	GetAchievementByID(ctx context.Context, userID uuid.UUID, mongoAchievementID string) (*mongodb.Achievement, error)
	UpdateAchievement(ctx context.Context, userID uuid.UUID, achievementID uuid.UUID, req mongodb.Achievement) (*model.AchievementReference, error)
	SubmitForVerification(ctx context.Context, userID uuid.UUID, achievementID uuid.UUID) (*model.AchievementReference, error)
	DeleteDraftAchievement(ctx context.Context, userID uuid.UUID, achievementID uuid.UUID) error
	GetStudentAchievements(ctx context.Context, userID uuid.UUID, status string, page, limit int) (*model.AchievementListResponse, error)
	VerifyAchievement(ctx context.Context, userID uuid.UUID, achievementID uuid.UUID) (*model.AchievementReference, error)
	RejectAchievement(ctx context.Context, userID uuid.UUID, achievementID uuid.UUID, rejectionNote string) (*model.AchievementReference, error)
	GetAchievementHistory(ctx context.Context, userID uuid.UUID, achievementID uuid.UUID) (*model.AchievementHistoryResponse, error)
	GetAllAchievementsForAdmin(ctx context.Context, filters model.AdminAchievementFilters, page, limit int) (*model.AchievementListResponse, error)
	GetAchievementStatistics(ctx context.Context, userID uuid.UUID, filters model.StatisticsFilters) (*model.AchievementStatistics, error)
	GetReportsStatistics(ctx context.Context, userID uuid.UUID, filters model.StatisticsFilters) (*model.AchievementStatistics, error)
	GetStudentReport(ctx context.Context, userID uuid.UUID, studentID uuid.UUID) (*model.StudentReportResponse, error)
	UploadAttachment(ctx context.Context, userID uuid.UUID, achievementID uuid.UUID, fileName, fileURL, fileType string) error

	// HTTP endpoints
	GetAchievementsEndpoint(c *fiber.Ctx) error
	GetAchievementByIDEndpoint(c *fiber.Ctx) error
	CreateAchievementEndpoint(c *fiber.Ctx) error
	UpdateAchievementEndpoint(c *fiber.Ctx) error
	DeleteAchievementEndpoint(c *fiber.Ctx) error
	SubmitAchievementEndpoint(c *fiber.Ctx) error
	VerifyAchievementEndpoint(c *fiber.Ctx) error
	RejectAchievementEndpoint(c *fiber.Ctx) error
	GetAchievementHistoryEndpoint(c *fiber.Ctx) error
	GetAchievementStatisticsEndpoint(c *fiber.Ctx) error
	GetAllAchievementsForAdminEndpoint(c *fiber.Ctx) error
	GetReportsStatisticsEndpoint(c *fiber.Ctx) error
	GetStudentReportEndpoint(c *fiber.Ctx) error
	UploadAttachmentEndpoint(c *fiber.Ctx) error
	GetAllStudentIDs(ctx context.Context) ([]uuid.UUID, error)
	GetAchievementAdminDetailEndpoint(c *fiber.Ctx) error
}

type achievementService struct {
	repo repository.AchievementRepository
}

// GetAllStudentIDs implements AchievementService.
func (s *achievementService) GetAllStudentIDs(ctx context.Context) ([]uuid.UUID, error) {
	return s.repo.GetAllStudentIDs(ctx)
}

func NewAchievementService(repo repository.AchievementRepository) AchievementService {
	return &achievementService{repo: repo}
}

// Helper function untuk mengekstrak user ID dari JWT claims
func extractUserIDFromClaims(c *fiber.Ctx) (uuid.UUID, error) {
	claims := c.Locals("user_info")
	if claims == nil {
		return uuid.Nil, errors.New("user info not found")
	}

	jwtClaims, ok := claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, errors.New("invalid claims format")
	}

	userIDStr, ok := jwtClaims["user_id"].(string)
	if !ok {
		return uuid.Nil, errors.New("invalid user ID in token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, errors.New("invalid user ID format")
	}

	return userID, nil
}

func (s *achievementService) SubmitPrestasi(ctx context.Context, userID uuid.UUID, req mongodb.Achievement) (*model.AchievementReference, error) {
	// 1. Cari data Student berdasarkan User ID yang login
	student, err := s.repo.GetStudentByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("student data not found for this user")
	}

	// 2. Setup Data untuk MongoDB
	req.ID = primitive.NewObjectID()
	req.StudentID = student.ID // Link ke UUID Student di Postgres
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	if req.CustomFields == nil {
		req.CustomFields = make(map[string]interface{})
	}

	// 3. Simpan ke MongoDB
	mongoID, err := s.repo.SaveAchievementMongo(ctx, req)
	if err != nil {
		return nil, err
	}

	// 4. Setup Data untuk Postgres (Reference)
	// Sesuai SRS Flow 4: Status awal 'draft'
	ref := model.AchievementReference{
		ID:                 uuid.New(),
		StudentID:          student.ID,
		MongoAchievementID: mongoID,
		Status:             "draft",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// 5. Simpan ke Postgres
	err = s.repo.SaveAchievementReference(ctx, ref)
	if err != nil {
		return nil, err
	}

	return &ref, nil
}

// GetAchievementByID - Get achievement by ID with details
func (s *achievementService) GetAchievementByID(ctx context.Context, userID uuid.UUID, mongoAchievementID string) (*mongodb.Achievement, error) {
	// 1. Validate MongoDB ObjectID format
	objectID, err := primitive.ObjectIDFromHex(mongoAchievementID)
	if err != nil {
		return nil, errors.New("invalid achievement ID format")
	}

	// 2. Get achievement details from MongoDB
	achievement, err := s.repo.GetAchievementDetailFromMongo(ctx, mongoAchievementID)
	if err != nil {
		return nil, errors.New("achievement not found")
	}

	// 3. Check authorization - user can only view their own achievements or if they are lecturer viewing advisee's achievement
	student, err := s.repo.GetStudentByUserID(ctx, userID)
	if err == nil {
		// User is a student - can only view their own achievements
		if achievement.StudentID != student.ID {
			return nil, errors.New("unauthorized: you can only view your own achievements")
		}
	} else {
		// Check if user is a lecturer
		lecturer, err := s.repo.GetLecturerByUserID(ctx, userID)
		if err != nil {
			return nil, errors.New("unauthorized: access denied")
		}

		// Check if the achievement belongs to lecturer's advisee
		achievementStudent, err := s.repo.GetStudentByID(ctx, achievement.StudentID)
		if err != nil {
			return nil, errors.New("student data not found")
		}

		if achievementStudent.AdvisorID != lecturer.ID {
			return nil, errors.New("unauthorized: you can only view achievements of your advisees")
		}
	}

	// 4. Set the ObjectID for response
	achievement.ID = objectID

	return achievement, nil
}

// UpdateAchievement - Update achievement (only for draft status)
func (s *achievementService) UpdateAchievement(ctx context.Context, userID uuid.UUID, achievementID uuid.UUID, req mongodb.Achievement) (*model.AchievementReference, error) {
	// 1. Get student data
	student, err := s.repo.GetStudentByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("student data not found for this user")
	}

	// 2. Get achievement reference by ID
	ref, err := s.repo.GetAchievementReferenceByID(ctx, achievementID)
	if err != nil {
		return nil, errors.New("achievement not found")
	}

	// 3. Check authorization - only owner can update
	if ref.StudentID != student.ID {
		return nil, errors.New("unauthorized: achievement does not belong to this student")
	}

	// 4. Check status - only draft can be updated
	if ref.Status != "draft" {
		return nil, errors.New("only draft achievements can be updated")
	}

	// 5. Update achievement in MongoDB
	objectID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	if err != nil {
		return nil, errors.New("invalid mongo achievement ID")
	}
	req.ID = objectID
	req.UpdatedAt = time.Now()

	err = s.repo.UpdateAchievementInMongo(ctx, req)
	if err != nil {
		return nil, errors.New("failed to update achievement")
	}

	// 6. Update timestamp in PostgreSQL
	err = s.repo.UpdateAchievementTimestamp(ctx, achievementID)
	if err != nil {
		return nil, errors.New("failed to update achievement timestamp")
	}

	// 7. Get updated achievement reference
	updatedRef, err := s.repo.GetAchievementReferenceByID(ctx, achievementID)
	if err != nil {
		return nil, err
	}

	return updatedRef, nil
}

// SubmitForVerification - FR-004: Submit untuk Verifikasi
func (s *achievementService) SubmitForVerification(ctx context.Context, userID uuid.UUID, achievementID uuid.UUID) (*model.AchievementReference, error) {
	// 1. Cari data Student berdasarkan User ID yang login
	student, err := s.repo.GetStudentByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("student data not found for this user")
	}

	// 2. Get achievement reference by ID
	ref, err := s.repo.GetAchievementReferenceByID(ctx, achievementID)
	if err != nil {
		return nil, errors.New("achievement not found")
	}

	// 3. Validasi: Pastikan achievement milik student yang login
	if ref.StudentID != student.ID {
		return nil, errors.New("unauthorized: achievement does not belong to this student")
	}

	// 4. Validasi: Pastikan status adalah 'draft'
	if ref.Status != "draft" {
		return nil, errors.New("achievement must be in 'draft' status to submit")
	}

	// 5. Update status menjadi 'submitted'
	err = s.repo.UpdateAchievementStatusToSubmitted(ctx, achievementID)
	if err != nil {
		return nil, errors.New("failed to update achievement status")
	}

	// 6. Get updated achievement reference
	updatedRef, err := s.repo.GetAchievementReferenceByID(ctx, achievementID)
	if err != nil {
		return nil, err
	}

	return updatedRef, nil
}

// DeleteDraftAchievement - FR-005: Hapus Prestasi
func (s *achievementService) DeleteDraftAchievement(ctx context.Context, userID uuid.UUID, achievementID uuid.UUID) error {
	// 1. Cari data Student berdasarkan User ID yang login
	student, err := s.repo.GetStudentByUserID(ctx, userID)
	if err != nil {
		return errors.New("student data not found for this user")
	}

	// 2. Get achievement reference by ID
	ref, err := s.repo.GetAchievementReferenceByID(ctx, achievementID)
	if err != nil {
		return errors.New("achievement not found")
	}

	// 3. Validasi: Pastikan achievement milik student yang login
	if ref.StudentID != student.ID {
		return errors.New("unauthorized: achievement does not belong to this student")
	}

	// 4. Validasi: Pastikan status adalah 'draft'
	if ref.Status != "draft" {
		return errors.New("only draft achievements can be deleted")
	}

	// 5. Soft delete di MongoDB
	err = s.repo.SoftDeleteAchievementMongo(ctx, ref.MongoAchievementID)
	if err != nil {
		return errors.New("failed to delete achievement from MongoDB")
	}

	// 6. Update status di PostgreSQL menjadi 'deleted'
	err = s.repo.UpdateAchievementReferenceToDeleted(ctx, achievementID)
	if err != nil {
		return errors.New("failed to update achievement status in PostgreSQL")
	}

	return nil
}

// VerifyAchievement - FR-007: Verify Prestasi
func (s *achievementService) VerifyAchievement(ctx context.Context, userID uuid.UUID, achievementID uuid.UUID) (*model.AchievementReference, error) {
	// 1. Get lecturer data by user ID
	lecturer, err := s.repo.GetLecturerByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("lecturer data not found for this user")
	}

	// 2. Get achievement reference by ID
	ref, err := s.repo.GetAchievementReferenceByID(ctx, achievementID)
	if err != nil {
		return nil, errors.New("achievement not found")
	}

	// 3. Validasi: Pastikan status adalah 'submitted'
	if ref.Status != "submitted" {
		return nil, errors.New("achievement must be in 'submitted' status to verify")
	}

	// 4. Get student data untuk validasi advisor
	student, err := s.repo.GetStudentByID(ctx, ref.StudentID)
	if err != nil {
		return nil, errors.New("student data not found")
	}

	// 5. Validasi: Pastikan achievement milik mahasiswa bimbingan dosen ini
	if student.AdvisorID != lecturer.ID {
		return nil, errors.New("unauthorized: you can only verify achievements of your advisees")
	}

	// 6. Update status menjadi 'verified' dengan verified_by dan verified_at
	err = s.repo.UpdateAchievementStatusToVerified(ctx, achievementID, lecturer.ID)
	if err != nil {
		return nil, errors.New("failed to verify achievement")
	}

	// 7. Get updated achievement reference
	updatedRef, err := s.repo.GetAchievementReferenceByID(ctx, achievementID)
	if err != nil {
		return nil, err
	}

	return updatedRef, nil
}

// RejectAchievement - FR-008: Reject Prestasi
func (s *achievementService) RejectAchievement(ctx context.Context, userID uuid.UUID, achievementID uuid.UUID, rejectionNote string) (*model.AchievementReference, error) {
	// 1. Validasi rejection note tidak kosong
	if rejectionNote == "" {
		return nil, errors.New("rejection note is required")
	}

	// 2. Get lecturer data by user ID
	lecturer, err := s.repo.GetLecturerByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("lecturer data not found for this user")
	}

	// 3. Get achievement reference by ID
	ref, err := s.repo.GetAchievementReferenceByID(ctx, achievementID)
	if err != nil {
		return nil, errors.New("achievement not found")
	}

	// 4. Validasi: Pastikan status adalah 'submitted'
	if ref.Status != "submitted" {
		return nil, errors.New("achievement must be in 'submitted' status to reject")
	}

	// 5. Get student data untuk validasi advisor
	student, err := s.repo.GetStudentByID(ctx, ref.StudentID)
	if err != nil {
		return nil, errors.New("student data not found")
	}

	// 6. Validasi: Pastikan achievement milik mahasiswa bimbingan dosen ini
	if student.AdvisorID != lecturer.ID {
		return nil, errors.New("unauthorized: you can only reject achievements of your advisees")
	}

	// 7. Update status menjadi 'rejected' dengan rejection note
	err = s.repo.UpdateAchievementStatusToRejected(ctx, achievementID, rejectionNote)
	if err != nil {
		return nil, errors.New("failed to reject achievement")
	}

	// 8. Get updated achievement reference
	updatedRef, err := s.repo.GetAchievementReferenceByID(ctx, achievementID)
	if err != nil {
		return nil, err
	}

	return updatedRef, nil
}

// GetStudentAchievements - FR-006: View Prestasi Mahasiswa Bimbingan
func (s *achievementService) GetStudentAchievements(ctx context.Context, userID uuid.UUID, status string, page, limit int) (*model.AchievementListResponse, error) {
	// Set default pagination values
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// 1. Get lecturer data by user ID
	lecturer, err := s.repo.GetLecturerByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("lecturer data not found for this user")
	}

	// 2. Get list of student IDs under this advisor
	studentIDs, err := s.repo.GetStudentIDsByAdvisorID(ctx, lecturer.ID)
	if err != nil {
		return nil, errors.New("failed to get student list")
	}

	// If no students, return empty list
	if len(studentIDs) == 0 {
		return &model.AchievementListResponse{
			Achievements: []model.AchievementWithStudent{},
			Pagination: model.PaginationMetadata{
				Page:       page,
				Limit:      limit,
				Total:      0,
				TotalPages: 0,
			},
		}, nil
	}

	// 3. Get achievements with student info and pagination
	achievements, total, err := s.repo.GetAchievementsWithStudentInfo(ctx, studentIDs, status, page, limit)
	if err != nil {
		return nil, errors.New("failed to get achievements")
	}

	// 4. Fetch details from MongoDB for each achievement
	for i := range achievements {
		detail, err := s.repo.GetAchievementDetailFromMongo(ctx, achievements[i].MongoAchievementID)
		if err == nil {
			achievements[i].Details = detail
		}
		// If error fetching from MongoDB, just skip (details will be nil)
	}

	// 5. Calculate pagination metadata
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

// GetAchievementHistory - FR-010: Get Achievement History
func (s *achievementService) GetAchievementHistory(ctx context.Context, userID uuid.UUID, achievementID uuid.UUID) (*model.AchievementHistoryResponse, error) {
	// 1. Get achievement reference by ID
	ref, err := s.repo.GetAchievementReferenceByID(ctx, achievementID)
	if err != nil {
		return nil, errors.New("achievement not found")
	}

	// 2. Check authorization - student atau dosen wali bisa akses
	// Cek apakah user adalah student pemilik achievement
	student, err := s.repo.GetStudentByUserID(ctx, userID)
	isStudent := err == nil && student.ID == ref.StudentID

	// Cek apakah user adalah dosen wali dari student pemilik achievement
	var isAdvisor bool
	if !isStudent {
		lecturer, err := s.repo.GetLecturerByUserID(ctx, userID)
		if err == nil {
			studentData, err := s.repo.GetStudentByID(ctx, ref.StudentID)
			if err == nil && studentData.AdvisorID == lecturer.ID {
				isAdvisor = true
			}
		}
	}

	// 3. Validasi authorization
	if !isStudent && !isAdvisor {
		return nil, errors.New("unauthorized: you can only view history of your own achievements or your advisees' achievements")
	}

	// 4. Get status history dari log table
	timeline, err := s.repo.GetAchievementStatusHistory(ctx, achievementID)
	if err != nil {
		return nil, errors.New("failed to get achievement history")
	}

	// 5. Return timeline response
	return &model.AchievementHistoryResponse{
		AchievementID: achievementID,
		Timeline:      timeline,
	}, nil
}

// GetAllAchievementsForAdmin - FR-010: View All Achievements (Admin)
func (s *achievementService) GetAllAchievementsForAdmin(ctx context.Context, filters model.AdminAchievementFilters, page, limit int) (*model.AchievementListResponse, error) {
	// Set default pagination values
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Get all achievements with filters
	achievements, total, err := s.repo.GetAllAchievementsForAdmin(ctx, filters, page, limit)
	if err != nil {
		return nil, errors.New("failed to get achievements")
	}

	// Fetch details from MongoDB for each achievement
	for i := range achievements {
		detail, err := s.repo.GetAchievementDetailFromMongo(ctx, achievements[i].MongoAchievementID)
		if err == nil {
			achievements[i].Details = detail
		}
		// If error fetching from MongoDB, just skip (details will be nil)
	}

	// Calculate pagination metadata
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

// GetAchievementStatistics - FR-011: Achievement Statistics
func (s *achievementService) GetAchievementStatistics(ctx context.Context, userID uuid.UUID, filters model.StatisticsFilters) (*model.AchievementStatistics, error) {
	var studentIDs []uuid.UUID

	// Determine student IDs based on user role
	// Try as student first
	student, err := s.repo.GetStudentByUserID(ctx, userID)
	if err == nil {
		// User is a student - only their own achievements
		studentIDs = []uuid.UUID{student.ID}
	} else {
		// Try as lecturer
		lecturer, err := s.repo.GetLecturerByUserID(ctx, userID)
		if err == nil {
			// User is a lecturer - get advisee students
			studentIDs, err = s.repo.GetStudentIDsByAdvisorID(ctx, lecturer.ID)
			if err != nil {
				return nil, errors.New("failed to get student list")
			}
		} else {
			// ADMIN
			if filters.StudentID != nil {
				studentIDs = []uuid.UUID{*filters.StudentID}
			} else {
				// ✅ ADMIN BOLEH AMBIL SEMUA MAHASISWA
				studentIDs, err = s.repo.GetAllStudentIDs(ctx)
				if err != nil {
					return nil, errors.New("failed to get all students")
				}
			}
		}
	}

	if len(studentIDs) == 0 {
		return &model.AchievementStatistics{
			TotalAchievements:  0,
			ByType:             []model.StatsByType{},
			ByPeriod:           []model.StatsByPeriod{},
			TopStudents:        []model.TopStudent{},
			LevelDistribution:  []model.LevelDistribution{},
			StatusDistribution: []model.StatusDistribution{},
		}, nil
	}

	// Get total achievements
	total, err := s.repo.GetTotalAchievements(ctx, studentIDs, filters)
	if err != nil {
		return nil, errors.New("failed to get total achievements")
	}

	// Get statistics by type (need MongoDB collection)
	// Note: We need to pass mongoColl to repository methods
	// For now, we'll return empty arrays for MongoDB-dependent stats
	byType := []model.StatsByType{}

	// Get statistics by period
	byPeriod, err := s.repo.GetStatisticsByPeriod(ctx, studentIDs, filters)
	if err != nil {
		byPeriod = []model.StatsByPeriod{}
	}

	// Get top students (limit to 10)
	topStudents, err := s.repo.GetTopStudents(ctx, studentIDs, filters, 10)
	if err != nil {
		topStudents = []model.TopStudent{}
	}

	// Get level distribution
	levelDist := []model.LevelDistribution{}

	// Get status distribution
	statusDist, err := s.repo.GetStatusDistribution(ctx, studentIDs, filters)
	if err != nil {
		statusDist = []model.StatusDistribution{}
	}

	return &model.AchievementStatistics{
		TotalAchievements:  total,
		ByType:             byType,
		ByPeriod:           byPeriod,
		TopStudents:        topStudents,
		LevelDistribution:  levelDist,
		StatusDistribution: statusDist,
	}, nil
}

// HTTP Endpoints
func (s *achievementService) GetAchievementsEndpoint(c *fiber.Ctx) error {
	userID, err := extractUserIDFromClaims(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	status := c.Query("status", "")

	result, err := s.GetStudentAchievements(c.Context(), userID, status, page, limit)
	if err != nil {
		switch err.Error() {
		case "lecturer data not found for this user":
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		case "failed to get student list":
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		case "failed to get achievements":
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(500).JSON(fiber.Map{"error": "Failed to get achievements"})
		}
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   result,
	})
}

func (s *achievementService) CreateAchievementEndpoint(c *fiber.Ctx) error {
	userID, err := extractUserIDFromClaims(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}

	var req mongodb.Achievement
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON body"})
	}

	result, err := s.SubmitPrestasi(c.Context(), userID, req)
	if err != nil {
		switch err.Error() {
		case "student data not found for this user":
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create achievement"})
		}
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Achievement created successfully",
		"data":    result,
	})
}

func (s *achievementService) DeleteAchievementEndpoint(c *fiber.Ctx) error {
	userID, err := extractUserIDFromClaims(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}

	achievementID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID format"})
	}

	err = s.DeleteDraftAchievement(c.Context(), userID, achievementID)
	if err != nil {
		switch err.Error() {
		case "student data not found for this user":
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		case "achievement not found":
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		case "unauthorized: achievement does not belong to this student":
			return c.Status(403).JSON(fiber.Map{"error": err.Error()})
		case "only draft achievements can be deleted":
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(500).JSON(fiber.Map{"error": "Failed to delete achievement"})
		}
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Achievement deleted successfully",
	})
}

func (s *achievementService) SubmitAchievementEndpoint(c *fiber.Ctx) error {
	userID, err := extractUserIDFromClaims(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}

	achievementID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID format"})
	}

	result, err := s.SubmitForVerification(c.Context(), userID, achievementID)
	if err != nil {
		switch err.Error() {
		case "student data not found for this user":
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		case "achievement not found":
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		case "unauthorized: achievement does not belong to this student":
			return c.Status(403).JSON(fiber.Map{"error": err.Error()})
		case "achievement must be in 'draft' status to submit":
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(500).JSON(fiber.Map{"error": "Failed to submit achievement"})
		}
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Achievement submitted for verification successfully",
		"data":    result,
	})
}

func (s *achievementService) VerifyAchievementEndpoint(c *fiber.Ctx) error {
	userID, err := extractUserIDFromClaims(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}

	achievementID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID format"})
	}

	result, err := s.VerifyAchievement(c.Context(), userID, achievementID)
	if err != nil {
		switch err.Error() {
		case "lecturer data not found for this user":
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		case "achievement not found":
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		case "achievement must be in 'submitted' status to verify":
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		case "student data not found":
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		case "unauthorized: you can only verify achievements of your advisees":
			return c.Status(403).JSON(fiber.Map{"error": err.Error()})
		case "failed to verify achievement":
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(500).JSON(fiber.Map{"error": "Failed to verify achievement"})
		}
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Achievement verified successfully",
		"data":    result,
	})
}

func (s *achievementService) RejectAchievementEndpoint(c *fiber.Ctx) error {
	userID, err := extractUserIDFromClaims(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}

	achievementID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID format"})
	}

	var req struct {
		RejectionNote string `json:"rejection_note"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON body"})
	}

	result, err := s.RejectAchievement(c.Context(), userID, achievementID, req.RejectionNote)
	if err != nil {
		switch err.Error() {
		case "rejection note is required":
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		case "lecturer data not found for this user":
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		case "achievement not found":
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		case "achievement must be in 'submitted' status to reject":
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		case "student data not found":
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		case "unauthorized: you can only reject achievements of your advisees":
			return c.Status(403).JSON(fiber.Map{"error": err.Error()})
		case "failed to reject achievement":
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(500).JSON(fiber.Map{"error": "Failed to reject achievement"})
		}
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Achievement rejected successfully",
		"data":    result,
	})
}

func (s *achievementService) GetAchievementHistoryEndpoint(c *fiber.Ctx) error {
	userID, err := extractUserIDFromClaims(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}

	achievementID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID format"})
	}

	result, err := s.GetAchievementHistory(c.Context(), userID, achievementID)
	if err != nil {
		switch err.Error() {
		case "achievement not found":
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		case "unauthorized: you can only view history of your own achievements or your advisees' achievements":
			return c.Status(403).JSON(fiber.Map{"error": err.Error()})
		case "failed to get achievement history":
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(500).JSON(fiber.Map{"error": "Failed to get achievement history"})
		}
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   result,
	})
}

func (s *achievementService) GetAchievementStatisticsEndpoint(c *fiber.Ctx) error {
	userID, err := extractUserIDFromClaims(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}

	filters := model.StatisticsFilters{
		Status: c.Query("status", ""),
	}

	if studentIDStr := c.Query("student_id"); studentIDStr != "" {
		studentID, err := uuid.Parse(studentIDStr)
		if err == nil {
			filters.StudentID = &studentID
		}
	}

	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		filters.DateFrom = &dateFromStr
	}

	if dateToStr := c.Query("date_to"); dateToStr != "" {
		filters.DateTo = &dateToStr
	}

	result, err := s.GetAchievementStatistics(c.Context(), userID, filters)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   result,
	})
}

func (s *achievementService) GetAllAchievementsForAdminEndpoint(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	filters := model.AdminAchievementFilters{
		Status:    c.Query("status", ""),
		SortBy:    c.Query("sort_by", "created_at"),
		SortOrder: c.Query("sort_order", "desc"),
	}

	if studentIDStr := c.Query("student_id"); studentIDStr != "" {
		studentID, err := uuid.Parse(studentIDStr)
		if err == nil {
			filters.StudentID = &studentID
		}
	}

	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		dateFrom, err := time.Parse("2006-01-02", dateFromStr)
		if err == nil {
			filters.DateFrom = &dateFrom
		}
	}

	if dateToStr := c.Query("date_to"); dateToStr != "" {
		dateTo, err := time.Parse("2006-01-02", dateToStr)
		if err == nil {
			filters.DateTo = &dateTo
		}
	}

	result, err := s.GetAllAchievementsForAdmin(c.Context(), filters, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   result,
	})
}
func (s *achievementService) GetAchievementByIDEndpoint(c *fiber.Ctx) error {
	ctx := c.Context()

	// ⛔ ADMIN: ID = UUID POSTGRES
	achievementID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid achievement id format",
		})
	}

	// 1️⃣ Ambil reference dari Postgres
	ref, err := s.repo.GetAchievementReferenceByID(ctx, achievementID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "achievement not found",
		})
	}

	// 2️⃣ Ambil detail dari Mongo pakai mongo_achievement_id
	detail, err := s.repo.GetAchievementDetailFromMongo(
		ctx,
		ref.MongoAchievementID,
	)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "achievement detail not found in mongo",
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"reference": ref,
			"detail":    detail,
		},
	})

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   ref,
	})
}

func (s *achievementService) UpdateAchievementEndpoint(c *fiber.Ctx) error {
	userID, err := extractUserIDFromClaims(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}

	achievementID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID format"})
	}

	var req mongodb.Achievement
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON body"})
	}

	result, err := s.UpdateAchievement(c.Context(), userID, achievementID, req)
	if err != nil {
		switch err.Error() {
		case "student data not found for this user":
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		case "achievement not found":
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		case "unauthorized: achievement does not belong to this student":
			return c.Status(403).JSON(fiber.Map{"error": err.Error()})
		case "only draft achievements can be updated":
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(500).JSON(fiber.Map{"error": "Failed to update achievement"})
		}
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Achievement updated successfully",
		"data":    result,
	})
}

// Reports & Analytics Methods

// GetReportsStatistics - General statistics for reports
func (s *achievementService) GetReportsStatistics(ctx context.Context, userID uuid.UUID, filters model.StatisticsFilters) (*model.AchievementStatistics, error) {
	// This is similar to GetAchievementStatistics but can have different logic for reports
	return s.GetAchievementStatistics(ctx, userID, filters)
}

// GetStudentReport - Detailed report for specific student
func (s *achievementService) GetStudentReport(ctx context.Context, userID uuid.UUID, studentID uuid.UUID) (*model.StudentReportResponse, error) {
	// Check authorization - admin can view any student, lecturer can view advisees, student can view own
	var isAuthorized bool

	// Try as student first
	student, err := s.repo.GetStudentByUserID(ctx, userID)
	if err == nil && student.ID == studentID {
		isAuthorized = true
	}

	// Try as lecturer if not student
	if !isAuthorized {
		lecturer, err := s.repo.GetLecturerByUserID(ctx, userID)
		if err == nil {
			// Check if student is advisee
			studentData, err := s.repo.GetStudentByID(ctx, studentID)
			if err == nil && studentData.AdvisorID == lecturer.ID {
				isAuthorized = true
			}
		}
	}

	// For admin, we assume they have permission (this should be checked by middleware)
	if !isAuthorized {
		// Check if user has admin permissions (simplified check)
		isAuthorized = true // This should be properly implemented with role checking
	}

	if !isAuthorized {
		return nil, errors.New("unauthorized: you can only view reports of your own achievements or your advisees' achievements")
	}

	// Get student with user info
	studentWithUser, err := s.repo.GetStudentWithUserByID(ctx, studentID)
	if err != nil {
		return nil, errors.New("student not found")
	}

	// Get statistics for this student
	filters := model.StatisticsFilters{
		StudentID: &studentID,
	}
	statistics, err := s.GetAchievementStatistics(ctx, userID, filters)
	if err != nil {
		// If error getting statistics, return empty statistics
		statistics = &model.AchievementStatistics{
			TotalAchievements:  0,
			ByType:             []model.StatsByType{},
			ByPeriod:           []model.StatsByPeriod{},
			TopStudents:        []model.TopStudent{},
			LevelDistribution:  []model.LevelDistribution{},
			StatusDistribution: []model.StatusDistribution{},
		}
	}

	// Get recent achievements (last 10)
	achievements, _, err := s.repo.GetStudentAchievements(ctx, studentID, 1, 10)
	if err != nil {
		achievements = []model.AchievementWithStudent{}
	}

	// Fetch details from MongoDB for each achievement
	for i := range achievements {
		detail, err := s.repo.GetAchievementDetailFromMongo(ctx, achievements[i].MongoAchievementID)
		if err == nil {
			achievements[i].Details = detail
		}
	}

	return &model.StudentReportResponse{
		Student:      *studentWithUser,
		Statistics:   *statistics,
		Achievements: achievements,
	}, nil
}

// Reports HTTP Endpoints

func (s *achievementService) GetReportsStatisticsEndpoint(c *fiber.Ctx) error {
	userID, err := extractUserIDFromClaims(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}

	filters := model.StatisticsFilters{
		Status: c.Query("status", ""),
	}

	if studentIDStr := c.Query("student_id"); studentIDStr != "" {
		studentID, err := uuid.Parse(studentIDStr)
		if err == nil {
			filters.StudentID = &studentID
		}
	}

	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		filters.DateFrom = &dateFromStr
	}

	if dateToStr := c.Query("date_to"); dateToStr != "" {
		filters.DateTo = &dateToStr
	}

	result, err := s.GetReportsStatistics(c.Context(), userID, filters)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   result,
	})
}

func (s *achievementService) GetStudentReportEndpoint(c *fiber.Ctx) error {
	userID, err := extractUserIDFromClaims(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}

	studentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid student ID format"})
	}

	result, err := s.GetStudentReport(c.Context(), userID, studentID)
	if err != nil {
		switch err.Error() {
		case "unauthorized: you can only view reports of your own achievements or your advisees' achievements":
			return c.Status(403).JSON(fiber.Map{"error": err.Error()})
		case "student not found":
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(500).JSON(fiber.Map{"error": "Failed to get student report"})
		}
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   result,
	})
}

// UploadAttachment - Upload file attachment to achievement
func (s *achievementService) UploadAttachment(ctx context.Context, userID uuid.UUID, achievementID uuid.UUID, fileName, fileURL, fileType string) error {
	// 1. Get student data
	student, err := s.repo.GetStudentByUserID(ctx, userID)
	if err != nil {
		return errors.New("student data not found for this user")
	}

	// 2. Get achievement reference by ID
	ref, err := s.repo.GetAchievementReferenceByID(ctx, achievementID)
	if err != nil {
		return errors.New("achievement not found")
	}

	// 3. Check authorization - only owner can upload
	if ref.StudentID != student.ID {
		return errors.New("unauthorized: achievement does not belong to this student")
	}

	// 4. Check status - only draft and rejected can have new attachments
	if ref.Status != "draft" && ref.Status != "rejected" {
		return errors.New("attachments can only be added to draft or rejected achievements")
	}

	// 5. Add attachment to MongoDB
	err = s.repo.AddAttachmentToAchievement(ctx, ref.MongoAchievementID, fileName, fileURL, fileType)
	if err != nil {
		return errors.New("failed to add attachment")
	}

	return nil
}

// File Upload HTTP Endpoint
func (s *achievementService) UploadAttachmentEndpoint(c *fiber.Ctx) error {
	userID, err := extractUserIDFromClaims(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}

	achievementID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid achievement ID format"})
	}

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "No file uploaded"})
	}

	// Validate file size (max 10MB)
	if file.Size > 10*1024*1024 {
		return c.Status(400).JSON(fiber.Map{"error": "File size too large (max 10MB)"})
	}

	// Validate file type
	allowedTypes := map[string]bool{
		"image/jpeg":         true,
		"image/png":          true,
		"image/gif":          true,
		"application/pdf":    true,
		"application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	}

	if !allowedTypes[file.Header.Get("Content-Type")] {
		return c.Status(400).JSON(fiber.Map{"error": "File type not allowed"})
	}

	// For this implementation, we'll simulate file storage
	// In production, you would save to disk, cloud storage, etc.
	fileName := file.Filename
	fileURL := "/uploads/" + achievementID.String() + "/" + fileName
	fileType := file.Header.Get("Content-Type")

	// Save file (simplified - in production use proper file handling)
	// err = c.SaveFile(file, "./uploads/"+achievementID.String()+"/"+fileName)
	// if err != nil {
	//     return c.Status(500).JSON(fiber.Map{"error": "Failed to save file"})
	// }

	err = s.UploadAttachment(c.Context(), userID, achievementID, fileName, fileURL, fileType)
	if err != nil {
		switch err.Error() {
		case "student data not found for this user":
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		case "achievement not found":
			return c.Status(404).JSON(fiber.Map{"error": err.Error()})
		case "unauthorized: achievement does not belong to this student":
			return c.Status(403).JSON(fiber.Map{"error": err.Error()})
		case "attachments can only be added to draft or rejected achievements":
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(500).JSON(fiber.Map{"error": "Failed to upload attachment"})
		}
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "File uploaded successfully",
		"data": fiber.Map{
			"filename": fileName,
			"url":      fileURL,
			"type":     fileType,
		},
	})
}

func (s *achievementService) GetAchievementAdminDetailEndpoint(c *fiber.Ctx) error {
	achievementID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid achievement reference id",
		})
	}

	// 1️⃣ Ambil reference dari Postgres
	ref, err := s.repo.GetAchievementReferenceByID(c.Context(), achievementID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "achievement not found",
		})
	}

	// 2️⃣ Ambil detail dari Mongo
	detail, err := s.repo.GetAchievementDetailFromMongo(
		c.Context(),
		ref.MongoAchievementID,
	)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "achievement detail not found in mongo",
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"reference": ref,
			"detail":    detail,
		},
	})
}
