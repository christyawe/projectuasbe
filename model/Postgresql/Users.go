package model

import (
	"time"

	"github.com/google/uuid"
)

// Users adalah model untuk tabel users
type Users struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"`
	RoleID       uuid.UUID `json:"role_id"`
	ISActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateUserRequest untuk membuat user baru
type CreateUserRequest struct {
	Username    string       `json:"username" validate:"required"`
	Email       string       `json:"email" validate:"required,email"`
	Password    string       `json:"password" validate:"required,min=6"`
	FullName    string       `json:"full_name" validate:"required"`
	RoleID      uuid.UUID    `json:"role_id" validate:"required"`
	IsActive    bool         `json:"is_active"`
	ProfileType string       `json:"profile_type,omitempty"` // "student" or "lecturer"
	ProfileData *ProfileData `json:"profile_data,omitempty"`
}

// UpdateUserRequest untuk update user
type UpdateUserRequest struct {
	Email    string `json:"email,omitempty" validate:"omitempty,email"`
	FullName string `json:"full_name,omitempty"`
	IsActive *bool  `json:"is_active,omitempty"`
}

// UpdateRoleRequest untuk update role user
type UpdateRoleRequest struct {
	RoleID uuid.UUID `json:"role_id" validate:"required"`
}

// UpdateUserRoleRequest untuk update role user (alias)
type UpdateUserRoleRequest struct {
	RoleID uuid.UUID `json:"role_id" validate:"required"`
}

// UpdateAdvisorRequest untuk update advisor mahasiswa
type UpdateAdvisorRequest struct {
	AdvisorID uuid.UUID `json:"advisor_id" validate:"required"`
}

// ProfileData untuk student atau lecturer profile
type ProfileData struct {
	// Student fields
	StudentID    string    `json:"student_id,omitempty"`
	ProgramStudy string    `json:"program_study,omitempty"`
	AcademicYear string    `json:"academic_year,omitempty"`
	AdvisorID    uuid.UUID `json:"advisor_id,omitempty"`

	// Lecturer fields
	LecturerID string `json:"lecturer_id,omitempty"`
	Department string `json:"department,omitempty"`
}
