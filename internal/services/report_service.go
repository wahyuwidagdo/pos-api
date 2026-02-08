package services

import (
	"errors"
	"time"

	"pos-api/internal/repositories"
)

// SalesReportRequest represents the request for sales report
type SalesReportRequest struct {
	StartDate string `json:"start_date" validate:"required"` // Format: YYYY-MM-DD
	EndDate   string `json:"end_date" validate:"required"`   // Format: YYYY-MM-DD
}

// SalesReportResponse represents the complete sales report
type SalesReportResponse struct {
	Summary    *repositories.SalesSummary   `json:"summary"`
	DailyData  []repositories.SalesReport   `json:"daily_data"`
	StartDate  string                       `json:"start_date"`
	EndDate    string                       `json:"end_date"`
}

// ProductReportResponse represents the product performance report
type ProductReportResponse struct {
	Products  []repositories.ProductReport `json:"products"`
	StartDate string                       `json:"start_date"`
	EndDate   string                       `json:"end_date"`
}

// ReportService defines the contract for report business logic
type ReportService interface {
	GetSalesReport(startDate, endDate string) (*SalesReportResponse, error)
	GetProductReport(startDate, endDate string, limit int) (*ProductReportResponse, error)
}

type reportService struct {
	repo repositories.ReportRepository
}

// NewReportService creates a new report service
func NewReportService(repo repositories.ReportRepository) ReportService {
	return &reportService{repo: repo}
}

// GetSalesReport retrieves the sales report for a date range
func (s *reportService) GetSalesReport(startDateStr, endDateStr string) (*SalesReportResponse, error) {
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return nil, errors.New("format tanggal mulai tidak valid (gunakan YYYY-MM-DD)")
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return nil, errors.New("format tanggal akhir tidak valid (gunakan YYYY-MM-DD)")
	}

	if endDate.Before(startDate) {
		return nil, errors.New("tanggal akhir harus setelah tanggal mulai")
	}

	// Get summary
	summary, err := s.repo.GetSalesSummary(startDate, endDate)
	if err != nil {
		return nil, errors.New("gagal mengambil ringkasan penjualan")
	}

	// Get daily data
	dailyData, err := s.repo.GetSalesReport(startDate, endDate)
	if err != nil {
		return nil, errors.New("gagal mengambil data penjualan harian")
	}

	return &SalesReportResponse{
		Summary:   summary,
		DailyData: dailyData,
		StartDate: startDateStr,
		EndDate:   endDateStr,
	}, nil
}

// GetProductReport retrieves the product performance report for a date range
func (s *reportService) GetProductReport(startDateStr, endDateStr string, limit int) (*ProductReportResponse, error) {
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return nil, errors.New("format tanggal mulai tidak valid (gunakan YYYY-MM-DD)")
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return nil, errors.New("format tanggal akhir tidak valid (gunakan YYYY-MM-DD)")
	}

	if endDate.Before(startDate) {
		return nil, errors.New("tanggal akhir harus setelah tanggal mulai")
	}

	// Default limit to 20 if not specified
	if limit <= 0 {
		limit = 20
	}

	products, err := s.repo.GetProductReport(startDate, endDate, limit)
	if err != nil {
		return nil, errors.New("gagal mengambil laporan produk")
	}

	return &ProductReportResponse{
		Products:  products,
		StartDate: startDateStr,
		EndDate:   endDateStr,
	}, nil
}
