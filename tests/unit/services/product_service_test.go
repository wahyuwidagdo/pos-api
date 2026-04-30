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
	"gorm.io/gorm"
)

func setupProductTest(t *testing.T) (*mocks.ProductRepository, services.ProductService) {
	mockRepo := mocks.NewProductRepository(t)
	service := services.NewProductService(mockRepo)
	return mockRepo, service
}

// --- CreateProduct ---

func TestProductService_Create_Success(t *testing.T) {
	mockRepo, service := setupProductTest(t)
	ctx := context.Background()

	mockRepo.On("CreateProduct", ctx, mock.AnythingOfType("*models.Product")).Return(nil).Once()

	product, err := service.CreateProduct(ctx, services.ProductRequest{
		Name:       "Mie Goreng",
		Price:      5000,
		Cost:       3000,
		Stock:      100,
		CategoryID: 1,
	})

	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, "Mie Goreng", product.Name)
	assert.NotEmpty(t, product.SKU)     // Auto-generated
	assert.NotEmpty(t, product.Barcode) // Auto-generated
}

func TestProductService_Create_WithCustomSKUAndBarcode(t *testing.T) {
	mockRepo, service := setupProductTest(t)
	ctx := context.Background()

	mockRepo.On("CreateProduct", ctx, mock.AnythingOfType("*models.Product")).Return(nil).Once()

	product, err := service.CreateProduct(ctx, services.ProductRequest{
		Name:       "Indomie",
		SKU:        "SKU-CUSTOM",
		Barcode:    "1234567890",
		Price:      3000,
		Cost:       2000,
		Stock:      50,
		CategoryID: 1,
	})

	assert.NoError(t, err)
	assert.Equal(t, "SKU-CUSTOM", product.SKU)
	assert.Equal(t, "1234567890", product.Barcode)
}

func TestProductService_Create_ValidationFailed_EmptyName(t *testing.T) {
	_, service := setupProductTest(t)
	ctx := context.Background()

	product, err := service.CreateProduct(ctx, services.ProductRequest{
		Name:       "",
		Price:      5000,
		Cost:       3000,
		CategoryID: 1,
	})

	assert.Error(t, err)
	assert.Nil(t, product)
	assert.Contains(t, err.Error(), "validasi gagal")
}

func TestProductService_Create_ValidationFailed_ZeroPrice(t *testing.T) {
	_, service := setupProductTest(t)
	ctx := context.Background()

	product, err := service.CreateProduct(ctx, services.ProductRequest{
		Name:       "Test Product",
		Price:      0,
		Cost:       1000,
		CategoryID: 1,
	})

	assert.Error(t, err)
	assert.Nil(t, product)
}

func TestProductService_Create_ValidationFailed_NoCategoryID(t *testing.T) {
	_, service := setupProductTest(t)
	ctx := context.Background()

	product, err := service.CreateProduct(ctx, services.ProductRequest{
		Name:  "Test Product",
		Price: 5000,
		Cost:  3000,
	})

	assert.Error(t, err)
	assert.Nil(t, product)
}

func TestProductService_Create_DuplicateSKU(t *testing.T) {
	mockRepo, service := setupProductTest(t)
	ctx := context.Background()

	mockRepo.On("CreateProduct", ctx, mock.AnythingOfType("*models.Product")).
		Return(errors.New("unique constraint violation")).Once()

	product, err := service.CreateProduct(ctx, services.ProductRequest{
		Name:       "Test",
		SKU:        "EXISTING-SKU",
		Price:      5000,
		Cost:       3000,
		Stock:      10,
		CategoryID: 1,
	})

	assert.Error(t, err)
	assert.Nil(t, product)
}

// --- GetProduct ---

func TestProductService_GetProduct_Success(t *testing.T) {
	mockRepo, service := setupProductTest(t)
	ctx := context.Background()

	expectedProduct := &models.Product{ID: 1, Name: "Test Product"}
	mockRepo.On("GetProductByID", ctx, uint(1)).Return(expectedProduct, nil).Once()

	product, err := service.GetProduct(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, expectedProduct, product)
}

func TestProductService_GetProduct_NotFound(t *testing.T) {
	mockRepo, service := setupProductTest(t)
	ctx := context.Background()

	mockRepo.On("GetProductByID", ctx, uint(999)).Return(nil, gorm.ErrRecordNotFound).Once()

	product, err := service.GetProduct(ctx, 999)

	assert.Error(t, err)
	assert.Nil(t, product)
}

// --- ListProducts ---

func TestProductService_ListProducts_Success(t *testing.T) {
	mockRepo, service := setupProductTest(t)
	ctx := context.Background()

	products := []models.Product{
		{ID: 1, Name: "Product A"},
		{ID: 2, Name: "Product B"},
	}
	mockRepo.On("GetAllProducts", ctx, 10, 0, "", "", "", "", false).Return(products, int64(2), nil).Once()

	result, count, err := service.ListProducts(ctx, 1, 10, "", "", "", "", false)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
	assert.Len(t, result, 2)
}

