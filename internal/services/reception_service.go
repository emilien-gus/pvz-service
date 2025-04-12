package services

import (
	"context"
	"pvz/internal/models"
	"pvz/internal/repository"

	"github.com/google/uuid"
)

type ReceptionService struct {
	receptionRepo repository.ReceptionRepositoryInterface
}

func NEwReceptionService(receptionRepo repository.ReceptionRepositoryInterface) *ReceptionService {
	return &ReceptionService{receptionRepo: receptionRepo}
}

func (s *ReceptionService) CreateReception(ctx context.Context, pvzID uuid.UUID, role string) (models.Reception, error) {
	reception := &models.Reception{}
	if role != "employee" {
		return *reception, ErrAccessDenied
	}

	reception, err := s.receptionRepo.InsertReception(ctx, pvzID)
	if err != nil {
		return *reception, err
	}

	return *reception, nil
}
