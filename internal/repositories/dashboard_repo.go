package repositories

import (
	"context"
	"pos-api/internal/models"
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

// LowStockProduct represents a product with low stock for the stock alert table
type LowStockProduct struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	SKU   string `json:"sku"`
	Stock int    `json:"stock"`
}

// PaymentMethodData represents payment method breakdown for charts
type PaymentMethodData struct {
	Method string  `json:"method"`
	Count  int64   `json:"count"`
	Total  float64 `json:"total"`
}

// DashboardRepository defines the contract for dashboard data access
type DashboardRepository interface {
	GetDashboardStats(ctx context.Context, startDate, endDate time.Time) (*DashboardStats, error)
	GetTopProducts(ctx context.Context, startDate, endDate time.Time, limit int) ([]TopProduct, error)
	GetLowStockCount(ctx context.Context, threshold int) (int64, error)
	GetLowStockProducts(ctx context.Context, threshold int, limit int) ([]LowStockProduct, error)
	GetRevenueTrend(ctx context.Context, startDate, endDate time.Time) ([]RevenueData, error)
	GetHourlyRevenueTrend(ctx context.Context, startDate, endDate time.Time) ([]RevenueData, error)
	GetRecentTransactions(ctx context.Context, limit int) ([]TransactionSummary, error)
	GetPaymentMethodBreakdown(ctx context.Context, startDate, endDate time.Time) ([]PaymentMethodData, error)
}

type RevenueData struct {
	Date    string  `json:"date"`
	Revenue float64 `json:"revenue"`
}

type TransactionSummary struct {
	ID              uint      `json:"id"`
	TransactionCode string    `json:"transaction_code"`
	CustomerName    string    `json:"customer"`
	TotalAmount     float64   `json:"amount"`
	Status          string    `json:"status"` // E.g., "Completed"
	Time            string    `json:"time"`
	CreatedAt       time.Time `json:"-"`
}

type dashboardRepository struct {
	db *gorm.DB
}

// NewDashboardRepository creates a new dashboard repository
func NewDashboardRepository(db *gorm.DB) DashboardRepository {
	return &dashboardRepository{db: db}
}

// GetDashboardStats retrieves aggregated stats for a date range
func (r *dashboardRepository) GetDashboardStats(ctx context.Context, startDate, endDate time.Time) (*DashboardStats, error) {
	var stats DashboardStats

	// Get sales and transaction count
	err := r.db.WithContext(ctx).Table("transactions").
		Select("COALESCE(SUM(grand_total), 0) as today_sales, COUNT(*) as today_transactions").
		Where("created_at >= ? AND created_at < ? AND status = 'completed'", startDate, endDate).
		Scan(&stats).Error
	if err != nil {
		return nil, err
	}

	// Get items sold count
	var itemsSold int64
	err = r.db.WithContext(ctx).Table("transaction_details").
		Joins("JOIN transactions ON transactions.id = transaction_details.transaction_id").
		Where("transactions.created_at >= ? AND transactions.created_at < ? AND transactions.status = 'completed'", startDate, endDate).
		Select("COALESCE(SUM(transaction_details.quantity), 0)").
		Scan(&itemsSold).Error
	if err != nil {
		return nil, err
	}
	stats.TodayItemsSold = itemsSold

	return &stats, nil
}

// GetTopProducts retrieves the top-selling products for a date range
func (r *dashboardRepository) GetTopProducts(ctx context.Context, startDate, endDate time.Time, limit int) ([]TopProduct, error) {
	var topProducts []TopProduct

	err := r.db.WithContext(ctx).Table("transaction_details").
		Select("transaction_details.product_id, transaction_details.product_name, SUM(transaction_details.quantity) as quantity, SUM(transaction_details.sub_total) as revenue").
		Joins("JOIN transactions ON transactions.id = transaction_details.transaction_id").
		Where("transactions.created_at >= ? AND transactions.created_at < ? AND transactions.status = 'completed'", startDate, endDate).
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
func (r *dashboardRepository) GetLowStockCount(ctx context.Context, threshold int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Product{}).
		Where("stock < ?", threshold).
		Count(&count).Error
	return count, err
}

