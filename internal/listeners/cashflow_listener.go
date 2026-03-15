package listeners

import (
	"context"
	"errors"
	"fmt"
	"time"

	"pos-api/internal/models"
	"pos-api/internal/pkg/events"
)

// HandleCashFlowOnTransaction listens for a TransactionCreatedEvent
// and writes an automatic CashFlow 'income' entry mapped to the active gorm TX.
func HandleCashFlowOnTransaction(ctx context.Context, p interface{}) error {
	payload, ok := p.(events.TransactionCreatedPayload)
	if !ok {
		return errors.New("invalid payload type for HandleCashFlowOnTransaction")
	}

	tx := payload.TX
	transaction := payload.Transaction

	cashFlow := models.CashFlow{
		Type:      "income",
		Source:    "sales",
		Amount:    transaction.GrandTotal,
		Date:      transaction.CreatedAt,
		Notes:     "Transaction " + transaction.TransactionCode,
		UserID:    payload.UserID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := tx.Create(&cashFlow).Error; err != nil {
		return fmt.Errorf("failed to create automatic cash flow on transaction %s: %w", transaction.TransactionCode, err)
	}

	return nil
}

// HandleCashFlowOnInventoryAdjusted listens for EventInventoryAdjusted
// and inserts an 'expense' record if the type is 'in' and source is 'purchase' or similar, denoting restock.
func HandleCashFlowOnInventoryAdjusted(ctx context.Context, p interface{}) error {
	payload, ok := p.(events.InventoryAdjustedPayload)
	if !ok {
		return errors.New("invalid payload type for HandleCashFlowOnInventoryAdjusted")
	}

	log := payload.InventoryLog

	// Only care if it's an "in" event that costs money
	if log.Type == "in" && log.TotalCost > 0 {
		var product models.Product
		productName := fmt.Sprintf("ID %d", log.ProductID)
		if err := payload.TX.First(&product, log.ProductID).Error; err == nil && product.Name != "" {
			productName = product.Name
		}

		cashFlow := models.CashFlow{
			Type:      "expense",
			Source:    "penambahan_stok", // Native category for Stock Replenishment
			Amount:    log.TotalCost,
			Date:      time.Now(),
			Notes:     fmt.Sprintf("Restock %s, Qty: %d", productName, log.Quantity),
			UserID:    payload.UserID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := payload.TX.Create(&cashFlow).Error; err != nil {
			return fmt.Errorf("failed to create cash flow for restock: %w", err)
		}
	}

	return nil
}

// HandleCashFlowOnTransactionReverted listens for TransactionReturned or Cancelled events
// and deletes the original income cash flow entry that was created when the transaction was made.
// This keeps income and expenses properly synchronized instead of creating offsetting entries.
func HandleCashFlowOnTransactionReverted(ctx context.Context, p interface{}) error {
	payload, ok := p.(events.TransactionCreatedPayload)
	if !ok {
		return errors.New("invalid payload type for HandleCashFlowOnTransactionReverted")
	}

	tx := payload.TX
	transaction := payload.Transaction

	// Delete the original income cash flow entry matching this transaction
	// The Notes field was set to "Transaction <code>" when the income entry was created
	result := tx.Where("source = ? AND notes = ?", "sales", "Transaction "+transaction.TransactionCode).
		Delete(&models.CashFlow{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete cash flow for reverted transaction %s: %w", transaction.TransactionCode, result.Error)
	}

	return nil
}
