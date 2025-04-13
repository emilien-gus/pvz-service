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

type PVZService struct {
	pvzRepo repository.PVZRepositoryInterface
}

func NewPVZService(pvzRepo repository.PVZRepositoryInterface) *PVZService {
	return &PVZService{pvzRepo: pvzRepo}
}

func (s *PVZService) CreatePVZ(ctx context.Context, city, role string) (models.PVZ, error) {
	pvz := &models.PVZ{}
	if role != "moderator" {
		return *pvz, ErrAccessDenied
	}

	if _, ok := allowedCities[city]; !ok {
		return *pvz, errors.New("not allowed city")
	}

	pvz, err := s.pvzRepo.InsertPVZ(ctx, city)
	if err != nil {
		return *pvz, err
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

	return s.pvzRepo.GetPVZList(ctx, startDate, endDate, page, limit)
}
