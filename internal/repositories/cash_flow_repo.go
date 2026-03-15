package repositories

import (
	"context"
	"pos-api/internal/models"
	"time"

	"gorm.io/gorm"
)

type CashFlowRepository interface {
	Create(ctx context.Context, cf *models.CashFlow) error
	Update(ctx context.Context, cf *models.CashFlow) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*models.CashFlow, error)
	GetAll(ctx context.Context, limit, offset int, cfType, source string, startDate, endDate *time.Time) ([]models.CashFlow, int64, error)
	GetSummary(ctx context.Context, startDate, endDate time.Time) (totalCapital, totalIncome, totalExpense float64, err error)
	GetSourceBreakdown(ctx context.Context, startDate, endDate time.Time) ([]CashFlowSourceData, error)
}

type CashFlowSourceData struct {
	Source      string  `json:"source"`
	Type        string  `json:"type"`
	TotalAmount float64 `json:"total_amount"`
}

type cashFlowRepository struct {
	DB *gorm.DB
}

func NewCashFlowRepository(db *gorm.DB) CashFlowRepository {
	return &cashFlowRepository{DB: db}
}

func (r *cashFlowRepository) Create(ctx context.Context, cf *models.CashFlow) error {
	return r.DB.WithContext(ctx).Create(cf).Error
}

func (r *cashFlowRepository) Update(ctx context.Context, cf *models.CashFlow) error {
	return r.DB.WithContext(ctx).Save(cf).Error
}

func (r *cashFlowRepository) Delete(ctx context.Context, id uint) error {
	return r.DB.WithContext(ctx).Delete(&models.CashFlow{}, id).Error
}

func (r *cashFlowRepository) GetByID(ctx context.Context, id uint) (*models.CashFlow, error) {
	var cf models.CashFlow
	err := r.DB.WithContext(ctx).Preload("User").First(&cf, id).Error
	return &cf, err
}

func (r *cashFlowRepository) GetAll(ctx context.Context, limit, offset int, cfType, source string, startDate, endDate *time.Time) ([]models.CashFlow, int64, error) {
	var flows []models.CashFlow
	var total int64

	query := r.DB.WithContext(ctx).Model(&models.CashFlow{})

	if cfType != "" {
		query = query.Where("type = ?", cfType)
	}
	if source != "" {
		query = query.Where("source = ?", source)
	}
	if startDate != nil {
		query = query.Where("date >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("date <= ?", *endDate)
	}

	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("created_at DESC").
		Limit(limit).Offset(offset).
		Preload("User").
		Find(&flows).Error

	return flows, total, err
}

func (r *cashFlowRepository) GetSummary(ctx context.Context, startDate, endDate time.Time) (totalCapital, totalIncome, totalExpense float64, err error) {
	// Get total modal (capital) — cumulative, NOT filtered by date range
	err = r.DB.WithContext(ctx).Model(&models.CashFlow{}).
		Where("type = ? AND source IN ?", "income", []string{"modal_awal", "modal_tambahan"}).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalCapital).Error
	if err != nil {
		return 0, 0, 0, err
	}

	// Get total income
	err = r.DB.WithContext(ctx).Model(&models.CashFlow{}).
		Where("type = ? AND date >= ? AND date <= ?", "income", startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalIncome).Error
	if err != nil {
		return 0, 0, 0, err
	}

	// Get total expense
	err = r.DB.WithContext(ctx).Model(&models.CashFlow{}).
		Where("type = ? AND date >= ? AND date <= ?", "expense", startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalExpense).Error
	if err != nil {
		return 0, 0, 0, err
	}

	return totalCapital, totalIncome, totalExpense, nil
}

func (r *cashFlowRepository) GetSourceBreakdown(ctx context.Context, startDate, endDate time.Time) ([]CashFlowSourceData, error) {
	var results []CashFlowSourceData
	err := r.DB.WithContext(ctx).Model(&models.CashFlow{}).
		Where("date >= ? AND date <= ?", startDate, endDate).
		Select("source, type, SUM(amount) as total_amount").
		Group("source, type").
		Scan(&results).Error
	return results, err
}