// GetLowStockProducts returns products below the stock threshold
func (r *dashboardRepository) GetLowStockProducts(ctx context.Context, threshold int, limit int) ([]LowStockProduct, error) {
	var products []LowStockProduct
	err := r.db.WithContext(ctx).Model(&models.Product{}).
		Select("id, name, sku, stock").
		Where("stock < ?", threshold).
		Order("stock ASC").
		Limit(limit).
		Scan(&products).Error
	return products, err
}

// GetRevenueTrend retrieves revenue data between two dates
func (r *dashboardRepository) GetRevenueTrend(ctx context.Context, startDate, endDate time.Time) ([]RevenueData, error) {
	var revenues []RevenueData

	query := `
		SELECT 
			TO_CHAR(d.day, 'Mon DD') as date,
			COALESCE(SUM(t.grand_total), 0) as revenue
		FROM 
			generate_series($1::date, $2::date, '1 day') AS d(day)
		LEFT JOIN 
			transactions t ON DATE(t.created_at) = d.day AND t.status = 'completed'
		GROUP BY 
			d.day
		ORDER BY 
			d.day ASC
	`

	err := r.db.WithContext(ctx).Raw(query, startDate, endDate).Scan(&revenues).Error
	return revenues, err
}

// GetHourlyRevenueTrend retrieves revenue data broken down by hour
func (r *dashboardRepository) GetHourlyRevenueTrend(ctx context.Context, startDate, endDate time.Time) ([]RevenueData, error) {
	var revenues []RevenueData

	// Generate series for every hour in the day (00:00 to 23:00)
	// We subtract 1 second from endDate to ensure we don't include the next day's 00:00 if it's exactly on the boundary,
	// or we handle it via the generate_series limit.
	// Using generate_series with timestamp requires careful handling of the upper bound.
	// We want 00:00 to 23:00.
	query := `
		SELECT 
			TO_CHAR(h.hour, 'HH24:00') as date,
			COALESCE(SUM(t.grand_total), 0) as revenue
		FROM 
			generate_series($1::timestamp, $1::timestamp + interval '23 hours', '1 hour') AS h(hour)
		LEFT JOIN 
			transactions t ON DATE_TRUNC('hour', t.created_at) = h.hour AND t.status = 'completed'
		GROUP BY 
			h.hour
		ORDER BY 
			h.hour ASC
	`

	err := r.db.WithContext(ctx).Raw(query, startDate).Scan(&revenues).Error
	return revenues, err
}

func (r *dashboardRepository) GetRecentTransactions(ctx context.Context, limit int) ([]TransactionSummary, error) {
	var transactions []models.Transaction
	var summaries []TransactionSummary

	err := r.db.WithContext(ctx).Order("created_at DESC").Limit(limit).Find(&transactions).Error
	if err != nil {
		return nil, err
	}

	for _, t := range transactions {
		summaries = append(summaries, TransactionSummary{
			ID:              t.ID,
			TransactionCode: t.TransactionCode,
			CustomerName:    "Umum",
			TotalAmount:     t.GrandTotal,
			Status:          t.Status,
			Time:            t.CreatedAt.Format("03:04 PM"),
			CreatedAt:       t.CreatedAt,
		})
	}

	return summaries, nil
}

// GetPaymentMethodBreakdown retrieves payment method distribution for charts
func (r *dashboardRepository) GetPaymentMethodBreakdown(ctx context.Context, startDate, endDate time.Time) ([]PaymentMethodData, error) {
	var results []PaymentMethodData
	err := r.db.WithContext(ctx).Table("transactions").
		Select("payment_method as method, COUNT(*) as count, COALESCE(SUM(grand_total), 0) as total").
		Where("created_at >= ? AND created_at < ? AND status = 'completed'", startDate, endDate).
		Group("payment_method").
		Order("total DESC").
		Find(&results).Error
	return results, err
}
