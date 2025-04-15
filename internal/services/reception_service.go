package services

import (
	"context"
	"pvz/internal/models"
	"pvz/internal/repository"

	"github.com/google/uuid"
)

type ReceptionServiceInterface interface {
	CreateReception(ctx context.Context, pvzID uuid.UUID, role string) (models.Reception, error)
	CloseReception(ctx context.Context, pvzID uuid.UUID, role string) (models.Reception, error)
}

type ReceptionService struct {
	receptionRepo repository.ReceptionRepositoryInterface
}

func NewReceptionService(receptionRepo repository.ReceptionRepositoryInterface) *ReceptionService {
	return &ReceptionService{receptionRepo: receptionRepo}
}

func (s *ReceptionService) CreateReception(ctx context.Context, pvzID uuid.UUID, role string) (models.Reception, error) {
	if role != "employee" {
		return models.Reception{}, ErrAccessDenied
	}

	reception, err := s.receptionRepo.InsertReception(ctx, pvzID)
	if err != nil {
		return models.Reception{}, err
	}

	return *reception, nil
}

func (s *ReceptionService) CloseReception(ctx context.Context, pvzID uuid.UUID, role string) (models.Reception, error) {
	if role != "employee" {
		return models.Reception{}, ErrAccessDenied
	}

	reception, err := s.receptionRepo.UpdateLastReceptionStatus(ctx, pvzID)
	if err != nil {
		return models.Reception{}, err
	}

	return *reception, nil
}
