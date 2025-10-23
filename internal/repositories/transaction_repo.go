package repositories

import (
	"pos-api/internal/models"

	"gorm.io/gorm"
)

// TransactionRepository mendefinisikan kontrak untuk interaksi database transaksi.
type TransactionRepository interface {
	// ProcessFullTransaction menjalankan semua operasi (stok, transaksi, detail) dalam satu DB Transaction.
	ProcessFullTransaction(transaction *models.Transaction, productUpdates map[uint]int) error
	GetTransactionByID(id uint) (*models.Transaction, error)
	ListTransactions() ([]models.Transaction, error)
}

type transactionRepository struct {
	DB *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{
		DB: db,
	}
}

func (r *transactionRepository) GetTransactionByID(id uint) (*models.Transaction, error) {
	var transaction models.Transaction

	// Gunakan Preload untuk mengambil relasi TransactionDetails dan Product di dalamnya
	result := r.DB.Preload("TransactionDetails").Preload("TransactionDetails.Product").First(&transaction, id)

	if result.Error != nil {
		return nil, result.Error
	}

	return &transaction, nil
}

func (r *transactionRepository) ListTransactions() ([]models.Transaction, error) {
	var transactions []models.Transaction

	result := r.DB.Preload("TransactionDetails").Preload("TransactionDetails.Product").Find(&transactions)

	if result.Error != nil {
		return nil, result.Error
	}

	return transactions, nil
}

// ProcessFullTransaction adalah fungsi KRITIS yang menggunakan GORM Transaction.
func (r *transactionRepository) ProcessFullTransaction(transaction *models.Transaction, productUpdates map[uint]int) error {
	// Mulai GORM Transaction
	tx := r.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Pastikan Commit atau Rollback dilakukan
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Catat Transaksi Utama
	if err := tx.Create(transaction).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 2. Update Stok Produk
	for productID, newStock := range productUpdates {
		// Gunakan .Model() untuk hanya mengupdate field 'stock'
		if err := tx.Model(&models.Product{}).Where("id = ?", productID).Update("stock", newStock).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 3. Commit Transaction
	return tx.Commit().Error
}
