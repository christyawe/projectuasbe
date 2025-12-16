package repository

import (
	"context"
	"database/sql"
	"time"

	model "uas_backend/app/model/Postgresql"

	"github.com/google/uuid"
)

type UserRepository interface {
	// User CRUD
	CreateUser(ctx context.Context, user *model.Users) error
	GetUserByID(ctx context.Context, userID uuid.UUID) (*model.Users, string, error)
	GetAllUsers(ctx context.Context, page, limit int) ([]model.Users, []string, int, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, req *model.UpdateUserRequest) error
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	UpdateUserRole(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error

	// Profile management
	CreateStudentProfile(ctx context.Context, student *model.Student) error
	CreateLecturerProfile(ctx context.Context, lecturer *model.Lecturers) error
	UpdateStudentAdvisor(ctx context.Context, studentID uuid.UUID, advisorID uuid.UUID) error

	// Students & Lecturers
	GetAllStudents(ctx context.Context, page, limit int) ([]model.StudentWithUser, int, error)
	GetStudentWithUserByID(ctx context.Context, studentID uuid.UUID) (*model.StudentWithUser, error)
	GetStudentByID(ctx context.Context, studentID uuid.UUID) (*model.Student, error)
	GetStudentAchievements(ctx context.Context, studentID uuid.UUID, page, limit int) ([]model.AchievementWithStudent, int, error)
	GetAllLecturers(ctx context.Context, page, limit int) ([]model.LecturerWithUser, int, error)
	GetLecturerByID(ctx context.Context, lecturerID uuid.UUID) (*model.Lecturers, error)
	GetStudentsByAdvisorID(ctx context.Context, advisorID uuid.UUID, page, limit int) ([]model.StudentWithUser, int, error)

	// Helper methods
	CheckUsernameExists(ctx context.Context, username string) (bool, error)
	CheckEmailExists(ctx context.Context, email string) (bool, error)
	GetRoleByID(ctx context.Context, roleID uuid.UUID) (*model.Roles, error)
}

type userRepo struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepo{db: db}
}

// CreateUser membuat user baru
func (r *userRepo) CreateUser(ctx context.Context, user *model.Users) error {
	query := `INSERT INTO users (id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash,
		user.FullName, user.RoleID, user.ISActive, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

// GetUserByID mengambil user berdasarkan ID dengan role name
func (r *userRepo) GetUserByID(ctx context.Context, userID uuid.UUID) (*model.Users, string, error) {
	query := `SELECT u.id, u.username, u.email, u.password_hash, u.full_name, u.role_id, u.is_active, u.created_at, u.updated_at, r.name as role_name
              FROM users u
              JOIN roles r ON u.role_id = r.id
              WHERE u.id = $1`

	var user model.Users
	var roleName string

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.FullName, &user.RoleID, &user.ISActive,
		&user.CreatedAt, &user.UpdatedAt, &roleName,
	)
	if err != nil {
		return nil, "", err
	}

	return &user, roleName, nil
}

// GetAllUsers mengambil semua users dengan pagination
func (r *userRepo) GetAllUsers(ctx context.Context, page, limit int) ([]model.Users, []string, int, error) {
	// Count total
	var total int
	countQuery := `SELECT COUNT(*) FROM users`
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, nil, 0, err
	}

	// Get users with pagination
	offset := (page - 1) * limit
	query := `SELECT u.id, u.username, u.email, u.password_hash, u.full_name, u.role_id, u.is_active, u.created_at, u.updated_at, r.name as role_name
              FROM users u
              JOIN roles r ON u.role_id = r.id
              ORDER BY u.created_at DESC
              LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, nil, 0, err
	}
	defer rows.Close()

	var users []model.Users
	var roleNames []string

	for rows.Next() {
		var user model.Users
		var roleName string

		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.PasswordHash,
			&user.FullName, &user.RoleID, &user.ISActive,
			&user.CreatedAt, &user.UpdatedAt, &roleName,
		)
		if err != nil {
			return nil, nil, 0, err
		}

		users = append(users, user)
		roleNames = append(roleNames, roleName)
	}

	return users, roleNames, total, nil
}

