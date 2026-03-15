package services_test

import (
	"context"
	"testing"
	"time"

	"pos-api/internal/models"
	"pos-api/internal/services"
	"pos-api/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProcessTransaction_Success(t *testing.T) {
	mockRepo := mocks.NewTransactionRepository(t)
	mockProductRepo := mocks.NewProductRepository(t)
	service := services.NewTransactionService(mockRepo, mockProductRepo)

	ctx := context.Background()
	productID := uint(1)
	quantity := 2
	price := 10000.0

	req := services.TransactionRequest{
		PaymentMethod: "CASH",
		Cash:          50000,
		Items: []services.ItemRequest{
			{ProductID: productID, Quantity: quantity},
		},
	}

	// Mock ProductRepo
	mockProductRepo.On("GetProductByID", ctx, productID).Return(&models.Product{
		ID:    productID,
		Name:  "Test Product",
		Price: price,
		Stock: 10, // Stock check is now atomic in repo, but we still fetch product for price
	}, nil)

	// Mock ProcessFullTransaction
	// We expect this to be called with a Transaction struct
	mockRepo.On("ProcessFullTransaction", ctx, mock.MatchedBy(func(trx *models.Transaction) bool {
		return trx.TotalAmount == price*float64(quantity) && len(trx.TransactionDetails) == 1
	})).Return(nil)

	// Mock GetTransactionByID (called after successful transaction to return full object)
	mockRepo.On("GetTransactionByID", ctx, mock.AnythingOfType("uint")).Return(&models.Transaction{
		ID:          1,
		TotalAmount: price * float64(quantity),
		CreatedAt:   time.Now(),
	}, nil)

	trx, err := service.ProcessTransaction(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, trx)
	assert.Equal(t, price*float64(quantity), trx.TotalAmount)

	mockRepo.AssertExpectations(t)
	mockProductRepo.AssertExpectations(t)
}

func TestProcessTransaction_InsufficientPayment(t *testing.T) {
	mockRepo := mocks.NewTransactionRepository(t)
	mockProductRepo := mocks.NewProductRepository(t)
	service := services.NewTransactionService(mockRepo, mockProductRepo)

	ctx := context.Background()
	productID := uint(1)

	req := services.TransactionRequest{
		PaymentMethod: "CASH",
		Cash:          5000, // Less than total
		Items: []services.ItemRequest{
			{ProductID: productID, Quantity: 1},
		},
	}

	mockProductRepo.On("GetProductByID", ctx, productID).Return(&models.Product{
		ID:    productID,
		Price: 10000.0,
	}, nil)

	trx, err := service.ProcessTransaction(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, trx)
	assert.Contains(t, err.Error(), "uang tunai kurang")
}
