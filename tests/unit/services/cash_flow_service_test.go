package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"pos-api/internal/models"
	"pos-api/internal/services"
	"pos-api/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupCashFlowTest(t *testing.T) (*mocks.CashFlowRepository, services.CashFlowService) {
	mockRepo := mocks.NewCashFlowRepository(t)
	service := services.NewCashFlowService(mockRepo)
	return mockRepo, service
}

// --- Create ---

func TestCashFlowService_Create_Income_Success(t *testing.T) {
	mockRepo, service := setupCashFlowTest(t)
	ctx := context.Background()

	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.CashFlow")).Return(nil).Once()

	cf, err := service.Create(ctx, services.CreateCashFlowRequest{
		Type:   "income",
		Source: "modal_awal",
		Amount: 1000000,
		Date:   "2026-01-15",
		Notes:  "Modal awal toko",
	}, 1)

	assert.NoError(t, err)
	assert.NotNil(t, cf)
	assert.Equal(t, "income", cf.Type)
	assert.Equal(t, float64(1000000), cf.Amount)
}

func TestCashFlowService_Create_Expense_Success(t *testing.T) {
	mockRepo, service := setupCashFlowTest(t)
	ctx := context.Background()

	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.CashFlow")).Return(nil).Once()

	cf, err := service.Create(ctx, services.CreateCashFlowRequest{
		Type:   "expense",
		Source: "electricity",
		Amount: 200000,
		Date:   "2026-01-20",
		Notes:  "Bayar listrik bulan Januari",
	}, 1)

	assert.NoError(t, err)
	assert.NotNil(t, cf)
	assert.Equal(t, "expense", cf.Type)
}

func TestCashFlowService_Create_InvalidType(t *testing.T) {
	_, service := setupCashFlowTest(t)
	ctx := context.Background()

	cf, err := service.Create(ctx, services.CreateCashFlowRequest{
		Type:   "invalid",
		Source: "test",
		Amount: 100,
		Date:   "2026-01-01",
	}, 1)

	assert.Error(t, err)
	assert.Nil(t, cf)
	assert.Contains(t, err.Error(), "income")
}

func TestCashFlowService_Create_InvalidDateFormat(t *testing.T) {
	_, service := setupCashFlowTest(t)
	ctx := context.Background()

	cf, err := service.Create(ctx, services.CreateCashFlowRequest{
		Type:   "income",
		Source: "test",
		Amount: 100,
		Date:   "01-20-2026", // Wrong format
	}, 1)

	assert.Error(t, err)
	assert.Nil(t, cf)
	assert.Contains(t, err.Error(), "YYYY-MM-DD")
}

// --- Update ---

func TestCashFlowService_Update_Success(t *testing.T) {
	mockRepo, service := setupCashFlowTest(t)
	ctx := context.Background()

	existing := &models.CashFlow{ID: 1, Type: "income", Source: "modal_awal", Amount: 1000000}
	mockRepo.On("GetByID", ctx, uint(1)).Return(existing, nil).Once()
	mockRepo.On("Update", ctx, mock.AnythingOfType("*models.CashFlow")).Return(nil).Once()

	cf, err := service.Update(ctx, 1, services.UpdateCashFlowRequest{
		Amount: 1500000,
		Notes:  "Updated amount",
	})

	assert.NoError(t, err)
	assert.Equal(t, float64(1500000), cf.Amount)
}

func TestCashFlowService_Update_NotFound(t *testing.T) {
	mockRepo, service := setupCashFlowTest(t)
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, uint(999)).Return(nil, errors.New("not found")).Once()

	cf, err := service.Update(ctx, 999, services.UpdateCashFlowRequest{Amount: 100})

	assert.Error(t, err)
	assert.Nil(t, cf)
	assert.Contains(t, err.Error(), "not found")
}

// --- Delete ---

func TestCashFlowService_Delete_Success(t *testing.T) {
	mockRepo, service := setupCashFlowTest(t)
	ctx := context.Background()

	existing := &models.CashFlow{ID: 1}
	mockRepo.On("GetByID", ctx, uint(1)).Return(existing, nil).Once()
	mockRepo.On("Delete", ctx, uint(1)).Return(nil).Once()

	err := service.Delete(ctx, 1)

	assert.NoError(t, err)
}

func TestCashFlowService_Delete_NotFound(t *testing.T) {
	mockRepo, service := setupCashFlowTest(t)
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, uint(999)).Return(nil, errors.New("not found")).Once()

	err := service.Delete(ctx, 999)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// --- GetByID ---

func TestCashFlowService_GetByID_Success(t *testing.T) {
	mockRepo, service := setupCashFlowTest(t)
	ctx := context.Background()

	expected := &models.CashFlow{ID: 1, Type: "income", Amount: 500000}
	mockRepo.On("GetByID", ctx, uint(1)).Return(expected, nil).Once()

	cf, err := service.GetByID(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, expected, cf)
}

// --- GetAll ---

func TestCashFlowService_GetAll_Success(t *testing.T) {
	mockRepo, service := setupCashFlowTest(t)
	ctx := context.Background()

	flows := []models.CashFlow{{ID: 1}, {ID: 2}}
	mockRepo.On("GetAll", ctx, 20, 0, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(flows, int64(2), nil).Once()

	result, total, err := service.GetAll(ctx, 1, 20, "", "", nil, nil)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, result, 2)
}

func TestCashFlowService_GetAll_DefaultPageSize(t *testing.T) {
	mockRepo, service := setupCashFlowTest(t)
	ctx := context.Background()

	mockRepo.On("GetAll", ctx, 20, 0, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return([]models.CashFlow{}, int64(0), nil).Once()

	// pageSize=0 should default to 20
	_, _, err := service.GetAll(ctx, 1, 0, "", "", nil, nil)

	assert.NoError(t, err)
}

// --- GetSummary ---

func TestCashFlowService_GetSummary_Success(t *testing.T) {
	mockRepo, service := setupCashFlowTest(t)
	ctx := context.Background()

	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	// Returns: capital, income, expense, error
	mockRepo.On("GetSummary", ctx, startDate, endDate).
		Return(float64(1000000), float64(500000), float64(200000), nil).Once()

	summary, err := service.GetSummary(ctx, startDate, endDate)

	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, float64(1000000), summary.TotalCapital)
	assert.Equal(t, float64(500000), summary.TotalIncome)
	assert.Equal(t, float64(200000), summary.TotalExpense)
	assert.Equal(t, float64(300000), summary.NetProfit) // income - expense
}

func TestCashFlowService_GetSummary_Error(t *testing.T) {
	mockRepo, service := setupCashFlowTest(t)
	ctx := context.Background()

	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	mockRepo.On("GetSummary", ctx, startDate, endDate).
		Return(float64(0), float64(0), float64(0), errors.New("db error")).Once()

	summary, err := service.GetSummary(ctx, startDate, endDate)

	assert.Error(t, err)
	assert.Nil(t, summary)
}
