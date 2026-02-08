package services

import (
	"errors"
	"time"

	"pos-api/internal/models"
	"pos-api/internal/repositories"

	"gorm.io/gorm"
)

// OpenShiftRequest represents the request to open a shift
type OpenShiftRequest struct {
	StartingCash float64 `json:"starting_cash" validate:"required,gte=0"`
}

// CloseShiftRequest represents the request to close a shift
type CloseShiftRequest struct {
	EndingCash float64 `json:"ending_cash" validate:"required,gte=0"`
	Notes      string  `json:"notes"`
}

// ShiftService defines the contract for shift business logic
type ShiftService interface {
	OpenShift(userID uint, req OpenShiftRequest) (*models.Shift, error)
	CloseShift(userID uint, req CloseShiftRequest) (*models.Shift, error)
	GetCurrentShift() (*models.Shift, error)
	GetShiftByID(id uint) (*models.Shift, error)
	ListShifts(page, pageSize int) ([]models.Shift, error)
}

type shiftService struct {
	repo repositories.ShiftRepository
}

// NewShiftService creates a new shift service
func NewShiftService(repo repositories.ShiftRepository) ShiftService {
	return &shiftService{repo: repo}
}

// OpenShift opens a new shift for a user
func (s *shiftService) OpenShift(userID uint, req OpenShiftRequest) (*models.Shift, error) {
	// Check if there's already an open shift
	existingShift, err := s.repo.GetAnyOpenShift()
	if err == nil && existingShift.ID > 0 {
		return nil, errors.New("sudah ada shift yang sedang berjalan. Tutup shift terlebih dahulu")
	}

	shift := &models.Shift{
		UserID:       userID,
		StartingCash: req.StartingCash,
		Status:       "open",
		OpenedAt:     time.Now(),
	}

	if err := s.repo.CreateShift(shift); err != nil {
		return nil, errors.New("gagal membuka shift: " + err.Error())
	}

	return shift, nil
}

// CloseShift closes the current open shift
func (s *shiftService) CloseShift(userID uint, req CloseShiftRequest) (*models.Shift, error) {
	// Get current open shift
	shift, err := s.repo.GetAnyOpenShift()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("tidak ada shift yang sedang berjalan")
		}
		return nil, errors.New("gagal mengambil data shift")
	}

	// Calculate expected cash from transactions during this shift
	totalSales, totalTx, err := s.repo.GetShiftTransactionSummary(shift.ID, shift.OpenedAt)
	if err != nil {
		return nil, errors.New("gagal menghitung total transaksi shift")
	}

	// Update shift with closing data
	now := time.Now()
	shift.EndingCash = req.EndingCash
	shift.ExpectedCash = shift.StartingCash + totalSales
	shift.CashDiff = req.EndingCash - shift.ExpectedCash
	shift.TotalSales = totalSales
	shift.TotalTx = totalTx
	shift.Status = "closed"
	shift.Notes = req.Notes
	shift.ClosedAt = &now

	if err := s.repo.CloseShift(shift); err != nil {
		return nil, errors.New("gagal menutup shift: " + err.Error())
	}

	// Reload with user relation
	return s.repo.GetShiftByID(shift.ID)
}

// GetCurrentShift gets the currently open shift
func (s *shiftService) GetCurrentShift() (*models.Shift, error) {
	shift, err := s.repo.GetAnyOpenShift()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("tidak ada shift yang sedang berjalan")
		}
		return nil, errors.New("gagal mengambil data shift")
	}

	// Get current transaction summary
	totalSales, totalTx, _ := s.repo.GetShiftTransactionSummary(shift.ID, shift.OpenedAt)
	shift.TotalSales = totalSales
	shift.TotalTx = totalTx
	shift.ExpectedCash = shift.StartingCash + totalSales

	return shift, nil
}

// GetShiftByID retrieves a shift by its ID
func (s *shiftService) GetShiftByID(id uint) (*models.Shift, error) {
	return s.repo.GetShiftByID(id)
}

// ListShifts retrieves a paginated list of shifts
func (s *shiftService) ListShifts(page, pageSize int) ([]models.Shift, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	return s.repo.ListShifts(pageSize, offset)
}
