package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	mongodb "UASBE/model/MongoDB"
	model "UASBE/model/Postgresql"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AchievementRepository interface {
	GetStudentByUserID(ctx context.Context, userID uuid.UUID) (*model.Student, error)
	SaveAchievementMongo(ctx context.Context, achievement mongodb.Achievement) (string, error)
	SaveAchievementReference(ctx context.Context, ref model.AchievementReference) error
	GetAchievementReferenceByID(ctx context.Context, achievementID uuid.UUID) (*model.AchievementReference, error)
	UpdateAchievementStatusToSubmitted(ctx context.Context, achievementID uuid.UUID) error
	GetAdvisorIDByStudentID(ctx context.Context, studentID uuid.UUID) (uuid.UUID, error)
	SoftDeleteAchievementMongo(ctx context.Context, mongoAchievementID string) error
	UpdateAchievementReferenceToDeleted(ctx context.Context, achievementID uuid.UUID) error
	GetLecturerByUserID(ctx context.Context, userID uuid.UUID) (*model.Lecturers, error)
	GetStudentIDsByAdvisorID(ctx context.Context, advisorID uuid.UUID) ([]uuid.UUID, error)
	GetAchievementsWithStudentInfo(ctx context.Context, studentIDs []uuid.UUID, status string, page, limit int) ([]model.AchievementWithStudent, int, error)
	GetAchievementDetailFromMongo(ctx context.Context, mongoAchievementID string) (*mongodb.Achievement, error)
	UpdateAchievementStatusToVerified(ctx context.Context, achievementID uuid.UUID, lecturerID uuid.UUID) error
	GetStudentByID(ctx context.Context, studentID uuid.UUID) (*model.Student, error)
	UpdateAchievementStatusToRejected(ctx context.Context, achievementID uuid.UUID, rejectionNote string) error
	GetAchievementStatusHistory(ctx context.Context, achievementID uuid.UUID) ([]model.AchievementStatusLog, error)
	LogAchievementStatusChange(ctx context.Context, log model.AchievementStatusLog) error
	GetAllAchievementsForAdmin(ctx context.Context, filters model.AdminAchievementFilters, page, limit int) ([]model.AchievementWithStudent, int, error)
	GetStatisticsByType(ctx context.Context, studentIDs []uuid.UUID, filters model.StatisticsFilters, mongoColl *mongo.Collection) ([]model.StatsByType, error)
	GetStatisticsByPeriod(ctx context.Context, studentIDs []uuid.UUID, filters model.StatisticsFilters) ([]model.StatsByPeriod, error)
	GetTopStudents(ctx context.Context, studentIDs []uuid.UUID, filters model.StatisticsFilters, limit int) ([]model.TopStudent, error)
	GetAchievementsByID(ctx context.Context, userID uuid.UUID) (*model.Users, error)
	UpdateAchievementInMongo(ctx context.Context, achievement mongodb.Achievement) error
	UpdateAchievementTimestamp(ctx context.Context, achievementID uuid.UUID) error
	GetLevelDistribution(ctx context.Context, studentIDs []uuid.UUID, filters model.StatisticsFilters, mongoColl *mongo.Collection) ([]model.LevelDistribution, error)
	GetStatusDistribution(ctx context.Context, studentIDs []uuid.UUID, filters model.StatisticsFilters) ([]model.StatusDistribution, error)
	GetTotalAchievements(ctx context.Context, studentIDs []uuid.UUID, filters model.StatisticsFilters) (int, error)
	AddAttachmentToAchievement(ctx context.Context, mongoAchievementID, fileName, fileURL, fileType string) error
	GetStudentWithUserByID(ctx context.Context, studentID uuid.UUID) (*model.StudentWithUser, error)
	GetStudentAchievements(ctx context.Context, studentID uuid.UUID, page, limit int) ([]model.AchievementWithStudent, int, error)
}

type achievementRepo struct {
	pgDB      *sql.DB
	mongoColl *mongo.Collection
}

func NewAchievementRepository(pgDB *sql.DB, mongoColl *mongo.Collection) AchievementRepository {
	return &achievementRepo{
		pgDB:      pgDB,
		mongoColl: mongoColl,
	}
}

