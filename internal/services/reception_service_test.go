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

type MockReceptionRepository struct {
	mock.Mock
}

func (m *MockReceptionRepository) InsertReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	args := m.Called(ctx, pvzID)
	return args.Get(0).(*models.Reception), args.Error(1)
}

func (m *MockReceptionRepository) UpdateLastReceptionStatus(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	args := m.Called(ctx, pvzID)
	return args.Get(0).(*models.Reception), args.Error(1)
}

func TestReceptionService_CreateReception(t *testing.T) {
	mockRepo := new(MockReceptionRepository)
	receptionService := NewReceptionService(mockRepo)

	pvzID := uuid.New()
	role := "employee"

	t.Run("successful reception creation", func(t *testing.T) {
		expectedReception := &models.Reception{
			ID:       uuid.New(),
			DateTime: time.Now(),
			PVZID:    pvzID,
			Status:   models.ReceptionStatusInProgress,
		}
		mockRepo.On("InsertReception", mock.Anything, pvzID).Return(expectedReception, nil)

		reception, err := receptionService.CreateReception(context.Background(), pvzID, role)

		assert.NoError(t, err)
		assert.Equal(t, *expectedReception, reception)
	})

	t.Run("access denied for non-employee role", func(t *testing.T) {
		role = "admin"

		reception, err := receptionService.CreateReception(context.Background(), pvzID, role)

		assert.Error(t, err)
		assert.Equal(t, ErrAccessDenied, err)
		assert.Empty(t, reception)
	})

	t.Run("error while creating reception", func(t *testing.T) {
		role = "employee"
		mockRepo.ExpectedCalls = []*mock.Call{}
		mockRepo.On("InsertReception", mock.Anything, pvzID).Return(&models.Reception{}, errors.New("some error"))

		reception, err := receptionService.CreateReception(context.Background(), pvzID, role)

		assert.Error(t, err)
		assert.Equal(t, "some error", err.Error())
		assert.Empty(t, reception)
	})
}

func TestReceptionService_CloseReception(t *testing.T) {
	mockRepo := new(MockReceptionRepository)
	receptionService := NewReceptionService(mockRepo)

	pvzID := uuid.New()
	role := "employee"

	t.Run("successful reception close", func(t *testing.T) {
		expectedReception := &models.Reception{
			ID:       uuid.New(),
			DateTime: time.Now(),
			PVZID:    pvzID,
			Status:   models.ReceptionStatusInProgress,
		}
		mockRepo.On("UpdateLastReceptionStatus", mock.Anything, pvzID).Return(expectedReception, nil)

		reception, err := receptionService.CloseReception(context.Background(), pvzID, role)

		assert.NoError(t, err)
		assert.Equal(t, *expectedReception, reception)
	})

	t.Run("access denied for non-employee role", func(t *testing.T) {
		role = "admin"

		reception, err := receptionService.CloseReception(context.Background(), pvzID, role)

		assert.Error(t, err)
		assert.Equal(t, ErrAccessDenied, err)
		assert.Empty(t, reception)
	})

	t.Run("error while closing reception", func(t *testing.T) {
		mockRepo.ExpectedCalls = []*mock.Call{}
		role = "employee"
		mockRepo.On("UpdateLastReceptionStatus", mock.Anything, pvzID).Return(&models.Reception{}, errors.New("some error"))

		reception, err := receptionService.CloseReception(context.Background(), pvzID, role)

		assert.Error(t, err)
		assert.Equal(t, "some error", err.Error())
		assert.Empty(t, reception)
	})
}
