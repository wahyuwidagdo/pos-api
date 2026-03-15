package services

import (
	"context"
	"errors"
	"fmt"
	"pos-api/internal/models"
	"pos-api/internal/repositories"
	"time"
)

type StockAdjustmentRequest struct {
	ProductID uint    `json:"product_id" validate:"required"`
	Type      string  `json:"type" validate:"required,oneof=in out adjustment"` // "in", "out", "adjustment"
	Source    string  `json:"source" validate:"required"`                       // "purchase", "return", "damage", "expired", "opname"
	Quantity  int     `json:"quantity" validate:"required,gt=0"`
	CostPrice float64 `json:"cost_price" validate:"gte=0"`
	Notes     string  `json:"notes"`
}

type InventoryLogService interface {
	AdjustStock(ctx context.Context, req StockAdjustmentRequest, userID uint) (*models.InventoryLog, error)
	GetLogsByProduct(ctx context.Context, productID uint, page, pageSize int) ([]models.InventoryLog, int64, error)
	GetAllLogs(ctx context.Context, page, pageSize int, logType, source string, startDate, endDate *time.Time) ([]models.InventoryLog, int64, error)
	GetInventoryStats(ctx context.Context, startDate, endDate *time.Time) (map[string]int64, error)
}

type inventoryLogService struct {
	logRepo     repositories.InventoryLogRepository
	productRepo repositories.ProductRepository
}

func NewInventoryLogService(logRepo repositories.InventoryLogRepository, productRepo repositories.ProductRepository) InventoryLogService {
	return &inventoryLogService{
		logRepo:     logRepo,
		productRepo: productRepo,
	}
}

func (s *inventoryLogService) AdjustStock(ctx context.Context, req StockAdjustmentRequest, userID uint) (*models.InventoryLog, error) {
	// Get current product
	product, err := s.productRepo.GetProductByID(ctx, req.ProductID)
	if err != nil {
		return nil, fmt.Errorf("product with ID %d not found", req.ProductID)
	}

	stockBefore := product.Stock
	var stockAfter int

	switch req.Type {
	case "in":
		stockAfter = stockBefore + req.Quantity
	case "out":
		if stockBefore < req.Quantity {
			return nil, errors.New("insufficient stock for this operation")
		}
		stockAfter = stockBefore - req.Quantity
	case "adjustment":
		// For adjustment, quantity is the NEW absolute stock level
		stockAfter = req.Quantity
		req.Quantity = stockAfter - stockBefore // Store the delta
	default:
		return nil, errors.New("invalid stock operation type")
	}

	// Calculate total cost
	costPrice := req.CostPrice
	if costPrice == 0 {
		costPrice = product.Cost
	}

	absQuantity := req.Quantity
	if absQuantity < 0 {
		absQuantity = -absQuantity
	}

	totalCost := costPrice * float64(absQuantity)

	// Create the inventory log
	log := &models.InventoryLog{
		ProductID:   req.ProductID,
		Type:        req.Type,
		Source:      req.Source,
		Quantity:    req.Quantity,
		CostPrice:   costPrice,
		TotalCost:   totalCost,
		StockBefore: stockBefore,
		StockAfter:  stockAfter,
		Notes:       req.Notes,
		UserID:      userID,
	}

	// Update the product stock before passing to repo
	product.Stock = stockAfter
	if req.Type == "in" && req.CostPrice > 0 {
		product.Cost = req.CostPrice // Update cost price on purchase
	}

	fmt.Println("DEBUG: Calling logRepo.ProcessAdjustment")
	// Process atomically and publish event
	if err := s.logRepo.ProcessAdjustment(ctx, log, product); err != nil {
		fmt.Println("DEBUG: logRepo.ProcessAdjustment failed", err)
		return nil, fmt.Errorf("failed to process inventory adjustment: %w", err)
	}
	fmt.Println("DEBUG: logRepo.ProcessAdjustment succeeded")

	return log, nil
}

func (s *inventoryLogService) GetLogsByProduct(ctx context.Context, productID uint, page, pageSize int) ([]models.InventoryLog, int64, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.logRepo.GetByProductID(ctx, productID, pageSize, offset)
}

func (s *inventoryLogService) GetAllLogs(ctx context.Context, page, pageSize int, logType, source string, startDate, endDate *time.Time) ([]models.InventoryLog, int64, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.logRepo.GetAll(ctx, pageSize, offset, logType, source, startDate, endDate)
}

func (s *inventoryLogService) GetInventoryStats(ctx context.Context, startDate, endDate *time.Time) (map[string]int64, error) {
	return s.logRepo.GetStats(ctx, startDate, endDate)
}