// UpdateUser mengupdate data user
func (r *userRepo) UpdateUser(ctx context.Context, userID uuid.UUID, req *model.UpdateUserRequest) error {
	query := `UPDATE users SET `
	args := []interface{}{}
	argCount := 1

	if req.Email != "" {
		query += `email = $` + string(rune(argCount+'0')) + `, `
		args = append(args, req.Email)
		argCount++
	}

	if req.FullName != "" {
		query += `full_name = $` + string(rune(argCount+'0')) + `, `
		args = append(args, req.FullName)
		argCount++
	}

	if req.IsActive != nil {
		query += `is_active = $` + string(rune(argCount+'0')) + `, `
		args = append(args, *req.IsActive)
		argCount++
	}

	query += `updated_at = $` + string(rune(argCount+'0')) + ` WHERE id = $` + string(rune(argCount+1+'0'))
	args = append(args, time.Now(), userID)

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// DeleteUser menghapus user (soft delete dengan set is_active = false)
func (r *userRepo) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET is_active = false, updated_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	return err
}

// UpdateUserRole mengupdate role user
func (r *userRepo) UpdateUserRole(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error {
	query := `UPDATE users SET role_id = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, roleID, time.Now(), userID)
	return err
}

// CreateStudentProfile membuat profile student
func (r *userRepo) CreateStudentProfile(ctx context.Context, student *model.Student) error {
	query := `INSERT INTO students (id, user_id, student_id, program_study, academic_year, advisor_id, created_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query,
		student.ID, student.UserID, student.StudentID, student.Program_Study,
		student.Academic_Year, student.AdvisorID, student.Created_at,
	)
	return err
}

// CreateLecturerProfile membuat profile lecturer
func (r *userRepo) CreateLecturerProfile(ctx context.Context, lecturer *model.Lecturers) error {
	query := `INSERT INTO lecturers (id, user_id, lecturer_id, department, created_at)
              VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.ExecContext(ctx, query,
		lecturer.ID, lecturer.UserID, lecturer.LecturerID, lecturer.Department, lecturer.Created_at,
	)
	return err
}

// UpdateStudentAdvisor mengupdate advisor mahasiswa
func (r *userRepo) UpdateStudentAdvisor(ctx context.Context, studentID uuid.UUID, advisorID uuid.UUID) error {
	query := `UPDATE students SET advisor_id = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, advisorID, studentID)
	return err
}

// CheckUsernameExists mengecek apakah username sudah ada
func (r *userRepo) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, username).Scan(&exists)
	return exists, err
}

// CheckEmailExists mengecek apakah email sudah ada
func (r *userRepo) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	return exists, err
}

// GetRoleByID mengambil role berdasarkan ID
func (r *userRepo) GetRoleByID(ctx context.Context, roleID uuid.UUID) (*model.Roles, error) {
	query := `SELECT id, name, description, created_at FROM roles WHERE id = $1`

	var role model.Roles
	err := r.db.QueryRowContext(ctx, query, roleID).Scan(
		&role.ID, &role.Name, &role.Description, &role.Created_at,
	)
	if err != nil {
		return nil, err
	}

	return &role, nil
}

// Students & Lecturers Repository Methods

