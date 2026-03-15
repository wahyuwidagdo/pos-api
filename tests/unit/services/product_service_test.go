package services_test

import (
	"context"
	"errors"
	"testing"

	"pos-api/internal/models"
	"pos-api/internal/services"
	"pos-api/tests/mocks"

	"github.com/stretchr/testify/assert"
)

func TestProductService_GetProduct(t *testing.T) {
	// Setup
	mockRepo := new(mocks.ProductRepository)
	service := services.NewProductService(mockRepo)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		expectedProduct := &models.Product{ID: 1, Name: "Test Product"}
		// Expect GetProductByID to be called with context and return expectedProduct
		mockRepo.On("GetProductByID", ctx, uint(1)).Return(expectedProduct, nil).Once()

		product, err := service.GetProduct(ctx, 1)

		assert.NoError(t, err)
		assert.Equal(t, expectedProduct, product)
		mockRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockRepo.On("GetProductByID", ctx, uint(999)).Return(nil, errors.New("record not found")).Once()

		product, err := service.GetProduct(ctx, 999)

		assert.Error(t, err)
		assert.Nil(t, product)
		mockRepo.AssertExpectations(t)
	})
}
