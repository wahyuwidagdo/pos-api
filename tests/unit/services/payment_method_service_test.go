package services_test

import (
	"context"
	"errors"
	"testing"

	"pos-api/internal/models"
	"pos-api/internal/services"
	"pos-api/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupPaymentMethodTest(t *testing.T) (*mocks.PaymentMethodRepository, services.PaymentMethodService) {
	mockRepo := mocks.NewPaymentMethodRepository(t)
	service := services.NewPaymentMethodService(mockRepo)
	return mockRepo, service
}

// --- Create ---

func TestPaymentMethodService_Create_Success(t *testing.T) {
	mockRepo, service := setupPaymentMethodTest(t)
	ctx := context.Background()

	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.PaymentMethod")).Return(nil).Once()

	pm, err := service.Create(ctx, services.CreatePaymentMethodRequest{
		Name:     "Cash",
		IsCash:   true,
		IsActive: true,
	})

	assert.NoError(t, err)
	assert.NotNil(t, pm)
	assert.Equal(t, "Cash", pm.Name)
	assert.True(t, pm.IsCash)
}

func TestPaymentMethodService_Create_EmptyName(t *testing.T) {
	_, service := setupPaymentMethodTest(t)
	ctx := context.Background()

	pm, err := service.Create(ctx, services.CreatePaymentMethodRequest{
		Name: "",
	})

	assert.Error(t, err)
	assert.Nil(t, pm)
	assert.Contains(t, err.Error(), "name is required")
}

func TestPaymentMethodService_Create_RepositoryError(t *testing.T) {
	mockRepo, service := setupPaymentMethodTest(t)
	ctx := context.Background()

	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.PaymentMethod")).
		Return(errors.New("duplicate key")).Once()

	pm, err := service.Create(ctx, services.CreatePaymentMethodRequest{
		Name: "QRIS",
	})

	assert.Error(t, err)
	assert.Nil(t, pm)
}

// --- Update ---

func TestPaymentMethodService_Update_Success(t *testing.T) {
	mockRepo, service := setupPaymentMethodTest(t)
	ctx := context.Background()

	existing := &models.PaymentMethod{ID: 1, Name: "Cash", IsCash: true, IsActive: true}
	mockRepo.On("GetByID", ctx, uint(1)).Return(existing, nil).Once()
	mockRepo.On("Update", ctx, mock.AnythingOfType("*models.PaymentMethod")).Return(nil).Once()

	isActive := false
	pm, err := service.Update(ctx, 1, services.UpdatePaymentMethodRequest{
		IsActive: &isActive,
	})

	assert.NoError(t, err)
	assert.NotNil(t, pm)
	assert.False(t, pm.IsActive) // Deactivated
}

func TestPaymentMethodService_Update_NotFound(t *testing.T) {
	mockRepo, service := setupPaymentMethodTest(t)
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, uint(999)).Return(nil, errors.New("not found")).Once()

	pm, err := service.Update(ctx, 999, services.UpdatePaymentMethodRequest{Name: "Whatever"})

	assert.Error(t, err)
	assert.Nil(t, pm)
}

// --- Delete ---

func TestPaymentMethodService_Delete_Success(t *testing.T) {
	mockRepo, service := setupPaymentMethodTest(t)
	ctx := context.Background()

	existing := &models.PaymentMethod{ID: 1}
	mockRepo.On("GetByID", ctx, uint(1)).Return(existing, nil).Once()
	mockRepo.On("Delete", ctx, uint(1)).Return(nil).Once()

	err := service.Delete(ctx, 1)

	assert.NoError(t, err)
}

func TestPaymentMethodService_Delete_NotFound(t *testing.T) {
	mockRepo, service := setupPaymentMethodTest(t)
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, uint(999)).Return(nil, errors.New("not found")).Once()

	err := service.Delete(ctx, 999)

	assert.Error(t, err)
}

// --- GetAll ---

func TestPaymentMethodService_GetAll_Success(t *testing.T) {
	mockRepo, service := setupPaymentMethodTest(t)
	ctx := context.Background()

	methods := []models.PaymentMethod{
		{ID: 1, Name: "Cash", IsCash: true, IsActive: true},
		{ID: 2, Name: "QRIS", IsCash: false, IsActive: true},
		{ID: 3, Name: "Transfer BCA", IsCash: false, IsActive: false},
	}
	mockRepo.On("GetAll", ctx).Return(methods, nil).Once()

	result, err := service.GetAll(ctx)

	assert.NoError(t, err)
	assert.Len(t, result, 3)
}

// --- GetActive ---

func TestPaymentMethodService_GetActive_Success(t *testing.T) {
	mockRepo, service := setupPaymentMethodTest(t)
	ctx := context.Background()

	activeMethods := []models.PaymentMethod{
		{ID: 1, Name: "Cash", IsActive: true},
		{ID: 2, Name: "QRIS", IsActive: true},
	}
	mockRepo.On("GetActive", ctx).Return(activeMethods, nil).Once()

	result, err := service.GetActive(ctx)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	for _, pm := range result {
		assert.True(t, pm.IsActive) // All should be active
	}
}
