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

func setupInventoryTest(t *testing.T) (*mocks.InventoryLogRepository, *mocks.ProductRepository, services.InventoryLogService) {
	mockLogRepo := mocks.NewInventoryLogRepository(t)
	mockProductRepo := mocks.NewProductRepository(t)
	service := services.NewInventoryLogService(mockLogRepo, mockProductRepo)
	return mockLogRepo, mockProductRepo, service
}

// --- AdjustStock: Stock In ---

func TestInventoryService_AdjustStock_In_Success(t *testing.T) {
	mockLogRepo, mockProductRepo, service := setupInventoryTest(t)
	ctx := context.Background()

	product := &models.Product{ID: 1, Name: "Test Product", Stock: 10, Cost: 5000}
	mockProductRepo.On("GetProductByID", ctx, uint(1)).Return(product, nil).Once()
	mockLogRepo.On("ProcessAdjustment", ctx, mock.AnythingOfType("*models.InventoryLog"), mock.AnythingOfType("*models.Product")).Return(nil).Once()

	log, err := service.AdjustStock(ctx, services.StockAdjustmentRequest{
		ProductID: 1,
		Type:      "in",
		Source:    "purchase",
		Quantity:  20,
		CostPrice: 5500,
	}, 1)

	assert.NoError(t, err)
	assert.NotNil(t, log)
	assert.Equal(t, 10, log.StockBefore)
	assert.Equal(t, 30, log.StockAfter)
	assert.Equal(t, "in", log.Type)
}

// --- AdjustStock: Stock Out ---

func TestInventoryService_AdjustStock_Out_Success(t *testing.T) {
	mockLogRepo, mockProductRepo, service := setupInventoryTest(t)
	ctx := context.Background()

	product := &models.Product{ID: 1, Name: "Test", Stock: 10, Cost: 3000}
	mockProductRepo.On("GetProductByID", ctx, uint(1)).Return(product, nil).Once()
	mockLogRepo.On("ProcessAdjustment", ctx, mock.AnythingOfType("*models.InventoryLog"), mock.AnythingOfType("*models.Product")).Return(nil).Once()

	log, err := service.AdjustStock(ctx, services.StockAdjustmentRequest{
		ProductID: 1,
		Type:      "out",
		Source:    "damage",
		Quantity:  3,
	}, 1)

	assert.NoError(t, err)
	assert.NotNil(t, log)
	assert.Equal(t, 10, log.StockBefore)
	assert.Equal(t, 7, log.StockAfter)
}

func TestInventoryService_AdjustStock_Out_InsufficientStock(t *testing.T) {
	_, mockProductRepo, service := setupInventoryTest(t)
	ctx := context.Background()

	product := &models.Product{ID: 1, Name: "Test", Stock: 5, Cost: 3000}
	mockProductRepo.On("GetProductByID", ctx, uint(1)).Return(product, nil).Once()

	log, err := service.AdjustStock(ctx, services.StockAdjustmentRequest{
		ProductID: 1,
		Type:      "out",
		Source:    "damage",
		Quantity:  10, // More than stock
	}, 1)

	assert.Error(t, err)
	assert.Nil(t, log)
	assert.Contains(t, err.Error(), "insufficient stock")
}

// --- AdjustStock: Adjustment ---

func TestInventoryService_AdjustStock_Adjustment_Success(t *testing.T) {
	mockLogRepo, mockProductRepo, service := setupInventoryTest(t)
	ctx := context.Background()

	product := &models.Product{ID: 1, Name: "Test", Stock: 10, Cost: 3000}
	mockProductRepo.On("GetProductByID", ctx, uint(1)).Return(product, nil).Once()
	mockLogRepo.On("ProcessAdjustment", ctx, mock.AnythingOfType("*models.InventoryLog"), mock.AnythingOfType("*models.Product")).Return(nil).Once()

	log, err := service.AdjustStock(ctx, services.StockAdjustmentRequest{
		ProductID: 1,
		Type:      "adjustment",
		Source:    "opname",
		Quantity:  25, // New absolute stock level
	}, 1)

	assert.NoError(t, err)
	assert.NotNil(t, log)
	assert.Equal(t, 10, log.StockBefore)
	assert.Equal(t, 25, log.StockAfter)
}

// --- AdjustStock: Product Not Found ---

func TestInventoryService_AdjustStock_ProductNotFound(t *testing.T) {
	_, mockProductRepo, service := setupInventoryTest(t)
	ctx := context.Background()

	mockProductRepo.On("GetProductByID", ctx, uint(999)).Return(nil, errors.New("not found")).Once()

	log, err := service.AdjustStock(ctx, services.StockAdjustmentRequest{
		ProductID: 999,
		Type:      "in",
		Source:    "purchase",
		Quantity:  10,
	}, 1)

	assert.Error(t, err)
	assert.Nil(t, log)
	assert.Contains(t, err.Error(), "not found")
}

// --- GetLogsByProduct ---

func TestInventoryService_GetLogsByProduct_Success(t *testing.T) {
	mockLogRepo, _, service := setupInventoryTest(t)
	ctx := context.Background()

	logs := []models.InventoryLog{
		{ID: 1, ProductID: 1, Type: "in", Quantity: 10},
		{ID: 2, ProductID: 1, Type: "out", Quantity: 2},
	}
	mockLogRepo.On("GetByProductID", ctx, uint(1), 20, 0).Return(logs, int64(2), nil).Once()

	result, total, err := service.GetLogsByProduct(ctx, 1, 1, 20)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, result, 2)
}

// --- GetAllLogs ---

func TestInventoryService_GetAllLogs_Success(t *testing.T) {
	mockLogRepo, _, service := setupInventoryTest(t)
	ctx := context.Background()

	logs := []models.InventoryLog{{ID: 1}, {ID: 2}, {ID: 3}}
	mockLogRepo.On("GetAll", ctx, 20, 0, "", "", (*interface{})(nil), (*interface{})(nil)).
		Return(logs, int64(3), nil).Maybe()
	// Using mock.Anything for the time pointers since they're nil
	mockLogRepo.On("GetAll", ctx, 20, 0, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(logs, int64(3), nil).Maybe()

	result, total, err := service.GetAllLogs(ctx, 1, 20, "", "", nil, nil)

	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, result, 3)
}

// --- GetInventoryStats ---

func TestInventoryService_GetStats_Success(t *testing.T) {
	mockLogRepo, _, service := setupInventoryTest(t)
	ctx := context.Background()

	stats := map[string]int64{"in": 50, "out": 10, "adjustment": 3}
	mockLogRepo.On("GetStats", ctx, mock.Anything, mock.Anything).Return(stats, nil).Once()

	result, err := service.GetInventoryStats(ctx, nil, nil)

	assert.NoError(t, err)
	assert.Equal(t, int64(50), result["in"])
	assert.Equal(t, int64(10), result["out"])
}
