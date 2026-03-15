package repositories

import (
	"context"
	"pos-api/internal/models"
	"pos-api/internal/pkg/events"
	"time"

	"gorm.io/gorm"
)

type InventoryLogRepository interface {
	Create(ctx context.Context, log *models.InventoryLog) error
	ProcessAdjustment(ctx context.Context, log *models.InventoryLog, product *models.Product) error
	GetByProductID(ctx context.Context, productID uint, limit, offset int) ([]models.InventoryLog, int64, error)
	GetAll(ctx context.Context, limit, offset int, logType, source string, startDate, endDate *time.Time) ([]models.InventoryLog, int64, error)
	GetStats(ctx context.Context, startDate, endDate *time.Time) (map[string]int64, error)
}

type inventoryLogRepository struct {
	DB       *gorm.DB
	EventBus events.EventBus
}

func NewInventoryLogRepository(db *gorm.DB, eventBus events.EventBus) InventoryLogRepository {
	return &inventoryLogRepository{
		DB:       db,
		EventBus: eventBus,
	}
}

func (r *inventoryLogRepository) Create(ctx context.Context, log *models.InventoryLog) error {
	return r.DB.WithContext(ctx).Create(log).Error
}

func (r *inventoryLogRepository) ProcessAdjustment(ctx context.Context, log *models.InventoryLog, product *models.Product) error {
	tx := r.DB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if rec := recover(); rec != nil {
			tx.Rollback()
		}
	}()

	// Update product stock carefully to avoid cascading save deadlocks on associations
	if err := tx.Model(product).Select("Stock", "Cost").Updates(product).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Create inventory log
	if err := tx.Create(log).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Publish Event
	payload := events.InventoryAdjustedPayload{
		TX:           tx,
		InventoryLog: log,
		UserID:       log.UserID,
	}

	if err := r.EventBus.Publish(ctx, events.EventInventoryAdjusted, payload); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *inventoryLogRepository) GetByProductID(ctx context.Context, productID uint, limit, offset int) ([]models.InventoryLog, int64, error) {
	var logs []models.InventoryLog
	var total int64

	query := r.DB.WithContext(ctx).Model(&models.InventoryLog{}).Where("product_id = ?", productID)

	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").
		Limit(limit).Offset(offset).
		Preload("Product").Preload("User").
		Find(&logs).Error

	return logs, total, err
}

func (r *inventoryLogRepository) GetAll(ctx context.Context, limit, offset int, logType, source string, startDate, endDate *time.Time) ([]models.InventoryLog, int64, error) {
	var logs []models.InventoryLog
	var total int64

	query := r.DB.WithContext(ctx).Model(&models.InventoryLog{})

	if logType != "" {
		query = query.Where("type = ?", logType)
	}
	if source != "" {
		query = query.Where("source = ?", source)
	}
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").
		Limit(limit).Offset(offset).
		Preload("Product").Preload("User").
		Find(&logs).Error

	return logs, total, err
}

func (r *inventoryLogRepository) GetStats(ctx context.Context, startDate, endDate *time.Time) (map[string]int64, error) {
	var results []struct {
		Type  string
		Count int64
	}

	query := r.DB.WithContext(ctx).Model(&models.InventoryLog{})
	if startDate != nil {
		query = query.Where("created_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", *endDate)
	}

	// Count logs by type
	err := query.Select("type, count(*) as count").
		Group("type").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	stats := make(map[string]int64)
	// Initialize with 0
	stats["in"] = 0
	stats["out"] = 0
	stats["adjustment"] = 0

	for _, r := range results {
		stats[r.Type] = r.Count
	}

	return stats, nil
}
