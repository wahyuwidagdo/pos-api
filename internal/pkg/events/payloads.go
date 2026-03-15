package events

import (
	"pos-api/internal/models"

	"gorm.io/gorm"
)

const (
	// EventTransactionCreated is emitted right before a transaction is committed.
	EventTransactionCreated = "transaction.created"

	// EventTransactionReturned is emitted when a transaction is fully returned.
	EventTransactionReturned = "transaction.returned"

	// EventTransactionCancelled is emitted when a transaction is cancelled.
	EventTransactionCancelled = "transaction.cancelled"

	// EventInventoryAdjusted is emitted right before an inventory adjustment (in/out) is committed.
	EventInventoryAdjusted = "inventory.adjusted"
)

// TransactionCreatedPayload is the data passed when a transaction is completed.
type TransactionCreatedPayload struct {
	// TX is the active database transaction. Listeners MUST use this
	// to ensure atomicity with the parent transaction.
	TX *gorm.DB

	// Transaction is the main transaction record just inserted.
	Transaction *models.Transaction

	// UserID is the ID of the user (cashier) who processed the transaction.
	UserID uint
}

// InventoryAdjustedPayload is the data passed when stock is adjusted manually or via restock.
type InventoryAdjustedPayload struct {
	TX           *gorm.DB
	InventoryLog *models.InventoryLog
	UserID       uint
}
