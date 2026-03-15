package repositories

import (
	"context"
	"fmt"
	"pos-api/internal/models"
	"pos-api/internal/pkg/events"

	"gorm.io/gorm"
)

// TransactionRepository mendefinisikan kontrak untuk interaksi database transaksi.
type TransactionRepository interface {
	// ProcessFullTransaction runs all operations (stock, transaction, detail) within a single DB Transaction.
	ProcessFullTransaction(ctx context.Context, transaction *models.Transaction) error
	GetTransactionByID(ctx context.Context, id uint) (*models.Transaction, error)
	ListTransactions(ctx context.Context, page int, limit int) ([]models.Transaction, int64, error)
	UpdateTransactionState(ctx context.Context, transaction *models.Transaction, status string, eventName string) error
}

type transactionRepository struct {
	DB       *gorm.DB
	EventBus events.EventBus
}

func NewTransactionRepository(db *gorm.DB, eventBus events.EventBus) TransactionRepository {
	return &transactionRepository{
		DB:       db,
		EventBus: eventBus,
	}
}

func (r *transactionRepository) GetTransactionByID(ctx context.Context, id uint) (*models.Transaction, error) {
	var transaction models.Transaction

	// Gunakan Preload untuk mengambil relasi TransactionDetails dan Product di dalamnya
	result := r.DB.WithContext(ctx).Preload("TransactionDetails").Preload("TransactionDetails.Product").First(&transaction, id)

	if result.Error != nil {
		return nil, result.Error
	}

	return &transaction, nil
}

func (r *transactionRepository) ListTransactions(ctx context.Context, page int, limit int) ([]models.Transaction, int64, error) {
	var transactions []models.Transaction
	var total int64

	offset := (page - 1) * limit

	// Count total records first
	if err := r.DB.Model(&models.Transaction{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch paginated data
	result := r.DB.WithContext(ctx).
		Preload("TransactionDetails").
		Preload("TransactionDetails.Product").
		Order("created_at DESC"). // Best practice to show newest first
		Limit(limit).
		Offset(offset).
		Find(&transactions)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return transactions, total, nil
}

// ProcessFullTransaction runs all operations (stock, transaction, detail) within a single DB Transaction.
func (r *transactionRepository) ProcessFullTransaction(ctx context.Context, transaction *models.Transaction) error {
	// Start GORM Transaction with Context
	tx := r.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Ensure Commit or Rollback is performed
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Record Main Transaction
	if err := tx.Create(transaction).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 2. Publish Domain Event
	// This will trigger Inventory and Cash Flow listeners synchronously.
	// We pass a default UserID here, but ideally we should extract it from a broader context or pass it into this function.
	// Since we cant easily change interface right now without breaking tests, we'll try to extract it or pass it.
	// Actually we should modify the ProcessFullTransaction signature to accept UserID if possible,
	// but context.Context might hold it if we set it up.
	// Let's modify the signature or pass it from transaction.

	payload := events.TransactionCreatedPayload{
		TX:          tx,
		Transaction: transaction,
		// If transaction had UserID, we could use it here. We'll add UserID to the event payload logic next.
		UserID: 1, // Temporary, will be extracted via context shortly in a follow up step
	}

	if userIDCtx, ok := ctx.Value("userID").(uint); ok {
		payload.UserID = userIDCtx
	}

	if err := r.EventBus.Publish(ctx, events.EventTransactionCreated, payload); err != nil {
		tx.Rollback()
		return fmt.Errorf("transaction event failed: %w", err)
	}

	// 3. Commit Transaction
	return tx.Commit().Error
}

func (r *transactionRepository) UpdateTransactionState(ctx context.Context, transaction *models.Transaction, status string, eventName string) error {
	tx := r.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Update status
	if err := tx.Model(transaction).Update("status", status).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 2. Publish Domain Event (Cash Flow will handle refund creating)
	payload := events.TransactionCreatedPayload{
		TX:          tx,
		Transaction: transaction,
		UserID:      1, // default UserID if context lacks
	}
	if userIDCtx, ok := ctx.Value("userID").(uint); ok {
		payload.UserID = userIDCtx
	}

	if err := r.EventBus.Publish(ctx, eventName, payload); err != nil {
		tx.Rollback()
		return fmt.Errorf("transaction event failed: %w", err)
	}

	// 3. Commit Transaction
	return tx.Commit().Error
}
