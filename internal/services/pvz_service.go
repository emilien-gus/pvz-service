package services

import (
	"context"
	"errors"
	"pvz/internal/models"
	"pvz/internal/repository"
	"time"
)

var allowedCities = map[string]bool{
	"Москва":          true,
	"Санкт-Петербург": true,
	"Казань":          true,
}

type PVZServiceInterface interface {
	CreatePVZ(ctx context.Context, city, role string) (models.PVZ, error)
	GetPVZList(ctx context.Context, startDate, endDate *time.Time, page, limit int, role string) ([]models.PVZWithReceptions, error)
}

type PVZService struct {
	pvzRepo repository.PVZRepositoryInterface
}

func NewPVZService(pvzRepo repository.PVZRepositoryInterface) *PVZService {
	return &PVZService{pvzRepo: pvzRepo}
}

func (s *PVZService) CreatePVZ(ctx context.Context, city, role string) (models.PVZ, error) {
	if role != "moderator" {
		return models.PVZ{}, ErrAccessDenied
	}

	if _, ok := allowedCities[city]; !ok {
		return models.PVZ{}, errors.New("not allowed city")
	}

	pvz, err := s.pvzRepo.InsertPVZ(ctx, city)
	if err != nil {
		return models.PVZ{}, err
	}

	return *pvz, err
}

func (s *PVZService) GetPVZList(ctx context.Context, startDate, endDate *time.Time, page, limit int, role string) ([]models.PVZWithReceptions, error) {
	if role != "employee" {
		return nil, ErrAccessDenied
	}

	if page < 1 {
		return nil, ErrPageParamIsInvalid
	}

	if limit <= 0 || limit > 30 {
		return nil, ErrLimitParamIsInvalid
	}

	if startDate != nil && endDate != nil && startDate.After(*endDate) {
		return nil, ErrStartLaterThenEnd
	}

	arr, err := s.pvzRepo.GetPVZList(ctx, startDate, endDate, page, limit)
	return arr, err
}