// GetAllStudents mengambil semua students dengan user info
func (r *userRepo) GetAllStudents(ctx context.Context, page, limit int) ([]model.StudentWithUser, int, error) {
	// Count total
	var total int
	countQuery := `SELECT COUNT(*) FROM students s JOIN users u ON s.user_id = u.id WHERE u.is_active = true`
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get students with pagination
	offset := (page - 1) * limit
	query := `SELECT s.id, s.user_id, s.student_id, s.program_study, s.academic_year, s.advisor_id, s.created_at,
                     u.username, u.full_name, u.email,
                     COALESCE(l_user.full_name, '') as advisor_name
              FROM students s
              JOIN users u ON s.user_id = u.id
              LEFT JOIN lecturers l ON s.advisor_id = l.id
              LEFT JOIN users l_user ON l.user_id = l_user.id
              WHERE u.is_active = true
              ORDER BY s.created_at DESC
              LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var students []model.StudentWithUser

	for rows.Next() {
		var student model.StudentWithUser

		err := rows.Scan(
			&student.ID, &student.UserID, &student.StudentID, &student.Program_Study,
			&student.Academic_Year, &student.AdvisorID, &student.Created_at,
			&student.Username, &student.FullName, &student.Email, &student.AdvisorName,
		)
		if err != nil {
			return nil, 0, err
		}

		students = append(students, student)
	}

	return students, total, nil
}

// GetStudentWithUserByID mengambil student dengan user info berdasarkan ID
func (r *userRepo) GetStudentWithUserByID(ctx context.Context, studentID uuid.UUID) (*model.StudentWithUser, error) {
	query := `SELECT s.id, s.user_id, s.student_id, s.program_study, s.academic_year, s.advisor_id, s.created_at,
                     u.username, u.full_name, u.email,
                     COALESCE(l_user.full_name, '') as advisor_name
              FROM students s
              JOIN users u ON s.user_id = u.id
              LEFT JOIN lecturers l ON s.advisor_id = l.id
              LEFT JOIN users l_user ON l.user_id = l_user.id
              WHERE s.id = $1 AND u.is_active = true`

	var student model.StudentWithUser

	err := r.db.QueryRowContext(ctx, query, studentID).Scan(
		&student.ID, &student.UserID, &student.StudentID, &student.Program_Study,
		&student.Academic_Year, &student.AdvisorID, &student.Created_at,
		&student.Username, &student.FullName, &student.Email, &student.AdvisorName,
	)
	if err != nil {
		return nil, err
	}

	return &student, nil
}

// GetStudentByID mengambil student berdasarkan ID
func (r *userRepo) GetStudentByID(ctx context.Context, studentID uuid.UUID) (*model.Student, error) {
	query := `SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
              FROM students WHERE id = $1`

	var student model.Student

	err := r.db.QueryRowContext(ctx, query, studentID).Scan(
		&student.ID, &student.UserID, &student.StudentID, &student.Program_Study,
		&student.Academic_Year, &student.AdvisorID, &student.Created_at,
	)
	if err != nil {
		return nil, err
	}

	return &student, nil
}

// GetStudentAchievements mengambil achievements student
func (r *userRepo) GetStudentAchievements(ctx context.Context, studentID uuid.UUID, page, limit int) ([]model.AchievementWithStudent, int, error) {
	// Count total
	var total int
	countQuery := `SELECT COUNT(*) FROM achievement_references ar WHERE ar.student_id = $1 AND ar.status != 'deleted'`
	err := r.db.QueryRowContext(ctx, countQuery, studentID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get achievements with pagination
	offset := (page - 1) * limit
	query := `SELECT ar.id, ar.student_id, ar.mongo_achievement_id, ar.status, ar.submitted_at, ar.verified_at, ar.verified_by, ar.rejection_note, ar.created_at, ar.updated_at,
                     s.student_id as student_nim, u.full_name as student_name, s.program_study
              FROM achievement_references ar
              JOIN students s ON ar.student_id = s.id
              JOIN users u ON s.user_id = u.id
              WHERE ar.student_id = $1 AND ar.status != 'deleted'
              ORDER BY ar.created_at DESC
              LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, studentID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var achievements []model.AchievementWithStudent

	for rows.Next() {
		var achievement model.AchievementWithStudent

		err := rows.Scan(
			&achievement.ID, &achievement.StudentID, &achievement.MongoAchievementID,
			&achievement.Status, &achievement.SubmittedAt, &achievement.VerifiedAt,
			&achievement.VerifiedBy, &achievement.RejectionNote, &achievement.CreatedAt,
			&achievement.UpdatedAt, &achievement.StudentNIM, &achievement.StudentName,
			&achievement.ProgramStudy,
		)
		if err != nil {
			return nil, 0, err
		}

		achievements = append(achievements, achievement)
	}

	return achievements, total, nil
}

