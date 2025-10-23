package services

import (
	"errors"
	"fmt"
	"time"

	"pos-api/internal/models"
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
}

type TransactionService interface {
	ProcessTransaction(req TransactionRequest) (*models.Transaction, error)
	GetTransaction(id uint) (*models.Transaction, error)
	ListTransactions() ([]models.Transaction, error)
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

func (s *transactionService) GetTransaction(id uint) (*models.Transaction, error) {
	transaction, err := s.repo.GetTransactionByID(id)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("transaksi dengan ID %d tidak ditemukan", id)
		}
		// Mengembalikan error dengan %w adalah best practice untuk menyimpan konteks error
		return nil, fmt.Errorf("gagal mengambil data transaksi: %w", err)
	}

	return transaction, nil
}

func (s *transactionService) ListTransactions() ([]models.Transaction, error) {
	transaction, err := s.repo.ListTransactions()

	if err != nil {
		// Asumsi error yang dikembalikan adalah error database, tidak perlu penanganan not found
		return nil, fmt.Errorf("gagal menampilkan daftar transaksi: %w", err)
	}

	return transaction, nil
}

func (s *transactionService) ProcessTransaction(req TransactionRequest) (*models.Transaction, error) {
	// 1. Validasi Input Request (Wajib dilakukan di Service Layer)
	if err := s.validator.Struct(req); err != nil {
		return nil, errors.New("validasi gagal: " + err.Error())
	}

	// Inisiasi variabel kalkulasi
	var (
		totalAmount        float64 // Total sebelum diskon
		transactionDetails []models.TransactionDetail
		productUpdates     = make(map[uint]int) // Untuk menyimpan (ProductID -> StokBaru)
	)

	// 2. Loop melalui setiap Item di Request untuk Kalkulasi dan Verifikasi Stok
	for _, itemReq := range req.Items {
		// 2a. Ambil Produk dari DB untuk mendapatkan harga & stok saat ini
		product, err := s.productRepo.GetProductByID(itemReq.ProductID)
		if err != nil {
			// Asumsi gorm.ErrRecordNotFound sudah di-handle di ProductRepo/Service
			return nil, fmt.Errorf("produk dengan ID %d tidak ditemukan", itemReq.ProductID)
		}

		// 2b. Verifikasi Stok
		if product.Stock < itemReq.Quantity {
			return nil, fmt.Errorf("stok %s tidak cukup. Tersedia: %d", product.Name, product.Stock)
		}

		// 2c. Kalkulasi Subtotal & Total Keseluruhan
		priceAtSale := product.Price
		subTotal := priceAtSale * float64(itemReq.Quantity)
		totalAmount += subTotal

		// 2d. Siapkan Detail Transaksi
		transactionDetails = append(transactionDetails, models.TransactionDetail{
			ProductID:   itemReq.ProductID,
			ProductName: product.Name, // Simpan cache nama
			Quantity:    itemReq.Quantity,
			PriceAtSale: priceAtSale,
			SubTotal:    subTotal,
		})

		// 2e. Siapkan Update Stok
		productUpdates[itemReq.ProductID] = product.Stock - itemReq.Quantity
	}

	// 3. Kalkulasi Akhir (Total, Diskon, Kembalian)
	grandTotal := totalAmount - req.Discount
	change := req.Cash - grandTotal

	if change < 0 {
		return nil, errors.New("jumlah uang tunai kurang")
	}

	// 4. Bangun Struct Transaction Utama
	transaction := models.Transaction{
		TransactionCode:    fmt.Sprintf("INV-%d", time.Now().UnixNano()), // Kode unik sementara
		TotalAmount:        totalAmount,
		Discount:           req.Discount,
		GrandTotal:         grandTotal,
		Cash:               req.Cash,
		Change:             change,
		PaymentMethod:      req.PaymentMethod,
		TransactionDetails: transactionDetails, // Relasi sudah terisi di sini
	}

	// 5. Panggil Transaction Repository (Running DB Transaction)
	if err := s.repo.ProcessFullTransaction(&transaction, productUpdates); err != nil {
		// Jika gagal, Repository sudah melakukan Rollback
		return nil, errors.New("gagal memproses transaksi di database: " + err.Error())
	}

	// 6. Selesai! Kembalikan Transaksi yang sudah tersimpan (dengan ID)
	finalTransaction, err := s.GetTransaction(transaction.ID)
	if err != nil {
		return &transaction, nil
	}

	return finalTransaction, nil
}
