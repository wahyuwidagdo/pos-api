package repositories

import (
	"time"

	"gorm.io/gorm"
)

// DashboardStats represents the aggregated statistics for the dashboard
type DashboardStats struct {
	TodaySales        float64 `json:"today_sales"`
	TodayTransactions int64   `json:"today_transactions"`
	TodayItemsSold    int64   `json:"today_items_sold"`
	TodayProfit       float64 `json:"today_profit"`
	LowStockCount     int64   `json:"low_stock_count"`
}

// TopProduct represents a top-selling product
type TopProduct struct {
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int64   `json:"quantity"`
	Revenue     float64 `json:"revenue"`
}

// DashboardRepository defines the contract for dashboard data access
type DashboardRepository interface {
	GetDashboardStats(date time.Time) (*DashboardStats, error)
	GetTopProducts(date time.Time, limit int) ([]TopProduct, error)
	GetLowStockCount(threshold int) (int64, error)
}

type dashboardRepository struct {
	db *gorm.DB
}

// NewDashboardRepository creates a new dashboard repository
func NewDashboardRepository(db *gorm.DB) DashboardRepository {
	return &dashboardRepository{db: db}
}

// GetDashboardStats retrieves aggregated stats for a specific date
func (r *dashboardRepository) GetDashboardStats(date time.Time) (*DashboardStats, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var stats DashboardStats

	// Get today's sales and transaction count
	err := r.db.Table("transactions").
		Select("COALESCE(SUM(grand_total), 0) as today_sales, COUNT(*) as today_transactions").
		Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).
		Scan(&stats).Error
	if err != nil {
		return nil, err
	}

	// Get today's items sold count
	var itemsSold int64
	err = r.db.Table("transaction_details").
		Joins("JOIN transactions ON transactions.id = transaction_details.transaction_id").
		Where("transactions.created_at >= ? AND transactions.created_at < ?", startOfDay, endOfDay).
		Select("COALESCE(SUM(transaction_details.quantity), 0)").
		Scan(&itemsSold).Error
	if err != nil {
		return nil, err
	}
	stats.TodayItemsSold = itemsSold

	return &stats, nil
}

// GetTopProducts retrieves the top-selling products for a specific date
func (r *dashboardRepository) GetTopProducts(date time.Time, limit int) ([]TopProduct, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var topProducts []TopProduct

	err := r.db.Table("transaction_details").
		Select("transaction_details.product_id, transaction_details.product_name, SUM(transaction_details.quantity) as quantity, SUM(transaction_details.sub_total) as revenue").
		Joins("JOIN transactions ON transactions.id = transaction_details.transaction_id").
		Where("transactions.created_at >= ? AND transactions.created_at < ?", startOfDay, endOfDay).
		Group("transaction_details.product_id, transaction_details.product_name").
		Order("quantity DESC").
		Limit(limit).
		Scan(&topProducts).Error
	if err != nil {
		return nil, err
	}

	return topProducts, nil
}

// GetLowStockCount returns the count of products below the threshold
func (r *dashboardRepository) GetLowStockCount(threshold int) (int64, error) {
	var count int64
	err := r.db.Table("products").
		Where("stock <= ?", threshold).
		Count(&count).Error
	return count, err
}
