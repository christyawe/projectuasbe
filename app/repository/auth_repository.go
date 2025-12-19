package repository

import (
	"context"
	"errors"
	model "UASBE/app/model/Postgresql"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepository struct {
	DB *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{DB: db}
}

func (r *AuthRepository) FindUserByEmailOrUsername(identifier string) (*model.Users, string, error) {
	var user model.Users
	var roleName string

	query := `
		SELECT 
			u.id, u.username, u.email, u.password_hash, u.full_name, 
			u.role_id, u.is_active, u.created_at, u.updated_at,
			r.name
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.email = $1 OR u.username = $1
		LIMIT 1
	`

	err := r.DB.QueryRow(context.Background(), query, identifier).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.RoleID,
		&user.ISActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&roleName,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, "", errors.New("user not found")
		}
		return nil, "", err
	}

	return &user, roleName, nil
}

func (r *AuthRepository) GetPermissionsByRoleID(roleID uuid.UUID) ([]string, error) {
	query := `
		SELECT p.name
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
	`

	rows, err := r.DB.Query(context.Background(), query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var permName string
		if err := rows.Scan(&permName); err != nil {
			return nil, err
		}
		permissions = append(permissions, permName)
	}

	return permissions, nil
}
func (r *AuthRepository) GetUserProfile(userID uuid.UUID) (*model.UserProfileResponse, error) {
	var (
		user model.Users
		roleName string
	)

	// 1. Ambil data user (WAJIB ADA)
	userQuery := `
		SELECT 
			u.id, u.username, u.email, u.full_name,
			u.role_id, u.is_active, u.created_at, u.updated_at,
			r.name
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1
		LIMIT 1
	`

	err := r.DB.QueryRow(context.Background(), userQuery, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FullName,
		&user.RoleID,
		&user.ISActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&roleName,
	)

	if err != nil {
		return nil, errors.New("user not found")
	}

	response := &model.UserProfileResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		FullName: user.FullName,
		Role:     roleName,
	}

	// 2. Cek student
	studentQuery := `
		SELECT student_id, program_study, academic_year, advisor_id
		FROM students
		WHERE user_id = $1
		LIMIT 1
	`

	var student model.ProfileData
	var advisorID uuid.UUID

	err = r.DB.QueryRow(context.Background(), studentQuery, userID).
		Scan(&student.StudentID, &student.ProgramStudy, &student.AcademicYear, &advisorID)

	if err == nil {
		student.AdvisorID = &advisorID
		response.Profile = &student
		return response, nil
	}

	// 3. Cek lecturer
	lecturerQuery := `
		SELECT lecturer_id, department
		FROM lecturers
		WHERE user_id = $1
		LIMIT 1
	`

	var lecturer model.ProfileData
	err = r.DB.QueryRow(context.Background(), lecturerQuery, userID).
		Scan(&lecturer.LecturerID, &lecturer.Department)

	if err == nil {
		response.Profile = &lecturer
		return response, nil
	}

	// 4. ADMIN / ROLE LAIN → tetap return user info
	return response, nil
}
