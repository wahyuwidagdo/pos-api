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

func setupTransactionTest(t *testing.T) (*mocks.TransactionRepository, *mocks.ProductRepository, services.TransactionService) {
	mockRepo := mocks.NewTransactionRepository(t)
	mockProductRepo := mocks.NewProductRepository(t)
	service := services.NewTransactionService(mockRepo, mockProductRepo)
	return mockRepo, mockProductRepo, service
}

// --- ProcessTransaction ---

func TestTransactionService_Process_Success(t *testing.T) {
	mockRepo, mockProductRepo, service := setupTransactionTest(t)
	ctx := context.Background()

	productID := uint(1)
	price := 10000.0

	mockProductRepo.On("GetProductByID", ctx, productID).Return(&models.Product{
		ID:    productID,
		Name:  "Test Product",
		Price: price,
		Cost:  6000,
		Stock: 10,
	}, nil)

	mockRepo.On("ProcessFullTransaction", ctx, mock.MatchedBy(func(trx *models.Transaction) bool {
		return trx.TotalAmount == price*2 && len(trx.TransactionDetails) == 1
	})).Return(nil)

	mockRepo.On("GetTransactionByID", ctx, mock.AnythingOfType("uint")).Return(&models.Transaction{
		ID:          1,
		TotalAmount: price * 2,
		GrandTotal:  price * 2,
		CreatedAt:   time.Now(),
	}, nil)

	trx, err := service.ProcessTransaction(ctx, services.TransactionRequest{
		PaymentMethod: "Cash",
		Cash:          50000,
		Items:         []services.ItemRequest{{ProductID: productID, Quantity: 2}},
	})

	assert.NoError(t, err)
	assert.NotNil(t, trx)
	assert.Equal(t, price*2, trx.TotalAmount)
}

func TestTransactionService_Process_InsufficientPayment(t *testing.T) {
	_, mockProductRepo, service := setupTransactionTest(t)
	ctx := context.Background()

	mockProductRepo.On("GetProductByID", ctx, uint(1)).Return(&models.Product{
		ID:    1,
		Price: 10000.0,
	}, nil)

	trx, err := service.ProcessTransaction(ctx, services.TransactionRequest{
		PaymentMethod: "Cash",
		Cash:          5000, // Less than total
		Items:         []services.ItemRequest{{ProductID: 1, Quantity: 1}},
	})

	assert.Error(t, err)
	assert.Nil(t, trx)
	assert.Contains(t, err.Error(), "uang tunai kurang")
}

func TestTransactionService_Process_ProductNotFound(t *testing.T) {
	_, mockProductRepo, service := setupTransactionTest(t)
	ctx := context.Background()

	mockProductRepo.On("GetProductByID", ctx, uint(999)).Return(nil, errors.New("record not found"))

	trx, err := service.ProcessTransaction(ctx, services.TransactionRequest{
		PaymentMethod: "Cash",
		Cash:          50000,
		Items:         []services.ItemRequest{{ProductID: 999, Quantity: 1}},
	})

	assert.Error(t, err)
	assert.Nil(t, trx)
	assert.Contains(t, err.Error(), "tidak ditemukan")
}

func TestTransactionService_Process_EmptyItems(t *testing.T) {
	_, _, service := setupTransactionTest(t)
	ctx := context.Background()

	trx, err := service.ProcessTransaction(ctx, services.TransactionRequest{
		PaymentMethod: "Cash",
		Cash:          50000,
		Items:         []services.ItemRequest{},
	})

	assert.Error(t, err)
	assert.Nil(t, trx)
	assert.Contains(t, err.Error(), "validasi gagal")
}

