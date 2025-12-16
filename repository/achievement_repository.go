package repository

import (
	"context"
	"database/sql"

	mongodb "UASBE/model/MongoDB"
	model "UASBE/model/Postgresql"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AchievementRepository interface {
	GetStudentByUserID(ctx context.Context, userID uuid.UUID) (*model.Student, error)
	SaveAchievementMongo(ctx context.Context, achievement mongodb.Achievement) (string, error)
	SaveAchievementReference(ctx context.Context, ref model.AchievementReference) error
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
