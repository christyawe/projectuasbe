package model

import (
	"time"

	"github.com/google/uuid"
)

type Lecturers struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	LecturerID string    `json:"lecturer_id"`
	Department string    `json:"department"`
	Created_at time.Time `json:"created_at"`
}

type LecturerWithUser struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	LecturerID string    `json:"lecturer_id"`
	Department string    `json:"department"`
	CreatedAt  time.Time `json:"created_at"`

	Username string `json:"username"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
}
