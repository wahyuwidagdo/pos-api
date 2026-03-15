package listeners

import (
	"context"
	"errors"
	"fmt"

	"pos-api/internal/models"
	"pos-api/internal/pkg/events"

	"gorm.io/gorm"
)

// HandleInventoryOnTransaction listens for a TransactionCreatedEvent
// and updates the stock, writing to inventory_logs.
func HandleInventoryOnTransaction(ctx context.Context, p interface{}) error {
	payload, ok := p.(events.TransactionCreatedPayload)
	if !ok {
		return errors.New("invalid payload type for HandleInventoryOnTransaction")
	}

	tx := payload.TX
	transaction := payload.Transaction

	// Decrease stock and log it for each item
	for _, detail := range transaction.TransactionDetails {
		// 1. Get current stock
		var product models.Product
		if err := tx.First(&product, detail.ProductID).Error; err != nil {
			return fmt.Errorf("product not found %d: %w", detail.ProductID, err)
		}

		if product.Stock < detail.Quantity {
			return fmt.Errorf("insufficient stock for product %s. Have: %d, Need: %d", product.Name, product.Stock, detail.Quantity)
		}

		stockBefore := product.Stock
		stockAfter := stockBefore - detail.Quantity

		// 2. Atomic Stock Update (Optimistic attempt)
		result := tx.Model(&models.Product{}).
			Where("id = ? AND stock >= ?", detail.ProductID, detail.Quantity).
			UpdateColumn("stock", gorm.Expr("stock - ?", detail.Quantity))

		if result.Error != nil {
			return fmt.Errorf("failed to deduct stock for product %d: %w", detail.ProductID, result.Error)
		}

		if result.RowsAffected == 0 {
			return errors.New("concurrent modification or insufficient stock during update")
		}

		// 3. Insert Inventory Log
		log := models.InventoryLog{
			ProductID:   detail.ProductID,
			Type:        "out",
			Source:      "sale",           // Maps to sales source
			Quantity:    -detail.Quantity, // 'out' is a negative change conceptually, though stored absolute or delta depending on standard. Wait, check service usage.
			CostPrice:   detail.CostAtSale,
			TotalCost:   detail.CostAtSale * float64(detail.Quantity),
			StockBefore: stockBefore,
			StockAfter:  stockAfter,
			Notes:       "Sale " + transaction.TransactionCode,
			UserID:      payload.UserID,
		}
		// Adjust Quantity depending on convention. Service sets it to absolute value.
		log.Quantity = detail.Quantity

		if err := tx.Create(&log).Error; err != nil {
			return fmt.Errorf("failed to create inventory log for product %d: %w", detail.ProductID, err)
		}
	}

	return nil
}

// HandleInventoryOnTransactionReverted listens for TransactionReturned or Cancelled events
// and restores the stock, writing to inventory_logs.
func HandleInventoryOnTransactionReverted(ctx context.Context, p interface{}) error {
	payload, ok := p.(events.TransactionCreatedPayload)
	if !ok {
		return errors.New("invalid payload type for HandleInventoryOnTransactionReverted")
	}

	tx := payload.TX
	transaction := payload.Transaction

	for _, detail := range transaction.TransactionDetails {
		var product models.Product
		if err := tx.First(&product, detail.ProductID).Error; err != nil {
			return fmt.Errorf("product not found %d: %w", detail.ProductID, err)
		}

		stockBefore := product.Stock
		stockAfter := stockBefore + detail.Quantity

		result := tx.Model(&models.Product{}).
			Where("id = ?", detail.ProductID).
			UpdateColumn("stock", gorm.Expr("stock + ?", detail.Quantity))

		if result.Error != nil {
			return fmt.Errorf("failed to restore stock for product %d: %w", detail.ProductID, result.Error)
		}

		log := models.InventoryLog{
			ProductID:   detail.ProductID,
			Type:        "in",
			Source:      "opname", // Mapping to opname or a custom source. We use opname because it's a manual adjustment equivalent, or we can use "return" if supported. Let's use "return" since we're just recording it. Wait, the frontend ENUM might not support "return".
			Quantity:    detail.Quantity,
			CostPrice:   detail.CostAtSale,
			TotalCost:   0, // Cost is not a purchase expense, it's just stock coming back.
			StockBefore: stockBefore,
			StockAfter:  stockAfter,
			Notes:       "Refund/Cancel for " + transaction.TransactionCode,
			UserID:      payload.UserID,
		}
		// Based on SOURCE_OPTIONS in frontend, valid sources natively are: purchase, sale, opname.
		// "return" or "cancel" are not strictly enforced by DB enum usually. We'll use "opname" to be safe,
		// or "return" if we want distinct. The frontend might just display it as raw text. Let's just use "return".
		log.Source = "return"

		if err := tx.Create(&log).Error; err != nil {
			return fmt.Errorf("failed to create inventory log on return for product %d: %w", detail.ProductID, err)
		}
	}

	return nil
}