func TestProductService_ListProducts_DefaultsPagination(t *testing.T) {
	mockRepo, service := setupProductTest(t)
	ctx := context.Background()

	mockRepo.On("GetAllProducts", ctx, 10, 0, "", "", "", "", false).Return([]models.Product{}, int64(0), nil).Once()

	// page=0 and pageSize=0 should default to 1 and 10
	_, _, err := service.ListProducts(ctx, 0, 0, "", "", "", "", false)

	assert.NoError(t, err)
}

// --- GetLowStockProducts ---

func TestProductService_GetLowStock_Success(t *testing.T) {
	mockRepo, service := setupProductTest(t)
	ctx := context.Background()

	lowStock := []models.Product{
		{ID: 1, Name: "Almost Out", Stock: 2},
	}
	mockRepo.On("GetLowStockProducts", ctx, 10).Return(lowStock, nil).Once()

	result, err := service.GetLowStockProducts(ctx, 10)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestProductService_GetLowStock_DefaultThreshold(t *testing.T) {
	mockRepo, service := setupProductTest(t)
	ctx := context.Background()

	mockRepo.On("GetLowStockProducts", ctx, 10).Return([]models.Product{}, nil).Once()

	// threshold=0 should default to 10
	_, err := service.GetLowStockProducts(ctx, 0)

	assert.NoError(t, err)
}

// --- UpdateProduct ---

func TestProductService_Update_Success(t *testing.T) {
	mockRepo, service := setupProductTest(t)
	ctx := context.Background()

	existing := &models.Product{ID: 1, Name: "Old Name", SKU: "SKU-OLD", Barcode: "111", Price: 1000, CategoryID: 1}
	updated := &models.Product{ID: 1, Name: "New Name", SKU: "SKU-OLD", Barcode: "111", Price: 7000, CategoryID: 2}

	mockRepo.On("GetProductByID", ctx, uint(1)).Return(existing, nil).Once()
	mockRepo.On("UpdateProduct", ctx, mock.AnythingOfType("*models.Product")).Return(nil).Once()
	mockRepo.On("GetProductByID", ctx, uint(1)).Return(updated, nil).Once()

	product, err := service.UpdateProduct(ctx, 1, services.ProductRequest{
		Name:       "New Name",
		Price:      7000,
		Cost:       5000,
		Stock:      50,
		CategoryID: 2,
	})

	assert.NoError(t, err)
	assert.Equal(t, "New Name", product.Name)
}

func TestProductService_Update_NotFound(t *testing.T) {
	mockRepo, service := setupProductTest(t)
	ctx := context.Background()

	mockRepo.On("GetProductByID", ctx, uint(999)).Return(nil, gorm.ErrRecordNotFound).Once()

	product, err := service.UpdateProduct(ctx, 999, services.ProductRequest{
		Name:       "Whatever",
		Price:      1000,
		Cost:       500,
		CategoryID: 1,
	})

	assert.Error(t, err)
	assert.Nil(t, product)
}

// --- DeleteProduct ---

func TestProductService_Delete_Success(t *testing.T) {
	mockRepo, service := setupProductTest(t)
	ctx := context.Background()

	mockRepo.On("DeleteProduct", ctx, uint(1)).Return(nil).Once()

	err := service.DeleteProduct(ctx, 1)

	assert.NoError(t, err)
}

func TestProductService_Delete_NotFound(t *testing.T) {
	mockRepo, service := setupProductTest(t)
	ctx := context.Background()

	mockRepo.On("DeleteProduct", ctx, uint(999)).Return(gorm.ErrRecordNotFound).Once()

	err := service.DeleteProduct(ctx, 999)

	assert.Error(t, err)
}

func TestProductService_Delete_ForeignKeyConstraint(t *testing.T) {
	mockRepo, service := setupProductTest(t)
	ctx := context.Background()

	mockRepo.On("DeleteProduct", ctx, uint(1)).
		Return(errors.New("foreign key constraint violation")).Once()

	err := service.DeleteProduct(ctx, 1)

	assert.Error(t, err)
}

// --- RestoreProduct ---

func TestProductService_Restore_Success(t *testing.T) {
	mockRepo, service := setupProductTest(t)
	ctx := context.Background()

	mockRepo.On("RestoreProduct", ctx, uint(1)).Return(nil).Once()

	err := service.RestoreProduct(ctx, 1)

	assert.NoError(t, err)
}

// --- ForceDeleteProduct ---

func TestProductService_ForceDelete_Success(t *testing.T) {
	mockRepo, service := setupProductTest(t)
	ctx := context.Background()

	mockRepo.On("ForceDeleteProduct", ctx, uint(1)).Return(nil).Once()

	err := service.ForceDeleteProduct(ctx, 1)

	assert.NoError(t, err)
}

// --- GetStockCounts ---

func TestProductService_GetStockCounts_Success(t *testing.T) {
	mockRepo, service := setupProductTest(t)
	ctx := context.Background()

	expected := map[string]int64{"all": 100, "high": 80, "low": 15, "out": 5}
	mockRepo.On("GetStockCounts", ctx).Return(expected, nil).Once()

	counts, err := service.GetStockCounts(ctx)

	assert.NoError(t, err)
	assert.Equal(t, int64(100), counts["all"])
	assert.Equal(t, int64(5), counts["out"])
}