// GetAllLecturers mengambil semua lecturers dengan user info
func (r *userRepo) GetAllLecturers(ctx context.Context, page, limit int) ([]model.LecturerWithUser, int, error) {
	// Count total
	var total int
	countQuery := `SELECT COUNT(*) FROM lecturers l JOIN users u ON l.user_id = u.id WHERE u.is_active = true`
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get lecturers with pagination
	offset := (page - 1) * limit
	query := `SELECT l.id, l.user_id, l.lecturer_id, l.department, l.created_at,
                     u.username, u.full_name, u.email
              FROM lecturers l
              JOIN users u ON l.user_id = u.id
              WHERE u.is_active = true
              ORDER BY l.created_at DESC
              LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var lecturers []model.LecturerWithUser

	for rows.Next() {
		var lecturer model.LecturerWithUser

		err := rows.Scan(
			&lecturer.ID, &lecturer.UserID, &lecturer.LecturerID, &lecturer.Department,
			&lecturer.Created_at, &lecturer.Username, &lecturer.FullName, &lecturer.Email,
		)
		if err != nil {
			return nil, 0, err
		}

		lecturers = append(lecturers, lecturer)
	}

	return lecturers, total, nil
}

// GetLecturerByID mengambil lecturer berdasarkan ID
func (r *userRepo) GetLecturerByID(ctx context.Context, lecturerID uuid.UUID) (*model.Lecturers, error) {
	query := `SELECT id, user_id, lecturer_id, department, created_at
              FROM lecturers WHERE id = $1`

	var lecturer model.Lecturers

	err := r.db.QueryRowContext(ctx, query, lecturerID).Scan(
		&lecturer.ID, &lecturer.UserID, &lecturer.LecturerID, &lecturer.Department, &lecturer.Created_at,
	)
	if err != nil {
		return nil, err
	}

	return &lecturer, nil
}

// GetStudentsByAdvisorID mengambil students berdasarkan advisor ID
func (r *userRepo) GetStudentsByAdvisorID(ctx context.Context, advisorID uuid.UUID, page, limit int) ([]model.StudentWithUser, int, error) {
	// Count total
	var total int
	countQuery := `SELECT COUNT(*) FROM students s JOIN users u ON s.user_id = u.id WHERE s.advisor_id = $1 AND u.is_active = true`
	err := r.db.QueryRowContext(ctx, countQuery, advisorID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get students with pagination
	offset := (page - 1) * limit
	query := `SELECT s.id, s.user_id, s.student_id, s.program_study, s.academic_year, s.advisor_id, s.created_at,
                     u.username, u.full_name, u.email,
                     COALESCE(l_user.full_name, '') as advisor_name
              FROM students s
              JOIN users u ON s.user_id = u.id
              LEFT JOIN lecturers l ON s.advisor_id = l.id
              LEFT JOIN users l_user ON l.user_id = l_user.id
              WHERE s.advisor_id = $1 AND u.is_active = true
              ORDER BY s.created_at DESC
              LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, advisorID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var students []model.StudentWithUser

	for rows.Next() {
		var student model.StudentWithUser

		err := rows.Scan(
			&student.ID, &student.UserID, &student.StudentID, &student.Program_Study,
			&student.Academic_Year, &student.AdvisorID, &student.Created_at,
			&student.Username, &student.FullName, &student.Email, &student.AdvisorName,
		)
		if err != nil {
			return nil, 0, err
		}

		students = append(students, student)
	}

	return students, total, nil
}
