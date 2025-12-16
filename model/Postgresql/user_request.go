package model

import "github.com/google/uuid"

type CreateUserRequest struct {
	Username string    `json:"username" validate:"required"`
	Email    string    `json:"email" validate:"required,email"`
	Password string    `json:"password" validate:"required,min=6"`
	FullName string    `json:"full_name" validate:"required"`
	RoleID   uuid.UUID `json:"role_id" validate:"required"`
}

type UpdateUserRequest struct {
	Username string    `json:"username"`
	Email    string    `json:"email"`
	FullName string    `json:"full_name"`
	RoleID   uuid.UUID `json:"role_id"`
}

type UpdateRoleRequest struct {
	RoleID uuid.UUID `json:"role_id" validate:"required"`
}

type UpdateAdvisorRequest struct {
	AdvisorID uuid.UUID `json:"advisor_id" validate:"required"`
}
