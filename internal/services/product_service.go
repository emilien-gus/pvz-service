package services

import (
	"context"
	"pvz/internal/models"
	"pvz/internal/repository"

	"github.com/google/uuid"
)

var allowedProductTypes = map[string]bool{
	"электроника": true,
	"одежда":      true,
	"обувь":       true,
}

type ProductServiceInterface interface {
	AddProduct(ctx context.Context, productType string, pvzID uuid.UUID, role string) (models.Product, error)
	DeleteProduct(ctx context.Context, pvzID uuid.UUID, role string) error
}

type ProductService struct {
	productRepo repository.ProductRepositoryInterface
}

func NewProductService(productRepo repository.ProductRepositoryInterface) *ProductService {
	return &ProductService{productRepo: productRepo}
}

func (s *ProductService) AddProduct(ctx context.Context, productType string, pvzID uuid.UUID, role string) (models.Product, error) {
	product := &models.Product{}
	if role != "employee" {
		return *product, ErrAccessDenied
	}

	if _, ok := allowedProductTypes[productType]; !ok {
		return *product, ErrProductTypeNotAllowed
	}

	product, err := s.productRepo.InsertProduct(ctx, productType, pvzID)
	if err != nil {
		return *product, err
	}

	return *product, err
}

func (s *ProductService) DeleteProduct(ctx context.Context, pvzID uuid.UUID, role string) error {
	if role != "employee" {
		return ErrAccessDenied
	}

	err := s.productRepo.DeleteLastProduct(ctx, pvzID)
	if err != nil {
		return err
	}

	return nil
}
