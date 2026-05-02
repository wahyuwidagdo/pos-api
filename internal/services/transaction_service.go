package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"pos-api/internal/models"
	"pos-api/internal/pkg/events"
	"pos-api/internal/repositories"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// ItemRequest merepresentasikan satu item yang dibeli dalam request API
type ItemRequest struct {
	ProductID uint `json:"product_id" validate:"required"`
	Quantity  int  `json:"quantity" validate:"required,gt=0"`
}

// TransactionRequest mendefinisikan DTO untuk pencatatan transaksi penjualan
type TransactionRequest struct {
	PaymentMethod string        `json:"payment_method" validate:"required"` // e.g., "Cash", "QRIS"
	Cash          float64       `json:"cash" validate:"required,gte=0"`     // Uang yang dibayarkan pelanggan
	Discount      float64       `json:"discount" validate:"gte=0"`
	Items         []ItemRequest `json:"items" validate:"required,min=1"` // Daftar produk yang dibeli
	UserID        uint          // Added for Event-Driven Architecture (Cashier ID)
}

// PaginationResult wraps data with metadata
type PaginationData struct {
	Total       int64       `json:"total_items"`
	TotalPages  int         `json:"total_pages"`
	CurrentPage int         `json:"current_page"`
	Limit       int         `json:"limit"`
	Data        interface{} `json:"data"`
}

type TransactionService interface {
	ProcessTransaction(ctx context.Context, req TransactionRequest) (*models.Transaction, error)
	GetTransaction(ctx context.Context, id uint) (*models.Transaction, error)
	ListTransactions(ctx context.Context, page int, limit int, search, startDate, endDate string) (*PaginationData, error)
	CancelTransaction(ctx context.Context, id uint) error
	ReturnTransaction(ctx context.Context, id uint) error
}

type transactionService struct {
	repo        repositories.TransactionRepository
	productRepo repositories.ProductRepository
	validator   *validator.Validate
}

func NewTransactionService(repo repositories.TransactionRepository, productRepo repositories.ProductRepository) TransactionService {
	return &transactionService{
		repo:        repo,
		productRepo: productRepo,
		validator:   validator.New(),
	}
}

func (s *transactionService) GetTransaction(ctx context.Context, id uint) (*models.Transaction, error) {
	transaction, err := s.repo.GetTransactionByID(ctx, id)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("transaksi dengan ID %d tidak ditemukan", id)
		}
		// Mengembalikan error dengan %w adalah best practice untuk menyimpan konteks error
		return nil, fmt.Errorf("gagal mengambil data transaksi: %w", err)
	}

	return transaction, nil
}

func (s *transactionService) ListTransactions(ctx context.Context, page int, limit int, search, startDate, endDate string) (*PaginationData, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	transactions, totalItem, err := s.repo.ListTransactions(ctx, page, limit, search, startDate, endDate)

	if err != nil {
		// Asumsi error yang dikembalikan adalah error database, tidak perlu penanganan not found
		return nil, fmt.Errorf("gagal menampilkan daftar transaksi: %w", err)
	}

	totalPages := int(totalItem) / limit
	if int(totalItem)%limit != 0 {
		totalPages++
	}

	return &PaginationData{
		Total:       totalItem,
		TotalPages:  totalPages,
		CurrentPage: page,
		Limit:       limit,
		Data:        transactions,
	}, nil
}

func (s *transactionService) ProcessTransaction(ctx context.Context, req TransactionRequest) (*models.Transaction, error) {
	// 1. Validasi Input Request (Wajib dilakukan di Service Layer)
	if err := s.validator.Struct(req); err != nil {
		return nil, errors.New("validasi gagal: " + err.Error())
	}

	// Inisiasi variabel kalkulasi
	var (
		totalAmount        float64 // Total sebelum diskon
		transactionDetails []models.TransactionDetail
	)
	// Note: We don't check for stock here anymore, because the Repository does it atomically.
	// However, we can still do a read-only check for better UX (fail fast), but we won't rely on it for data integrity.
	for _, itemReq := range req.Items {
		// 2a. Get Product
		product, err := s.productRepo.GetProductByID(ctx, itemReq.ProductID)
		if err != nil {
			return nil, fmt.Errorf("produk dengan ID %d tidak ditemukan", itemReq.ProductID)
		}

		// 2c. Calculate Subtotal & Total
		priceAtSale := product.Price
		subTotal := priceAtSale * float64(itemReq.Quantity)
		totalAmount += subTotal

		// 2d. Prepare Transaction Detail
		transactionDetails = append(transactionDetails, models.TransactionDetail{
			ProductID:   itemReq.ProductID,
			ProductName: product.Name,
			Quantity:    itemReq.Quantity,
			PriceAtSale: priceAtSale,
			CostAtSale:  product.Cost,
			SubTotal:    subTotal,
		})
	}

	// 3. Final Calculation
	grandTotal := totalAmount - req.Discount
	change := req.Cash - grandTotal

	if change < 0 {
		return nil, errors.New("jumlah uang tunai kurang")
	}

	// 4. Build Main Transaction Struct
	transaction := models.Transaction{
		TransactionCode:    fmt.Sprintf("INV-%d", time.Now().UnixNano()),
		TotalAmount:        totalAmount,
		Discount:           req.Discount,
		GrandTotal:         grandTotal,
		Cash:               req.Cash,
		Change:             change,
		PaymentMethod:      req.PaymentMethod,
		TransactionDetails: transactionDetails,
	}

	// 5. Call Repository (Atomic Transaction)
	if err := s.repo.ProcessFullTransaction(ctx, &transaction); err != nil {
		return nil, errors.New("gagal memproses transaksi: " + err.Error())
	}

	// 6. Selesai! Kembalikan Transaksi yang sudah tersimpan (dengan ID)
	finalTransaction, err := s.GetTransaction(ctx, transaction.ID)
	if err != nil {
		return &transaction, nil
	}

	return finalTransaction, nil
}

func (s *transactionService) CancelTransaction(ctx context.Context, id uint) error {
	tx, err := s.repo.GetTransactionByID(ctx, id)
	if err != nil {
		return fmt.Errorf("transaksi dengan ID %d tidak ditemukan", id)
	}
	if tx.Status != "completed" {
		return fmt.Errorf("transaksi sudah berstatus '%s', tidak bisa dibatalkan", tx.Status)
	}

	// Update status and fire event
	if err := s.repo.UpdateTransactionState(ctx, tx, "cancelled", events.EventTransactionCancelled); err != nil {
		return fmt.Errorf("gagal membatalkan transaksi: %w", err)
	}

	return nil
}

func (s *transactionService) ReturnTransaction(ctx context.Context, id uint) error {
	tx, err := s.repo.GetTransactionByID(ctx, id)
	if err != nil {
		return fmt.Errorf("transaksi dengan ID %d tidak ditemukan", id)
	}
	if tx.Status != "completed" {
		return fmt.Errorf("transaksi sudah berstatus '%s', tidak bisa diretur", tx.Status)
	}

	// Update status and fire event
	if err := s.repo.UpdateTransactionState(ctx, tx, "returned", events.EventTransactionReturned); err != nil {
		return fmt.Errorf("gagal meretur transaksi: %w", err)
	}

	return nil
}
