package services

import (
	"time"

	"pos-api/internal/repositories"
)

// DashboardResponse represents the complete dashboard data
type DashboardResponse struct {
	TodaySales        float64                   `json:"today_sales"`
	TodayTransactions int64                     `json:"today_transactions"`
	TodayItemsSold    int64                     `json:"today_items_sold"`
	LowStockCount     int64                     `json:"low_stock_count"`
	TopProducts       []repositories.TopProduct `json:"top_products"`
}

// DashboardService defines the contract for dashboard business logic
type DashboardService interface {
	GetDashboard() (*DashboardResponse, error)
}

type dashboardService struct {
	repo repositories.DashboardRepository
}

// NewDashboardService creates a new dashboard service
func NewDashboardService(repo repositories.DashboardRepository) DashboardService {
	return &dashboardService{repo: repo}
}

// GetDashboard retrieves today's dashboard data
func (s *dashboardService) GetDashboard() (*DashboardResponse, error) {
	today := time.Now()

	// Get basic stats
	stats, err := s.repo.GetDashboardStats(today)
	if err != nil {
		return nil, err
	}

	// Get top 5 products
	topProducts, err := s.repo.GetTopProducts(today, 5)
	if err != nil {
		return nil, err
	}

	// Get low stock count (threshold: 10 items)
	lowStockCount, err := s.repo.GetLowStockCount(10)
	if err != nil {
		return nil, err
	}

	return &DashboardResponse{
		TodaySales:        stats.TodaySales,
		TodayTransactions: stats.TodayTransactions,
		TodayItemsSold:    stats.TodayItemsSold,
		LowStockCount:     lowStockCount,
		TopProducts:       topProducts,
	}, nil
}