func TestTransactionService_Process_MultipleItems(t *testing.T) {
	mockRepo, mockProductRepo, service := setupTransactionTest(t)
	ctx := context.Background()

	mockProductRepo.On("GetProductByID", ctx, uint(1)).Return(&models.Product{
		ID: 1, Name: "A", Price: 5000, Cost: 3000, Stock: 10,
	}, nil)
	mockProductRepo.On("GetProductByID", ctx, uint(2)).Return(&models.Product{
		ID: 2, Name: "B", Price: 15000, Cost: 10000, Stock: 5,
	}, nil)

	mockRepo.On("ProcessFullTransaction", ctx, mock.MatchedBy(func(trx *models.Transaction) bool {
		// 5000*2 + 15000*1 = 25000
		return trx.TotalAmount == 25000 && len(trx.TransactionDetails) == 2
	})).Return(nil)

	mockRepo.On("GetTransactionByID", ctx, mock.AnythingOfType("uint")).Return(&models.Transaction{
		ID:          1,
		TotalAmount: 25000,
	}, nil)

	trx, err := service.ProcessTransaction(ctx, services.TransactionRequest{
		PaymentMethod: "Cash",
		Cash:          30000,
		Items: []services.ItemRequest{
			{ProductID: 1, Quantity: 2},
			{ProductID: 2, Quantity: 1},
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, trx)
}

func TestTransactionService_Process_WithDiscount(t *testing.T) {
	mockRepo, mockProductRepo, service := setupTransactionTest(t)
	ctx := context.Background()

	mockProductRepo.On("GetProductByID", ctx, uint(1)).Return(&models.Product{
		ID: 1, Name: "A", Price: 20000, Cost: 12000, Stock: 10,
	}, nil)

	mockRepo.On("ProcessFullTransaction", ctx, mock.MatchedBy(func(trx *models.Transaction) bool {
		return trx.Discount == 5000 && trx.GrandTotal == 15000
	})).Return(nil)

	mockRepo.On("GetTransactionByID", ctx, mock.AnythingOfType("uint")).Return(&models.Transaction{
		ID: 1, TotalAmount: 20000, Discount: 5000, GrandTotal: 15000,
	}, nil)

	trx, err := service.ProcessTransaction(ctx, services.TransactionRequest{
		PaymentMethod: "Cash",
		Cash:          20000,
		Discount:      5000,
		Items:         []services.ItemRequest{{ProductID: 1, Quantity: 1}},
	})

	assert.NoError(t, err)
	assert.NotNil(t, trx)
}

func TestTransactionService_Process_RepositoryError(t *testing.T) {
	mockRepo, mockProductRepo, service := setupTransactionTest(t)
	ctx := context.Background()

	mockProductRepo.On("GetProductByID", ctx, uint(1)).Return(&models.Product{
		ID: 1, Price: 5000, Cost: 3000, Stock: 10,
	}, nil)

	mockRepo.On("ProcessFullTransaction", ctx, mock.Anything).
		Return(errors.New("insufficient stock")).Once()

	trx, err := service.ProcessTransaction(ctx, services.TransactionRequest{
		PaymentMethod: "Cash",
		Cash:          50000,
		Items:         []services.ItemRequest{{ProductID: 1, Quantity: 1}},
	})

	assert.Error(t, err)
	assert.Nil(t, trx)
	assert.Contains(t, err.Error(), "gagal memproses transaksi")
}

// --- GetTransaction ---

func TestTransactionService_Get_Success(t *testing.T) {
	mockRepo, _, service := setupTransactionTest(t)
	ctx := context.Background()

	expected := &models.Transaction{ID: 1, TotalAmount: 50000, Status: "completed"}
	mockRepo.On("GetTransactionByID", ctx, uint(1)).Return(expected, nil).Once()

	trx, err := service.GetTransaction(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, expected, trx)
}

func TestTransactionService_Get_NotFound(t *testing.T) {
	mockRepo, _, service := setupTransactionTest(t)
	ctx := context.Background()

	mockRepo.On("GetTransactionByID", ctx, uint(999)).Return(nil, errors.New("record not found")).Once()

	trx, err := service.GetTransaction(ctx, 999)

	assert.Error(t, err)
	assert.Nil(t, trx)
}

// --- ListTransactions ---

func TestTransactionService_List_Success(t *testing.T) {
	mockRepo, _, service := setupTransactionTest(t)
	ctx := context.Background()

	transactions := []models.Transaction{
		{ID: 1, TotalAmount: 10000},
		{ID: 2, TotalAmount: 20000},
	}
	mockRepo.On("ListTransactions", ctx, 1, 10).Return(transactions, int64(2), nil).Once()

	result, err := service.ListTransactions(ctx, 1, 10)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(2), result.Total)
}

func TestTransactionService_List_DefaultsPagination(t *testing.T) {
	mockRepo, _, service := setupTransactionTest(t)
	ctx := context.Background()

	mockRepo.On("ListTransactions", ctx, 1, 10).Return([]models.Transaction{}, int64(0), nil).Once()

	// page=0 and limit=0 should default to 1 and 10
	result, err := service.ListTransactions(ctx, 0, 0)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// --- CancelTransaction ---

func TestTransactionService_Cancel_Success(t *testing.T) {
	mockRepo, _, service := setupTransactionTest(t)
	ctx := context.Background()

	trx := &models.Transaction{ID: 1, Status: "completed"}
	mockRepo.On("GetTransactionByID", ctx, uint(1)).Return(trx, nil).Once()
	mockRepo.On("UpdateTransactionState", ctx, trx, "cancelled", mock.AnythingOfType("string")).Return(nil).Once()

	err := service.CancelTransaction(ctx, 1)

	assert.NoError(t, err)
}

func TestTransactionService_Cancel_AlreadyCancelled(t *testing.T) {
	mockRepo, _, service := setupTransactionTest(t)
	ctx := context.Background()

	trx := &models.Transaction{ID: 1, Status: "cancelled"}
	mockRepo.On("GetTransactionByID", ctx, uint(1)).Return(trx, nil).Once()

	err := service.CancelTransaction(ctx, 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tidak bisa dibatalkan")
}

func TestTransactionService_Cancel_NotFound(t *testing.T) {
	mockRepo, _, service := setupTransactionTest(t)
	ctx := context.Background()

	mockRepo.On("GetTransactionByID", ctx, uint(999)).Return(nil, errors.New("not found")).Once()

	err := service.CancelTransaction(ctx, 999)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tidak ditemukan")
}

// --- ReturnTransaction ---

func TestTransactionService_Return_Success(t *testing.T) {
	mockRepo, _, service := setupTransactionTest(t)
	ctx := context.Background()

	trx := &models.Transaction{ID: 1, Status: "completed"}
	mockRepo.On("GetTransactionByID", ctx, uint(1)).Return(trx, nil).Once()
	mockRepo.On("UpdateTransactionState", ctx, trx, "returned", mock.AnythingOfType("string")).Return(nil).Once()

	err := service.ReturnTransaction(ctx, 1)

	assert.NoError(t, err)
}

func TestTransactionService_Return_AlreadyReturned(t *testing.T) {
	mockRepo, _, service := setupTransactionTest(t)
	ctx := context.Background()

	trx := &models.Transaction{ID: 1, Status: "returned"}
	mockRepo.On("GetTransactionByID", ctx, uint(1)).Return(trx, nil).Once()

	err := service.ReturnTransaction(ctx, 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tidak bisa diretur")
}
