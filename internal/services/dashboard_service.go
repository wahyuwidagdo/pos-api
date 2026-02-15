package services

import (
	"fmt"
	"time"

	"pos-api/internal/repositories"
)

// DashboardResponse represents the complete dashboard data
type DashboardResponse struct {
	TodaySales             float64                           `json:"today_sales"`
	TodayTransactions      int64                             `json:"today_transactions"`
	TodayItemsSold         int64                             `json:"today_items_sold"`
	LowStockCount          int64                             `json:"low_stock_count"`
	TopProducts            []repositories.TopProduct         `json:"top_products"`
	RevenueTrend           []repositories.RevenueData        `json:"revenue_trend"`
	RecentTransactions     []repositories.TransactionSummary `json:"recent_transactions"`
	LowStockProducts       []repositories.LowStockProduct    `json:"low_stock_products"`
	PaymentMethodBreakdown []repositories.PaymentMethodData  `json:"payment_method_breakdown"`
}

// DashboardService defines the contract for dashboard business logic
type DashboardService interface {
	GetDashboard(dateRange string) (*DashboardResponse, error)
}

type dashboardService struct {
	repo repositories.DashboardRepository
}

// NewDashboardService creates a new dashboard service
func NewDashboardService(repo repositories.DashboardRepository) DashboardService {
	return &dashboardService{repo: repo}
}

// parseDateRange converts a range string to start and end time
func parseDateRange(dateRange string) (time.Time, time.Time) {
	now := time.Now()
	switch dateRange {
	case "this_week":
		// Start of the week (Monday)
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start := time.Date(now.Year(), now.Month(), now.Day()-(weekday-1), 0, 0, 0, 0, now.Location())
		end := start.AddDate(0, 0, 7)
		return start, end
	case "this_month":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end := start.AddDate(0, 1, 0)
		return start, end
	case "this_year":
		start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		end := start.AddDate(1, 0, 0)
		return start, end
	default: // "today"
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end := start.Add(24 * time.Hour)
		return start, end
	}
}

// GetDashboard retrieves dashboard data for the given date range
func (s *dashboardService) GetDashboard(dateRange string) (*DashboardResponse, error) {
	startDate, endDate := parseDateRange(dateRange)

	// Get basic stats
	stats, err := s.repo.GetDashboardStats(startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Get top 5 products
	topProducts, err := s.repo.GetTopProducts(startDate, endDate, 5)
	if err != nil {
		return nil, err
	}

	// Get low stock count (threshold: 10 items)
	lowStockCount, err := s.repo.GetLowStockCount(10)
	if err != nil {
		return nil, err
	}

	// Get low stock products (top 10 lowest stock)
	lowStockProducts, err := s.repo.GetLowStockProducts(10, 10)
	if err != nil {
		fmt.Println("Error fetching low stock products:", err)
		lowStockProducts = []repositories.LowStockProduct{}
	}

	// Get revenue trend for the date range
	revenueTrend, err := s.repo.GetRevenueTrend(startDate, endDate.Add(-24*time.Hour))
	if err != nil {
		fmt.Println("Error fetching revenue trend:", err)
		revenueTrend = []repositories.RevenueData{}
	}

	// Get recent transactions (limit 20)
	recentTransactions, err := s.repo.GetRecentTransactions(20)
	if err != nil {
		fmt.Println("Error fetching recent transactions:", err)
		recentTransactions = []repositories.TransactionSummary{}
	}

	// Get payment method breakdown
	paymentMethods, err := s.repo.GetPaymentMethodBreakdown(startDate, endDate)
	if err != nil {
		fmt.Println("Error fetching payment method breakdown:", err)
		paymentMethods = []repositories.PaymentMethodData{}
	}

	return &DashboardResponse{
		TodaySales:             stats.TodaySales,
		TodayTransactions:      stats.TodayTransactions,
		TodayItemsSold:         stats.TodayItemsSold,
		LowStockCount:          lowStockCount,
		TopProducts:            topProducts,
		RevenueTrend:           revenueTrend,
		RecentTransactions:     recentTransactions,
		LowStockProducts:       lowStockProducts,
		PaymentMethodBreakdown: paymentMethods,
	}, nil
}
