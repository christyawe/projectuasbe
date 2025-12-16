package repository

import (
	"database/sql"
	"errors"
	model "UASBE/model/Postgresql"

	"github.com/google/uuid"
)

type AuthRepository struct {
	DB *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
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

	err := r.DB.QueryRow(query, identifier).Scan(
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
		if err == sql.ErrNoRows {
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

	rows, err := r.DB.Query(query, roleID)
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
func (r *AuthRepository) GetUserProfile(userID uuid.UUID) (*model.ProfileData, error) {
	var user model.Users
	var roleName string

	// Get user basic info
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

	err := r.DB.QueryRow(userQuery, userID).Scan(
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
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	profileData := &model.ProfileData{}

	// Check if user is a student
	studentQuery := `
		SELECT student_id, program_study, academic_year, advisor_id
		FROM students
		WHERE user_id = $1
		LIMIT 1
	`

	var studentID, programStudy, academicYear string
	var advisorID uuid.UUID
	err = r.DB.QueryRow(studentQuery, userID).Scan(&studentID, &programStudy, &academicYear, &advisorID)
	if err == nil {
		// User is a student
		profileData.StudentID = studentID
		profileData.ProgramStudy = programStudy
		profileData.AcademicYear = academicYear
		profileData.AdvisorID = advisorID
		return profileData, nil
	}

	// Check if user is a lecturer
	lecturerQuery := `
		SELECT lecturer_id, department
		FROM lecturers
		WHERE user_id = $1
		LIMIT 1
	`

	var lecturerID, department string
	err = r.DB.QueryRow(lecturerQuery, userID).Scan(&lecturerID, &department)
	if err == nil {
		// User is a lecturer
		profileData.LecturerID = lecturerID
		profileData.Department = department
		return profileData, nil
	}

	// User has no specific profile (admin or other role)
	return profileData, nil
}
