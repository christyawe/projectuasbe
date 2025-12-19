package model

import "github.com/google/uuid"

type CreateUserDTO struct {
	Username string    `json:"username" validate:"required"`
	Email    string    `json:"email" validate:"required,email"`
	Password string    `json:"password" validate:"required,min=6"`
	FullName string    `json:"full_name" validate:"required"`
	RoleID   uuid.UUID `json:"role_id" validate:"required"`
}

type UpdateUserDTO struct {
	Username string    `json:"username"`
	Email    string    `json:"email"`
	FullName string    `json:"full_name"`
	RoleID   uuid.UUID `json:"role_id"`
}

type UpdateRoleDTO struct {
	RoleID uuid.UUID `json:"role_id" validate:"required"`
}

type UpdateAdvisorDTO struct {
	AdvisorID uuid.UUID `json:"advisor_id" validate:"required"`
}
