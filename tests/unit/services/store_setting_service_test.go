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

func setupStoreSettingTest(t *testing.T) (*mocks.StoreSettingRepository, services.StoreSettingService) {
	mockRepo := mocks.NewStoreSettingRepository(t)
	service := services.NewStoreSettingService(mockRepo)
	return mockRepo, service
}

// --- GetSettings ---

func TestStoreSettingService_Get_Success(t *testing.T) {
	mockRepo, service := setupStoreSettingTest(t)
	ctx := context.Background()

	expected := &models.StoreSetting{
		ID:         1,
		StoreName:  "Toko Serba Ada",
		Address:    "Jl. Merdeka No. 45",
		Phone:      "081234567890",
		FooterText: "Terima kasih!",
	}
	mockRepo.On("GetSettings", ctx).Return(expected, nil).Once()

	settings, err := service.GetSettings(ctx)

	assert.NoError(t, err)
	assert.Equal(t, expected, settings)
	assert.Equal(t, "Toko Serba Ada", settings.StoreName)
}

func TestStoreSettingService_Get_NotFound(t *testing.T) {
	mockRepo, service := setupStoreSettingTest(t)
	ctx := context.Background()

	mockRepo.On("GetSettings", ctx).Return(nil, errors.New("record not found")).Once()

	settings, err := service.GetSettings(ctx)

	assert.Error(t, err)
	assert.Nil(t, settings)
}

// --- UpdateSettings ---

func TestStoreSettingService_Update_Success(t *testing.T) {
	mockRepo, service := setupStoreSettingTest(t)
	ctx := context.Background()

	input := &models.StoreSetting{
		StoreName:  "New Store Name",
		Address:    "New Address",
		Phone:      "089999999999",
		FooterText: "New Footer",
	}
	expected := &models.StoreSetting{
		ID:         1,
		StoreName:  "New Store Name",
		Address:    "New Address",
		Phone:      "089999999999",
		FooterText: "New Footer",
	}
	mockRepo.On("UpsertSettings", ctx, mock.AnythingOfType("*models.StoreSetting")).Return(expected, nil).Once()

	settings, err := service.UpdateSettings(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, settings)
	assert.Equal(t, "New Store Name", settings.StoreName)
}

func TestStoreSettingService_Update_Error(t *testing.T) {
	mockRepo, service := setupStoreSettingTest(t)
	ctx := context.Background()

	input := &models.StoreSetting{StoreName: "Test"}
	mockRepo.On("UpsertSettings", ctx, mock.AnythingOfType("*models.StoreSetting")).
		Return(nil, errors.New("db error")).Once()

	settings, err := service.UpdateSettings(ctx, input)

	assert.Error(t, err)
	assert.Nil(t, settings)
}
