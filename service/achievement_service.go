package service

import (
	"context"
	"errors"
	"time"
	mongodb "UASBE/model/MongoDB"
	model "UASBE/model/Postgresql"
	"UASBE/repository"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AchievementService interface {
	SubmitPrestasi(ctx context.Context, userID uuid.UUID, req mongodb.Achievement) (*model.AchievementReference, error)
}

type achievementService struct {
	repo repository.AchievementRepository
}

func NewAchievementService(repo repository.AchievementRepository) AchievementService {
	return &achievementService{repo: repo}
}

func (s *achievementService) SubmitPrestasi(ctx context.Context, userID uuid.UUID, req mongodb.Achievement) (*model.AchievementReference, error) {
	// 1. Cari data Student berdasarkan User ID yang login
	student, err := s.repo.GetStudentByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("student data not found for this user")
	}

	// 2. Setup Data untuk MongoDB
	req.ID = primitive.NewObjectID()
	req.StudentID = student.ID // Link ke UUID Student di Postgres
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()
	
	if req.CustomFields == nil {
		req.CustomFields = make(map[string]interface{})
	}

	// 3. Simpan ke MongoDB
	mongoID, err := s.repo.SaveAchievementMongo(ctx, req)
	if err != nil {
		return nil, err
	}

	// 4. Setup Data untuk Postgres (Reference)
	// Sesuai SRS Flow 4: Status awal 'draft'
	ref := model.AchievementReference{
		ID:                 uuid.New(),
		StudentID:          student.ID,
		MongoAchievementID: mongoID,
		Status:             "draft",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// 5. Simpan ke Postgres
	err = s.repo.SaveAchievementReference(ctx, ref)
	if err != nil {
		return nil, err
	}

	return &ref, nil
}