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

type MockPVZRepository struct {
	mock.Mock
}

func (m *MockPVZRepository) InsertPVZ(ctx context.Context, city string) (*models.PVZ, error) {
	args := m.Called(ctx, city)
	return args.Get(0).(*models.PVZ), args.Error(1)
}

func (m *MockPVZRepository) GetPVZList(ctx context.Context, startDate, endDate *time.Time, page, limit int) ([]models.PVZWithReceptions, error) {
	args := m.Called(ctx, startDate, endDate, page, limit)
	return args.Get(0).([]models.PVZWithReceptions), args.Error(1)
}

func TestPVZService_CreatePVZ(t *testing.T) {
	mockRepo := new(MockPVZRepository)
	pvzService := NewPVZService(mockRepo)

	t.Run("successful PVZ creation", func(t *testing.T) {
		mockRepo.On("InsertPVZ", mock.Anything, "Москва").Return(&models.PVZ{ID: uuid.New(), Registration: time.Now(), City: "Москва"}, nil)

		pvz, err := pvzService.CreatePVZ(context.Background(), "Москва", "moderator")

		assert.NoError(t, err)
		assert.NotNil(t, pvz)
		assert.Equal(t, "Москва", pvz.City)
	})

	t.Run("access denied for non-moderator", func(t *testing.T) {
		pvz, err := pvzService.CreatePVZ(context.Background(), "Москва", "employee")

		assert.Error(t, err)
		assert.Equal(t, ErrAccessDenied, err)
		assert.Empty(t, pvz)
	})

	t.Run("not allowed city", func(t *testing.T) {
		pvz, err := pvzService.CreatePVZ(context.Background(), "Лондон", "moderator")

		assert.Error(t, err)
		assert.Equal(t, "not allowed city", err.Error())
		assert.Empty(t, pvz)
	})

	t.Run("error while inserting PVZ", func(t *testing.T) {
		mockRepo.ExpectedCalls = []*mock.Call{}
		mockRepo.On("InsertPVZ", mock.Anything, "Москва").Return(&models.PVZ{}, errors.New("database error"))

		pvz, err := pvzService.CreatePVZ(context.Background(), "Москва", "moderator")

		assert.Error(t, err)
		assert.Empty(t, pvz)
		mockRepo.AssertExpectations(t)
	})
}

func TestPVZService_GetPVZList(t *testing.T) {
	mockRepo := new(MockPVZRepository)
	pvzService := NewPVZService(mockRepo)

	t.Run("successful get PVZ list", func(t *testing.T) {
		startDate := time.Now().Add(-24 * time.Hour)
		endDate := time.Now()
		mockRepo.On("GetPVZList", mock.Anything, &startDate, &endDate, 1, 10).Return([]models.PVZWithReceptions{{ID: uuid.New(), City: "Москва"}}, nil)

		pvzList, err := pvzService.GetPVZList(context.Background(), &startDate, &endDate, 1, 10, "employee")

		assert.NoError(t, err)
		assert.Len(t, pvzList, 1)
		assert.Equal(t, "Москва", pvzList[0].City)
		mockRepo.AssertExpectations(t)
	})

	t.Run("access denied for non-employee", func(t *testing.T) {
		pvzList, err := pvzService.GetPVZList(context.Background(), nil, nil, 1, 10, "moderator")

		assert.Error(t, err)
		assert.Equal(t, ErrAccessDenied, err)
		assert.Nil(t, pvzList)
	})

	t.Run("invalid page parameter", func(t *testing.T) {
		pvzList, err := pvzService.GetPVZList(context.Background(), nil, nil, -1, 10, "employee")

		assert.Error(t, err)
		assert.Equal(t, ErrPageParamIsInvalid, err)
		assert.Nil(t, pvzList)
	})

	t.Run("invalid limit parameter", func(t *testing.T) {
		pvzList, err := pvzService.GetPVZList(context.Background(), nil, nil, 1, 31, "employee")

		assert.Error(t, err)
		assert.Equal(t, ErrLimitParamIsInvalid, err)
		assert.Nil(t, pvzList)
	})

	t.Run("start date after end date", func(t *testing.T) {
		startDate := time.Now().Add(24 * time.Hour)
		endDate := time.Now()

		pvzList, err := pvzService.GetPVZList(context.Background(), &startDate, &endDate, 1, 10, "employee")

		assert.Error(t, err)
		assert.Equal(t, ErrStartLaterThenEnd, err)
		assert.Nil(t, pvzList)
	})

	t.Run("error while getting PVZ list", func(t *testing.T) {
		mockRepo.ExpectedCalls = []*mock.Call{}
		mockRepo.On("GetPVZList", mock.Anything, mock.Anything, mock.Anything, 1, 10).Return([]models.PVZWithReceptions{}, errors.New("database error"))

		pvzList, err := pvzService.GetPVZList(context.Background(), nil, nil, 1, 10, "employee")

		assert.Error(t, err)
		assert.Equal(t, "database error", err.Error())
		assert.Empty(t, pvzList)
	})
}