// GetStudentByUserID mengambil data student dari Postgres berdasarkan user_id (akun login)
func (r *achievementRepo) GetStudentByUserID(ctx context.Context, userID uuid.UUID) (*model.Student, error) {
	query := `SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at 
              FROM students WHERE user_id = $1`

	var s model.Student
	err := r.pgDB.QueryRowContext(ctx, query, userID).Scan(
		&s.ID, &s.UserID, &s.StudentID, &s.Program_Study,
		&s.Academic_Year, &s.AdvisorID, &s.Created_at,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// SaveAchievementMongo menyimpan detail lengkap prestasi ke MongoDB
func (r *achievementRepo) SaveAchievementMongo(ctx context.Context, achievement mongodb.Achievement) (string, error) {
	res, err := r.mongoColl.InsertOne(ctx, achievement)
	if err != nil {
		return "", err
	}
	// Mengembalikan ID Mongo dalam bentuk Hex String
	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

// SaveAchievementReference menyimpan referensi status ke PostgreSQL
func (r *achievementRepo) SaveAchievementReference(ctx context.Context, ref model.AchievementReference) error {
	query := `INSERT INTO achievement_references (
		id, student_id, mongo_achievement_id, status, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.pgDB.ExecContext(ctx, query,
		ref.ID, ref.StudentID, ref.MongoAchievementID, ref.Status, ref.CreatedAt, ref.UpdatedAt,
	)
	return err
}

// GetAchievementReferenceByID mengambil data achievement reference berdasarkan ID
func (r *achievementRepo) GetAchievementReferenceByID(ctx context.Context, achievementID uuid.UUID) (*model.AchievementReference, error) {
	query := `SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at, 
              verified_by, rejection_note, created_at, updated_at 
              FROM achievement_references WHERE id = $1`

	var ref model.AchievementReference
	err := r.pgDB.QueryRowContext(ctx, query, achievementID).Scan(
		&ref.ID, &ref.StudentID, &ref.MongoAchievementID, &ref.Status,
		&ref.SubmittedAt, &ref.VerifiedAt, &ref.VerifiedBy, &ref.RejectionNote,
		&ref.CreatedAt, &ref.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &ref, nil
}

// UpdateAchievementStatusToSubmitted mengupdate status achievement menjadi 'submitted'
func (r *achievementRepo) UpdateAchievementStatusToSubmitted(ctx context.Context, achievementID uuid.UUID) error {
	query := `UPDATE achievement_references 
              SET status = 'submitted', submitted_at = $1, updated_at = $2 
              WHERE id = $3`

	now := time.Now()
	_, err := r.pgDB.ExecContext(ctx, query, now, now, achievementID)
	return err
}

// GetAdvisorIDByStudentID mengambil advisor_id dari student
func (r *achievementRepo) GetAdvisorIDByStudentID(ctx context.Context, studentID uuid.UUID) (uuid.UUID, error) {
	query := `SELECT advisor_id FROM students WHERE id = $1`

	var advisorID uuid.UUID
	err := r.pgDB.QueryRowContext(ctx, query, studentID).Scan(&advisorID)
	if err != nil {
		return uuid.Nil, err
	}
	return advisorID, nil
}

// SoftDeleteAchievementMongo melakukan soft delete di MongoDB dengan menambahkan field deleted_at
func (r *achievementRepo) SoftDeleteAchievementMongo(ctx context.Context, mongoAchievementID string) error {
	objectID, err := primitive.ObjectIDFromHex(mongoAchievementID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$set": bson.M{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		},
	}

	_, err = r.mongoColl.UpdateOne(ctx, filter, update)
	return err
}

// UpdateAchievementReferenceToDeleted mengupdate status achievement reference menjadi 'deleted'
func (r *achievementRepo) UpdateAchievementReferenceToDeleted(ctx context.Context, achievementID uuid.UUID) error {
	query := `UPDATE achievement_references 
              SET status = 'deleted', updated_at = $1 
              WHERE id = $2`

	now := time.Now()
	_, err := r.pgDB.ExecContext(ctx, query, now, achievementID)
	return err
}

// GetLecturerByUserID mengambil data lecturer dari Postgres berdasarkan user_id
func (r *achievementRepo) GetLecturerByUserID(ctx context.Context, userID uuid.UUID) (*model.Lecturers, error) {
	query := `SELECT id, user_id, lecturer_id, department, created_at 
              FROM lecturers WHERE user_id = $1`

	var l model.Lecturers
	err := r.pgDB.QueryRowContext(ctx, query, userID).Scan(
		&l.ID, &l.UserID, &l.LecturerID, &l.Department, &l.Created_at,
	)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

// GetStudentIDsByAdvisorID mengambil list student IDs berdasarkan advisor_id
func (r *achievementRepo) GetStudentIDsByAdvisorID(ctx context.Context, advisorID uuid.UUID) ([]uuid.UUID, error) {
	query := `SELECT id FROM students WHERE advisor_id = $1`

	rows, err := r.pgDB.QueryContext(ctx, query, advisorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var studentIDs []uuid.UUID
	for rows.Next() {
		var studentID uuid.UUID
		if err := rows.Scan(&studentID); err != nil {
			return nil, err
		}
		studentIDs = append(studentIDs, studentID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return studentIDs, nil
}

// GetAchievementsWithStudentInfo mengambil achievements dengan info student, dengan pagination
func (r *achievementRepo) GetAchievementsWithStudentInfo(ctx context.Context, studentIDs []uuid.UUID, status string, page, limit int) ([]model.AchievementWithStudent, int, error) {
	if len(studentIDs) == 0 {
		return []model.AchievementWithStudent{}, 0, nil
	}

	// Build query dengan filter
	baseQuery := `
		SELECT 
			ar.id, ar.student_id, ar.mongo_achievement_id, ar.status,
			ar.submitted_at, ar.verified_at, ar.verified_by, ar.rejection_note,
			ar.created_at, ar.updated_at,
			s.student_id as student_nim, u.full_name as student_name, s.program_study
		FROM achievement_references ar
		JOIN students s ON ar.student_id = s.id
		JOIN users u ON s.user_id = u.id
		WHERE ar.student_id = ANY($1)
	`

	countQuery := `
		SELECT COUNT(*)
		FROM achievement_references ar
		WHERE ar.student_id = ANY($1)
	`

	args := []interface{}{pq.Array(studentIDs)}
	argCount := 2

	// Add status filter if provided
	if status != "" {
		baseQuery += ` AND ar.status = $` + fmt.Sprintf("%d", argCount)
		countQuery += ` AND ar.status = $` + fmt.Sprintf("%d", argCount)
		args = append(args, status)
		argCount++
	}

	// Get total count
	var total int
	err := r.pgDB.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Add ordering and pagination
	baseQuery += ` ORDER BY ar.created_at DESC LIMIT $` + fmt.Sprintf("%d", argCount) + ` OFFSET $` + fmt.Sprintf("%d", argCount+1)
	offset := (page - 1) * limit
	args = append(args, limit, offset)

	// Execute query
	rows, err := r.pgDB.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var achievements []model.AchievementWithStudent
	for rows.Next() {
		var a model.AchievementWithStudent
		err := rows.Scan(
			&a.ID, &a.StudentID, &a.MongoAchievementID, &a.Status,
			&a.SubmittedAt, &a.VerifiedAt, &a.VerifiedBy, &a.RejectionNote,
			&a.CreatedAt, &a.UpdatedAt,
			&a.StudentNIM, &a.StudentName, &a.ProgramStudy,
		)
		if err != nil {
			return nil, 0, err
		}
		achievements = append(achievements, a)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return achievements, total, nil
}

// GetAchievementDetailFromMongo mengambil detail achievement dari MongoDB
func (r *achievementRepo) GetAchievementDetailFromMongo(ctx context.Context, mongoAchievementID string) (*mongodb.Achievement, error) {
	objectID, err := primitive.ObjectIDFromHex(mongoAchievementID)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objectID}
	var achievement mongodb.Achievement

	err = r.mongoColl.FindOne(ctx, filter).Decode(&achievement)
	if err != nil {
		return nil, err
	}

	return &achievement, nil
}

// UpdateAchievementStatusToVerified mengupdate status achievement menjadi 'verified'
func (r *achievementRepo) UpdateAchievementStatusToVerified(ctx context.Context, achievementID uuid.UUID, lecturerID uuid.UUID) error {
	query := `UPDATE achievement_references 
              SET status = 'verified', verified_by = $1, verified_at = $2, updated_at = $3 
              WHERE id = $4`

	now := time.Now()
	_, err := r.pgDB.ExecContext(ctx, query, lecturerID, now, now, achievementID)
	return err
}

// GetStudentByID mengambil data student dari Postgres berdasarkan student ID
func (r *achievementRepo) GetStudentByID(ctx context.Context, studentID uuid.UUID) (*model.Student, error) {
	query := `SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at 
              FROM students WHERE id = $1`

	var s model.Student
	err := r.pgDB.QueryRowContext(ctx, query, studentID).Scan(
		&s.ID, &s.UserID, &s.StudentID, &s.Program_Study,
		&s.Academic_Year, &s.AdvisorID, &s.Created_at,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// UpdateAchievementStatusToRejected mengupdate status achievement menjadi 'rejected' dengan rejection note
func (r *achievementRepo) UpdateAchievementStatusToRejected(ctx context.Context, achievementID uuid.UUID, rejectionNote string) error {
	query := `UPDATE achievement_references 
              SET status = 'rejected', rejection_note = $1, updated_at = $2 
              WHERE id = $3`

	now := time.Now()
	_, err := r.pgDB.ExecContext(ctx, query, rejectionNote, now, achievementID)
	return err
}

// GetAchievementStatusHistory mengambil riwayat perubahan status achievement
func (r *achievementRepo) GetAchievementStatusHistory(ctx context.Context, achievementID uuid.UUID) ([]model.AchievementStatusLog, error) {
	query := `
		SELECT 
			asl.id, asl.achievement_id, asl.status, asl.changed_by, 
			u.full_name as changed_by_name, asl.rejection_note, asl.created_at
		FROM achievement_status_logs asl
		LEFT JOIN users u ON asl.changed_by = u.id
		WHERE asl.achievement_id = $1
		ORDER BY asl.created_at ASC
	`

	rows, err := r.pgDB.QueryContext(ctx, query, achievementID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []model.AchievementStatusLog
	for rows.Next() {
		var log model.AchievementStatusLog
		err := rows.Scan(
			&log.ID, &log.AchievementID, &log.Status, &log.ChangedBy,
			&log.ChangedByName, &log.RejectionNote, &log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}

// LogAchievementStatusChange mencatat perubahan status achievement ke log table
func (r *achievementRepo) LogAchievementStatusChange(ctx context.Context, log model.AchievementStatusLog) error {
	query := `INSERT INTO achievement_status_logs (
		id, achievement_id, status, changed_by, rejection_note, created_at
	) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.pgDB.ExecContext(ctx, query,
		log.ID, log.AchievementID, log.Status, log.ChangedBy, log.RejectionNote, log.CreatedAt,
	)
	return err
}

// GetAllAchievementsForAdmin mengambil semua achievements untuk admin dengan filters
func (r *achievementRepo) GetAllAchievementsForAdmin(ctx context.Context, filters model.AdminAchievementFilters, page, limit int) ([]model.AchievementWithStudent, int, error) {
	// Build query dengan filters
	baseQuery := `
		SELECT 
			ar.id, ar.student_id, ar.mongo_achievement_id, ar.status,
			ar.submitted_at, ar.verified_at, ar.verified_by, ar.rejection_note,
			ar.created_at, ar.updated_at,
			s.student_id as student_nim, u.full_name as student_name, s.program_study
		FROM achievement_references ar
		JOIN students s ON ar.student_id = s.id
		JOIN users u ON s.user_id = u.id
		WHERE 1=1
	`

	countQuery := `
		SELECT COUNT(*)
		FROM achievement_references ar
		JOIN students s ON ar.student_id = s.id
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 1

	// Apply filters
	if filters.Status != "" {
		baseQuery += fmt.Sprintf(" AND ar.status = $%d", argCount)
		countQuery += fmt.Sprintf(" AND ar.status = $%d", argCount)
		args = append(args, filters.Status)
		argCount++
	}

	if filters.StudentID != nil {
		baseQuery += fmt.Sprintf(" AND ar.student_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND ar.student_id = $%d", argCount)
		args = append(args, *filters.StudentID)
		argCount++
	}

	if filters.DateFrom != nil {
		baseQuery += fmt.Sprintf(" AND ar.created_at >= $%d", argCount)
		countQuery += fmt.Sprintf(" AND ar.created_at >= $%d", argCount)
		args = append(args, *filters.DateFrom)
		argCount++
	}

	if filters.DateTo != nil {
		baseQuery += fmt.Sprintf(" AND ar.created_at <= $%d", argCount)
		countQuery += fmt.Sprintf(" AND ar.created_at <= $%d", argCount)
		args = append(args, *filters.DateTo)
		argCount++
	}

	// Get total count
	var total int
	err := r.pgDB.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortBy := "ar.created_at"
	if filters.SortBy != "" {
		switch filters.SortBy {
		case "created_at":
			sortBy = "ar.created_at"
		case "updated_at":
			sortBy = "ar.updated_at"
		case "status":
			sortBy = "ar.status"
		}
	}

	sortOrder := "DESC"
	if filters.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	baseQuery += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	// Add pagination
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	offset := (page - 1) * limit
	args = append(args, limit, offset)

	// Execute query
	rows, err := r.pgDB.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var achievements []model.AchievementWithStudent
	for rows.Next() {
		var a model.AchievementWithStudent
		err := rows.Scan(
			&a.ID, &a.StudentID, &a.MongoAchievementID, &a.Status,
			&a.SubmittedAt, &a.VerifiedAt, &a.VerifiedBy, &a.RejectionNote,
			&a.CreatedAt, &a.UpdatedAt,
			&a.StudentNIM, &a.StudentName, &a.ProgramStudy,
		)
		if err != nil {
			return nil, 0, err
		}
		achievements = append(achievements, a)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return achievements, total, nil
}

// GetTotalAchievements menghitung total achievements
func (r *achievementRepo) GetTotalAchievements(ctx context.Context, studentIDs []uuid.UUID, filters model.StatisticsFilters) (int, error) {
	query := `SELECT COUNT(*) FROM achievement_references WHERE student_id = ANY($1)`
	args := []interface{}{pq.Array(studentIDs)}
	argCount := 2

	if filters.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, filters.Status)
		argCount++
	}

	if filters.DateFrom != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, *filters.DateFrom)
		argCount++
	}

	if filters.DateTo != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, *filters.DateTo)
		argCount++
	}

	var total int
	err := r.pgDB.QueryRowContext(ctx, query, args...).Scan(&total)
	return total, err
}

// GetStatisticsByType menghitung statistik berdasarkan tipe achievement dari MongoDB
func (r *achievementRepo) GetStatisticsByType(ctx context.Context, studentIDs []uuid.UUID, filters model.StatisticsFilters, mongoColl *mongo.Collection) ([]model.StatsByType, error) {
	// Get achievement references
	query := `SELECT mongo_achievement_id FROM achievement_references WHERE student_id = ANY($1)`
	args := []interface{}{pq.Array(studentIDs)}
	argCount := 2

	if filters.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, filters.Status)
		argCount++
	}

	if filters.DateFrom != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, *filters.DateFrom)
		argCount++
	}

	if filters.DateTo != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, *filters.DateTo)
		argCount++
	}

	rows, err := r.pgDB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Collect mongo IDs
	var mongoIDs []primitive.ObjectID
	for rows.Next() {
		var mongoIDStr string
		if err := rows.Scan(&mongoIDStr); err != nil {
			continue
		}
		mongoID, err := primitive.ObjectIDFromHex(mongoIDStr)
		if err != nil {
			continue
		}
		mongoIDs = append(mongoIDs, mongoID)
	}

	// Count by category from MongoDB
	typeCount := make(map[string]int)
	for _, mongoID := range mongoIDs {
		var achievement mongodb.Achievement
		err := mongoColl.FindOne(ctx, bson.M{"_id": mongoID}).Decode(&achievement)
		if err == nil {
			typeCount[achievement.AchievementType]++
		}
	}

	// Convert to slice
	var stats []model.StatsByType
	for typ, count := range typeCount {
		stats = append(stats, model.StatsByType{
			Type:  typ,
			Count: count,
		})
	}

	return stats, nil
}

// GetStatisticsByPeriod menghitung statistik berdasarkan periode (bulan)
func (r *achievementRepo) GetStatisticsByPeriod(ctx context.Context, studentIDs []uuid.UUID, filters model.StatisticsFilters) ([]model.StatsByPeriod, error) {
	query := `
		SELECT 
			TO_CHAR(created_at, 'YYYY-MM') as period,
			COUNT(*) as count
		FROM achievement_references
		WHERE student_id = ANY($1)
	`
	args := []interface{}{pq.Array(studentIDs)}
	argCount := 2

	if filters.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, filters.Status)
		argCount++
	}

	if filters.DateFrom != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, *filters.DateFrom)
		argCount++
	}

	if filters.DateTo != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, *filters.DateTo)
		argCount++
	}

	query += " GROUP BY period ORDER BY period DESC"

	rows, err := r.pgDB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []model.StatsByPeriod
	for rows.Next() {
		var stat model.StatsByPeriod
		if err := rows.Scan(&stat.Period, &stat.Count); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

// GetTopStudents mendapatkan top students berdasarkan jumlah achievement
func (r *achievementRepo) GetTopStudents(ctx context.Context, studentIDs []uuid.UUID, filters model.StatisticsFilters, limit int) ([]model.TopStudent, error) {
	query := `
		SELECT 
			ar.student_id,
			s.student_id as student_nim,
			u.full_name as student_name,
			s.program_study,
			COUNT(*) as count
		FROM achievement_references ar
		JOIN students s ON ar.student_id = s.id
		JOIN users u ON s.user_id = u.id
		WHERE ar.student_id = ANY($1)
	`
	args := []interface{}{pq.Array(studentIDs)}
	argCount := 2

	if filters.Status != "" {
		query += fmt.Sprintf(" AND ar.status = $%d", argCount)
		args = append(args, filters.Status)
		argCount++
	}

	if filters.DateFrom != nil {
		query += fmt.Sprintf(" AND ar.created_at >= $%d", argCount)
		args = append(args, *filters.DateFrom)
		argCount++
	}

	if filters.DateTo != nil {
		query += fmt.Sprintf(" AND ar.created_at <= $%d", argCount)
		args = append(args, *filters.DateTo)
		argCount++
	}

	query += fmt.Sprintf(" GROUP BY ar.student_id, s.student_id, u.full_name, s.program_study ORDER BY count DESC LIMIT $%d", argCount)
	args = append(args, limit)

	rows, err := r.pgDB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topStudents []model.TopStudent
	for rows.Next() {
		var student model.TopStudent
		if err := rows.Scan(&student.StudentID, &student.StudentNIM, &student.StudentName, &student.ProgramStudy, &student.Count); err != nil {
			return nil, err
		}
		topStudents = append(topStudents, student)
	}

	return topStudents, nil
}

// GetLevelDistribution mendapatkan distribusi berdasarkan level dari MongoDB
func (r *achievementRepo) GetLevelDistribution(ctx context.Context, studentIDs []uuid.UUID, filters model.StatisticsFilters, mongoColl *mongo.Collection) ([]model.LevelDistribution, error) {
	// Get achievement references
	query := `SELECT mongo_achievement_id FROM achievement_references WHERE student_id = ANY($1)`
	args := []interface{}{pq.Array(studentIDs)}
	argCount := 2

	if filters.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, filters.Status)
		argCount++
	}

	if filters.DateFrom != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, *filters.DateFrom)
		argCount++
	}

	if filters.DateTo != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, *filters.DateTo)
		argCount++
	}

	rows, err := r.pgDB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Collect mongo IDs
	var mongoIDs []primitive.ObjectID
	for rows.Next() {
		var mongoIDStr string
		if err := rows.Scan(&mongoIDStr); err != nil {
			continue
		}
		mongoID, err := primitive.ObjectIDFromHex(mongoIDStr)
		if err != nil {
			continue
		}
		mongoIDs = append(mongoIDs, mongoID)
	}

	// Count by level from MongoDB
	levelCount := make(map[string]int)
	for _, mongoID := range mongoIDs {
		var achievement mongodb.Achievement
		err := mongoColl.FindOne(ctx, bson.M{"_id": mongoID}).Decode(&achievement)
		if err == nil {
			// Get level from Details.CompetitionLevel
			if achievement.Details.CompetitionLevel != nil {
				levelCount[*achievement.Details.CompetitionLevel]++
			}
		}
	}

	// Convert to slice
	var stats []model.LevelDistribution
	for level, count := range levelCount {
		stats = append(stats, model.LevelDistribution{
			Level: level,
			Count: count,
		})
	}

	return stats, nil
}

// GetStatusDistribution mendapatkan distribusi berdasarkan status
func (r *achievementRepo) GetStatusDistribution(ctx context.Context, studentIDs []uuid.UUID, filters model.StatisticsFilters) ([]model.StatusDistribution, error) {
	query := `
		SELECT 
			status,
			COUNT(*) as count
		FROM achievement_references
		WHERE student_id = ANY($1)
	`
	args := []interface{}{pq.Array(studentIDs)}
	argCount := 2

	if filters.DateFrom != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, *filters.DateFrom)
		argCount++
	}

	if filters.DateTo != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, *filters.DateTo)
		argCount++
	}

	query += " GROUP BY status ORDER BY count DESC"

	rows, err := r.pgDB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []model.StatusDistribution
	for rows.Next() {
		var stat model.StatusDistribution
		if err := rows.Scan(&stat.Status, &stat.Count); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

// UpdateAchievementInMongo updates achievement data in MongoDB
func (r *achievementRepo) UpdateAchievementInMongo(ctx context.Context, achievement mongodb.Achievement) error {
	filter := bson.M{"_id": achievement.ID}
	update := bson.M{"$set": achievement}

	_, err := r.mongoColl.UpdateOne(ctx, filter, update)
	return err
}

// UpdateAchievementTimestamp updates the updated_at timestamp in PostgreSQL
func (r *achievementRepo) UpdateAchievementTimestamp(ctx context.Context, achievementID uuid.UUID) error {
	query := `UPDATE achievement_references SET updated_at = $1 WHERE id = $2`
	_, err := r.pgDB.ExecContext(ctx, query, time.Now(), achievementID)
	return err
}

// GetUserByID gets user data by ID
func (r *achievementRepo) GetAchievementsByID(ctx context.Context, userID uuid.UUID) (*model.Users, error) {
	query := `SELECT id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at
              FROM achievements WHERE id = $1`

	var user model.Users
	err := r.pgDB.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName,
		&user.RoleID, &user.ISActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// AddAttachmentToAchievement adds an attachment to an achievement in MongoDB
func (r *achievementRepo) AddAttachmentToAchievement(ctx context.Context, mongoAchievementID, fileName, fileURL, fileType string) error {
	objectID, err := primitive.ObjectIDFromHex(mongoAchievementID)
	if err != nil {
		return err
	}

	attachment := mongodb.Attachment{
		FileName:   fileName,
		FileUrl:    fileURL,
		FileType:   fileType,
		UploadedAt: time.Now(),
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$push": bson.M{"attachments": attachment},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	_, err = r.mongoColl.UpdateOne(ctx, filter, update)
	return err
}

// GetStudentWithUserByID gets student with user info by student ID
func (r *achievementRepo) GetStudentWithUserByID(ctx context.Context, studentID uuid.UUID) (*model.StudentWithUser, error) {
	query := `SELECT s.id, s.user_id, s.student_id, s.program_study, s.academic_year, s.advisor_id, s.created_at,
                     u.username, u.full_name, u.email,
                     COALESCE(l_user.full_name, '') as advisor_name
              FROM students s
              JOIN users u ON s.user_id = u.id
              LEFT JOIN lecturers l ON s.advisor_id = l.id
              LEFT JOIN users l_user ON l.user_id = l_user.id
              WHERE s.id = $1 AND u.is_active = true`

	var student model.StudentWithUser

	err := r.pgDB.QueryRowContext(ctx, query, studentID).Scan(
		&student.ID, &student.UserID, &student.StudentID, &student.Program_Study,
		&student.Academic_Year, &student.AdvisorID, &student.Created_at,
		&student.Username, &student.FullName, &student.Email, &student.AdvisorName,
	)
	if err != nil {
		return nil, err
	}

	return &student, nil
}

// GetStudentAchievements gets achievements for a specific student
func (r *achievementRepo) GetStudentAchievements(ctx context.Context, studentID uuid.UUID, page, limit int) ([]model.AchievementWithStudent, int, error) {
	// Count total
	var total int
	countQuery := `SELECT COUNT(*) FROM achievement_references ar WHERE ar.student_id = $1 AND ar.status != 'deleted'`
	err := r.pgDB.QueryRowContext(ctx, countQuery, studentID).Scan(&total)
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

	rows, err := r.pgDB.QueryContext(ctx, query, studentID, limit, offset)
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
