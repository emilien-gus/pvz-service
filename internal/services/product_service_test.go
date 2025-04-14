package services

import (
	"context"
	"errors"
	"pvz/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) InsertProduct(ctx context.Context, productType string, pvzID uuid.UUID) (*models.Product, error) {
	args := m.Called(ctx, productType, pvzID)
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepository) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
	args := m.Called(ctx, pvzID)
	return args.Error(0)
}

func TestProductService_AddProduct(t *testing.T) {
	mockRepo := new(MockProductRepository)
	productService := NewProductService(mockRepo)

	pvzID := uuid.New()
	receptionID := uuid.New()
	productType := "электроника"
	role := "employee"

	t.Run("successful product addition", func(t *testing.T) {

		product := &models.Product{
			ID:          uuid.New(),
			DateTime:    time.Now(),
			ProductType: productType,
			ReceptionID: receptionID,
		}

		mockRepo.On("InsertProduct", mock.Anything, productType, pvzID).Return(product, nil)

		outProduct, err := productService.AddProduct(context.Background(), productType, pvzID, role)

		assert.NoError(t, err)
		assert.NotEmpty(t, outProduct.ID)
	})

	t.Run("access denied for non-employee role", func(t *testing.T) {
		role = "admin"

		product, err := productService.AddProduct(context.Background(), productType, pvzID, role)

		assert.Error(t, err)
		assert.Equal(t, ErrAccessDenied, err)
		assert.Empty(t, product.ID)
	})

	t.Run("product type not allowed", func(t *testing.T) {
		role = "employee"
		productType = "неизвестный тип"

		product, err := productService.AddProduct(context.Background(), productType, pvzID, role)

		assert.Error(t, err)
		assert.Equal(t, ErrProductTypeNotAllowed, err)
		assert.Empty(t, product.ID)
	})
}

func TestProductService_DeleteProduct(t *testing.T) {
	mockRepo := new(MockProductRepository)
	productService := NewProductService(mockRepo)

	pvzID := uuid.New()
	role := "employee"

	t.Run("successful product deletion", func(t *testing.T) {
		// Mock the DeleteLastProduct method
		mockRepo.On("DeleteLastProduct", mock.Anything, pvzID).Return(nil)

		err := productService.DeleteProduct(context.Background(), pvzID, role)

		assert.NoError(t, err)
	})

	t.Run("error while deleting product", func(t *testing.T) {
		mockRepo.ExpectedCalls = []*mock.Call{}

		mockRepo.On("DeleteLastProduct", mock.Anything, pvzID).Return(errors.New("some error"))

		err := productService.DeleteProduct(context.Background(), pvzID, role)

		assert.Error(t, err)
		assert.Equal(t, "some error", err.Error())
	})

	t.Run("access denied for non-employee role", func(t *testing.T) {
		role = "admin"
		mockRepo.On("DeleteLastProduct", mock.Anything, pvzID).Return(nil)

		err := productService.DeleteProduct(context.Background(), pvzID, role)

		assert.Error(t, err)
		assert.Equal(t, ErrAccessDenied, err)
	})

}
