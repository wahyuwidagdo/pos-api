package repositories

import (
	"time"

	"pos-api/internal/models"

	"gorm.io/gorm"
)

// ShiftRepository defines the contract for shift data access
type ShiftRepository interface {
	CreateShift(shift *models.Shift) error
	GetShiftByID(id uint) (*models.Shift, error)
	GetCurrentOpenShift(userID uint) (*models.Shift, error)
	GetAnyOpenShift() (*models.Shift, error)
	CloseShift(shift *models.Shift) error
	GetShiftTransactionSummary(shiftID uint, openedAt time.Time) (float64, int, error)
	ListShifts(limit, offset int) ([]models.Shift, error)
}

type shiftRepository struct {
	db *gorm.DB
}

// NewShiftRepository creates a new shift repository
func NewShiftRepository(db *gorm.DB) ShiftRepository {
	return &shiftRepository{db: db}
}

// CreateShift creates a new shift record
func (r *shiftRepository) CreateShift(shift *models.Shift) error {
	return r.db.Create(shift).Error
}

// GetShiftByID retrieves a shift by ID with user relation
func (r *shiftRepository) GetShiftByID(id uint) (*models.Shift, error) {
	var shift models.Shift
	err := r.db.Preload("User").First(&shift, id).Error
	return &shift, err
}

// GetCurrentOpenShift gets the currently open shift for a specific user
func (r *shiftRepository) GetCurrentOpenShift(userID uint) (*models.Shift, error) {
	var shift models.Shift
	err := r.db.Preload("User").
		Where("user_id = ? AND status = ?", userID, "open").
		First(&shift).Error
	return &shift, err
}

// GetAnyOpenShift gets any currently open shift (for checking if store has open shift)
func (r *shiftRepository) GetAnyOpenShift() (*models.Shift, error) {
	var shift models.Shift
	err := r.db.Preload("User").
		Where("status = ?", "open").
		First(&shift).Error
	return &shift, err
}

// CloseShift updates a shift with closing information
func (r *shiftRepository) CloseShift(shift *models.Shift) error {
	return r.db.Save(shift).Error
}

// GetShiftTransactionSummary calculates total sales and transaction count for a shift
func (r *shiftRepository) GetShiftTransactionSummary(shiftID uint, openedAt time.Time) (float64, int, error) {
	var result struct {
		TotalSales float64
		TotalTx    int
	}

	err := r.db.Table("transactions").
		Select("COALESCE(SUM(grand_total), 0) as total_sales, COUNT(*) as total_tx").
		Where("created_at >= ?", openedAt).
		Scan(&result).Error

	return result.TotalSales, result.TotalTx, err
}

// ListShifts retrieves a paginated list of shifts
func (r *shiftRepository) ListShifts(limit, offset int) ([]models.Shift, error) {
	var shifts []models.Shift
	err := r.db.Preload("User").
		Order("opened_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&shifts).Error
	return shifts, err
}
