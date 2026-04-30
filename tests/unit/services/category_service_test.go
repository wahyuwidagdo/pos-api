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

func setupCategoryTest(t *testing.T) (*mocks.CategoryRepository, services.CategoryService) {
	mockRepo := mocks.NewCategoryRepository(t)
	service := services.NewCategoryService(mockRepo)
	return mockRepo, service
}

// --- CreateCategory ---

func TestCategoryService_Create_Success(t *testing.T) {
	mockRepo, service := setupCategoryTest(t)
	ctx := context.Background()

	mockRepo.On("CreateCategory", ctx, mock.AnythingOfType("*models.Category")).Return(nil).Once()

	cat, err := service.CreateCategory(ctx, services.CategoryRequest{Name: "Makanan"})

	assert.NoError(t, err)
	assert.NotNil(t, cat)
	assert.Equal(t, "Makanan", cat.Name)
}

func TestCategoryService_Create_ValidationFailed_NameTooShort(t *testing.T) {
	_, service := setupCategoryTest(t)
	ctx := context.Background()

	cat, err := service.CreateCategory(ctx, services.CategoryRequest{Name: "AB"})

	assert.Error(t, err)
	assert.Nil(t, cat)
	assert.Contains(t, err.Error(), "validasi gagal")
}

func TestCategoryService_Create_ValidationFailed_EmptyName(t *testing.T) {
	_, service := setupCategoryTest(t)
	ctx := context.Background()

	cat, err := service.CreateCategory(ctx, services.CategoryRequest{Name: ""})

	assert.Error(t, err)
	assert.Nil(t, cat)
}

func TestCategoryService_Create_DuplicateName(t *testing.T) {
	mockRepo, service := setupCategoryTest(t)
	ctx := context.Background()

	mockRepo.On("CreateCategory", ctx, mock.AnythingOfType("*models.Category")).
		Return(errors.New("unique constraint violation")).Once()

	cat, err := service.CreateCategory(ctx, services.CategoryRequest{Name: "Makanan"})

	assert.Error(t, err)
	assert.Nil(t, cat)
}

// --- GetCategory ---

func TestCategoryService_Get_Success(t *testing.T) {
	mockRepo, service := setupCategoryTest(t)
	ctx := context.Background()

	expected := &models.Category{ID: 1, Name: "Minuman"}
	mockRepo.On("GetCategoryByID", ctx, uint(1)).Return(expected, nil).Once()

	cat, err := service.GetCategory(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, expected, cat)
}

func TestCategoryService_Get_NotFound(t *testing.T) {
	mockRepo, service := setupCategoryTest(t)
	ctx := context.Background()

	mockRepo.On("GetCategoryByID", ctx, uint(999)).Return(nil, gorm.ErrRecordNotFound).Once()

	cat, err := service.GetCategory(ctx, 999)

	assert.Error(t, err)
	assert.Nil(t, cat)
}

// --- ListCategories ---

func TestCategoryService_List_Success(t *testing.T) {
	mockRepo, service := setupCategoryTest(t)
	ctx := context.Background()

	categories := []models.Category{
		{ID: 1, Name: "Makanan"},
		{ID: 2, Name: "Minuman"},
	}
	mockRepo.On("GetAllCategories", ctx, false).Return(categories, nil).Once()

	result, err := service.ListCategories(ctx, false)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestCategoryService_List_OnlyTrashed(t *testing.T) {
	mockRepo, service := setupCategoryTest(t)
	ctx := context.Background()

	mockRepo.On("GetAllCategories", ctx, true).Return([]models.Category{}, nil).Once()

	result, err := service.ListCategories(ctx, true)

	assert.NoError(t, err)
	assert.Empty(t, result)
}

// --- UpdateCategory ---

func TestCategoryService_Update_Success(t *testing.T) {
	mockRepo, service := setupCategoryTest(t)
	ctx := context.Background()

	existing := &models.Category{ID: 1, Name: "Old Name"}
	mockRepo.On("GetCategoryByID", ctx, uint(1)).Return(existing, nil).Once()
	mockRepo.On("UpdateCategory", ctx, mock.AnythingOfType("*models.Category")).Return(nil).Once()

	cat, err := service.UpdateCategory(ctx, 1, services.CategoryRequest{Name: "New Name"})

	assert.NoError(t, err)
	assert.Equal(t, "New Name", cat.Name)
}

func TestCategoryService_Update_NotFound(t *testing.T) {
	mockRepo, service := setupCategoryTest(t)
	ctx := context.Background()

	mockRepo.On("GetCategoryByID", ctx, uint(999)).Return(nil, gorm.ErrRecordNotFound).Once()

	cat, err := service.UpdateCategory(ctx, 999, services.CategoryRequest{Name: "Whatever"})

	assert.Error(t, err)
	assert.Nil(t, cat)
}

func TestCategoryService_Update_DuplicateName(t *testing.T) {
	mockRepo, service := setupCategoryTest(t)
	ctx := context.Background()

	existing := &models.Category{ID: 1, Name: "Old Name"}
	mockRepo.On("GetCategoryByID", ctx, uint(1)).Return(existing, nil).Once()
	mockRepo.On("UpdateCategory", ctx, mock.AnythingOfType("*models.Category")).
		Return(errors.New("duplicate key")).Once()

	cat, err := service.UpdateCategory(ctx, 1, services.CategoryRequest{Name: "Taken Name"})

	assert.Error(t, err)
	assert.Nil(t, cat)
}

// --- DeleteCategory ---

func TestCategoryService_Delete_Success(t *testing.T) {
	mockRepo, service := setupCategoryTest(t)
	ctx := context.Background()

	mockRepo.On("DeleteCategory", ctx, uint(1)).Return(nil).Once()

	err := service.DeleteCategory(ctx, 1)

	assert.NoError(t, err)
}

func TestCategoryService_Delete_NotFound(t *testing.T) {
	mockRepo, service := setupCategoryTest(t)
	ctx := context.Background()

	mockRepo.On("DeleteCategory", ctx, uint(999)).Return(gorm.ErrRecordNotFound).Once()

	err := service.DeleteCategory(ctx, 999)

	assert.Error(t, err)
}

func TestCategoryService_Delete_ForeignKeyConstraint(t *testing.T) {
	mockRepo, service := setupCategoryTest(t)
	ctx := context.Background()

	mockRepo.On("DeleteCategory", ctx, uint(1)).
		Return(errors.New("foreign key constraint violation")).Once()

	err := service.DeleteCategory(ctx, 1)

	assert.Error(t, err)
}

// --- RestoreCategory ---

func TestCategoryService_Restore_Success(t *testing.T) {
	mockRepo, service := setupCategoryTest(t)
	ctx := context.Background()

	mockRepo.On("RestoreCategory", ctx, uint(1)).Return(nil).Once()

	err := service.RestoreCategory(ctx, 1)

	assert.NoError(t, err)
}

func TestCategoryService_Restore_NotFound(t *testing.T) {
	mockRepo, service := setupCategoryTest(t)
	ctx := context.Background()

	mockRepo.On("RestoreCategory", ctx, uint(999)).Return(gorm.ErrRecordNotFound).Once()

	err := service.RestoreCategory(ctx, 999)

	assert.Error(t, err)
}

// --- ForceDeleteCategory ---

func TestCategoryService_ForceDelete_Success(t *testing.T) {
	mockRepo, service := setupCategoryTest(t)
	ctx := context.Background()

	mockRepo.On("ForceDeleteCategory", ctx, uint(1)).Return(nil).Once()

	err := service.ForceDeleteCategory(ctx, 1)

	assert.NoError(t, err)
}
