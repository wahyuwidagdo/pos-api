package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"pos-api/internal/repositories"
	"pos-api/internal/services"
	"pos-api/tests/mocks"

	"github.com/stretchr/testify/assert"
)

func setupReportTest(t *testing.T) (*mocks.ReportRepository, services.ReportService) {
	mockRepo := mocks.NewReportRepository(t)
	service := services.NewReportService(mockRepo)
	return mockRepo, service
}

// --- GetSalesReport ---

func TestReportService_GetSalesReport_Success(t *testing.T) {
	mockRepo, service := setupReportTest(t)
	ctx := context.Background()

	startDate, _ := time.Parse("2006-01-02", "2026-01-01")
	endDate, _ := time.Parse("2006-01-02", "2026-01-31")

	mockRepo.On("GetSalesSummary", ctx, startDate, endDate).Return(&repositories.SalesSummary{
		TotalSales:        5000000,
		TotalTransactions: 100,
	}, nil).Once()

	mockRepo.On("GetSalesReport", ctx, startDate, endDate).Return([]repositories.SalesReport{
		{Date: "2026-01-01", TotalSales: 200000},
	}, nil).Once()

	mockRepo.On("GetSalesByHour", ctx, startDate, endDate).Return([]repositories.HourlySales{}, nil).Once()

	report, err := service.GetSalesReport(ctx, "2026-01-01", "2026-01-31")

	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, float64(5000000), report.Summary.TotalSales)
	assert.Len(t, report.DailyData, 1)
}

func TestReportService_GetSalesReport_InvalidStartDate(t *testing.T) {
	_, service := setupReportTest(t)
	ctx := context.Background()

	report, err := service.GetSalesReport(ctx, "invalid-date", "2026-01-31")

	assert.Error(t, err)
	assert.Nil(t, report)
	assert.Contains(t, err.Error(), "tanggal mulai tidak valid")
}

func TestReportService_GetSalesReport_InvalidEndDate(t *testing.T) {
	_, service := setupReportTest(t)
	ctx := context.Background()

	report, err := service.GetSalesReport(ctx, "2026-01-01", "invalid-date")

	assert.Error(t, err)
	assert.Nil(t, report)
	assert.Contains(t, err.Error(), "tanggal akhir tidak valid")
}

func TestReportService_GetSalesReport_EndBeforeStart(t *testing.T) {
	_, service := setupReportTest(t)
	ctx := context.Background()

	report, err := service.GetSalesReport(ctx, "2026-01-31", "2026-01-01")

	assert.Error(t, err)
	assert.Nil(t, report)
	assert.Contains(t, err.Error(), "setelah tanggal mulai")
}

// --- GetProductReport ---

func TestReportService_GetProductReport_Success(t *testing.T) {
	mockRepo, service := setupReportTest(t)
	ctx := context.Background()

	startDate, _ := time.Parse("2006-01-02", "2026-01-01")
	endDate, _ := time.Parse("2006-01-02", "2026-01-31")

	mockRepo.On("GetProductReport", ctx, startDate, endDate, 20).Return([]repositories.ProductReport{
		{ProductName: "Mie Goreng", TotalSold: 50},
	}, nil).Once()

	report, err := service.GetProductReport(ctx, "2026-01-01", "2026-01-31", 0) // limit=0 defaults to 20

	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Len(t, report.Products, 1)
}

func TestReportService_GetProductReport_InvalidDate(t *testing.T) {
	_, service := setupReportTest(t)
	ctx := context.Background()

	report, err := service.GetProductReport(ctx, "bad", "2026-01-31", 10)

	assert.Error(t, err)
	assert.Nil(t, report)
}

// --- GetStockValue ---

func TestReportService_GetStockValue_Success(t *testing.T) {
	mockRepo, service := setupReportTest(t)
	ctx := context.Background()

	mockRepo.On("GetStockValue", ctx).Return(&repositories.StockValue{
		TotalProducts: 50,
		TotalUnits:    500,
		TotalValue:    15000000,
		TotalRetail:   25000000,
	}, nil).Once()

	result, err := service.GetStockValue(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(500), result.TotalUnits)
	assert.Equal(t, float64(15000000), result.TotalValue)
}

func TestReportService_GetStockValue_Error(t *testing.T) {
	mockRepo, service := setupReportTest(t)
	ctx := context.Background()

	mockRepo.On("GetStockValue", ctx).Return(nil, errors.New("db error")).Once()

	result, err := service.GetStockValue(ctx)

	assert.Error(t, err)
	assert.Nil(t, result)
}
