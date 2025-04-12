package services

import (
	"context"
	"errors"
	"pvz/internal/models"
	"pvz/internal/repository"
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
