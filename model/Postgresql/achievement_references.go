package model

import (
	"time"
	mongodb "UASBE/model/MongoDB"

	"github.com/google/uuid"
)

type AchievementReference struct {
	ID                 uuid.UUID  `json:"id"`
	StudentID          uuid.UUID  `json:"student_id"`
	MongoAchievementID string     `json:"mongo_achievement_id"`
	Status             string     `json:"status"`
	SubmittedAt        *time.Time `json:"submitted_at"`
	VerifiedAt         *time.Time `json:"verified_at"`
	VerifiedBy         *uuid.UUID `json:"verified_by"`
	RejectionNote      *string    `json:"rejection_note"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type AchievementWithStudent struct {
	ID                 uuid.UUID            `json:"id"`
	StudentID          uuid.UUID            `json:"student_id"`
	StudentNIM         string               `json:"student_nim"`
	StudentName        string               `json:"student_name"`
	ProgramStudy       string               `json:"program_study"`
	MongoAchievementID string               `json:"mongo_achievement_id"`
	Status             string               `json:"status"`
	SubmittedAt        *time.Time           `json:"submitted_at"`
	VerifiedAt         *time.Time           `json:"verified_at"`
	VerifiedBy         *uuid.UUID           `json:"verified_by"`
	RejectionNote      *string              `json:"rejection_note"`
	CreatedAt          time.Time            `json:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at"`
	Details            *mongodb.Achievement `json:"details,omitempty"`
}

// PaginationMetadata contains pagination information
type PaginationMetadata struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// AchievementListResponse is the response structure for achievement list
type AchievementListResponse struct {
	Achievements []AchievementWithStudent `json:"achievements"`
	Pagination   PaginationMetadata       `json:"pagination"`
}

// AdminAchievementFilters for filtering achievements in admin view
type AdminAchievementFilters struct {
	Status    string     `json:"status"`
	StudentID *uuid.UUID `json:"student_id"`
	DateFrom  *time.Time `json:"date_from"`
	DateTo    *time.Time `json:"date_to"`
	SortBy    string     `json:"sort_by"`    // created_at, updated_at, status
	SortOrder string     `json:"sort_order"` // asc, desc
}

// AchievementStatistics represents overall achievement statistics
type AchievementStatistics struct {
	TotalAchievements  int                  `json:"total_achievements"`
	ByType             []StatsByType        `json:"by_type"`
	ByPeriod           []StatsByPeriod      `json:"by_period"`
	TopStudents        []TopStudent         `json:"top_students"`
	LevelDistribution  []LevelDistribution  `json:"level_distribution"`
	StatusDistribution []StatusDistribution `json:"status_distribution"`
}

// StatsByType represents statistics grouped by achievement type/category
type StatsByType struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// StatsByPeriod represents statistics grouped by time period
type StatsByPeriod struct {
	Period string `json:"period"` // Format: YYYY-MM
	Count  int    `json:"count"`
}

// TopStudent represents top students by achievement count
type TopStudent struct {
	StudentID    uuid.UUID `json:"student_id"`
	StudentNIM   string    `json:"student_nim"`
	StudentName  string    `json:"student_name"`
	ProgramStudy string    `json:"program_study"`
	Count        int       `json:"count"`
}

// LevelDistribution represents distribution by competition level
type LevelDistribution struct {
	Level string `json:"level"` // international, national, regional, local
	Count int    `json:"count"`
}

// StatusDistribution represents distribution by achievement status
type StatusDistribution struct {
	Status string `json:"status"` // draft, submitted, verified, rejected
	Count  int    `json:"count"`
}

// StatisticsFilters for filtering statistics queries
type StatisticsFilters struct {
	DateFrom  *string    `json:"date_from"`  // Format: YYYY-MM-DD
	DateTo    *string    `json:"date_to"`    // Format: YYYY-MM-DD
	Status    string     `json:"status"`     // Filter by status
	StudentID *uuid.UUID `json:"student_id"` // Filter by specific student (admin only)
}

// AchievementStatusLog represents a log entry for achievement status changes
type AchievementStatusLog struct {
	ID            uuid.UUID  `json:"id"`
	AchievementID uuid.UUID  `json:"achievement_id"`
	Status        string     `json:"status"`
	ChangedBy     *uuid.UUID `json:"changed_by"`
	ChangedByName *string    `json:"changed_by_name"`
	RejectionNote *string    `json:"rejection_note"`
	CreatedAt     time.Time  `json:"created_at"`
}

// AchievementHistoryResponse represents the timeline response
type AchievementHistoryResponse struct {
	AchievementID uuid.UUID              `json:"achievement_id"`
	Timeline      []AchievementStatusLog `json:"timeline"`
}

// StudentReportResponse represents detailed student report
type StudentReportResponse struct {
	Student      StudentWithUser          `json:"student"`
	Statistics   AchievementStatistics    `json:"statistics"`
	Achievements []AchievementWithStudent `json:"recent_achievements"`
}
