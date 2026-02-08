package repositories

import (
	"time"

	"gorm.io/gorm"
)

// SalesReport represents a sales report entry
type SalesReport struct {
	Date              string  `json:"date"`
	TotalSales        float64 `json:"total_sales"`
	TotalTransactions int64   `json:"total_transactions"`
	TotalItemsSold    int64   `json:"total_items_sold"`
	AverageTransaction float64 `json:"average_transaction"`
}

// ProductReport represents a product performance report
type ProductReport struct {
	ProductID    uint    `json:"product_id"`
	ProductName  string  `json:"product_name"`
	CategoryName string  `json:"category_name"`
	TotalSold    int64   `json:"total_sold"`
	TotalRevenue float64 `json:"total_revenue"`
	CurrentStock int     `json:"current_stock"`
}

// ReportRepository defines the contract for report data access
type ReportRepository interface {
	GetSalesReport(startDate, endDate time.Time) ([]SalesReport, error)
	GetProductReport(startDate, endDate time.Time, limit int) ([]ProductReport, error)
	GetSalesSummary(startDate, endDate time.Time) (*SalesSummary, error)
}

// SalesSummary represents the summary of sales for a period
type SalesSummary struct {
	TotalSales        float64 `json:"total_sales"`
	TotalTransactions int64   `json:"total_transactions"`
	TotalItemsSold    int64   `json:"total_items_sold"`
	AveragePerDay     float64 `json:"average_per_day"`
}

type reportRepository struct {
	db *gorm.DB
}

// NewReportRepository creates a new report repository
func NewReportRepository(db *gorm.DB) ReportRepository {
	return &reportRepository{db: db}
}

// GetSalesReport retrieves daily sales data for a date range
func (r *reportRepository) GetSalesReport(startDate, endDate time.Time) ([]SalesReport, error) {
	var reports []SalesReport

	err := r.db.Table("transactions").
		Select(`
			DATE(created_at) as date,
			COALESCE(SUM(grand_total), 0) as total_sales,
			COUNT(*) as total_transactions
		`).
		Where("created_at >= ? AND created_at < ?", startDate, endDate.Add(24*time.Hour)).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&reports).Error
	if err != nil {
		return nil, err
	}

	// Get items sold per day
	for i := range reports {
		var itemsSold int64
		dateStr := reports[i].Date
		r.db.Table("transaction_details").
			Joins("JOIN transactions ON transactions.id = transaction_details.transaction_id").
			Where("DATE(transactions.created_at) = ?", dateStr).
			Select("COALESCE(SUM(transaction_details.quantity), 0)").
			Scan(&itemsSold)
		reports[i].TotalItemsSold = itemsSold
		if reports[i].TotalTransactions > 0 {
			reports[i].AverageTransaction = reports[i].TotalSales / float64(reports[i].TotalTransactions)
		}
	}

	return reports, nil
}

// GetProductReport retrieves product sales performance for a date range
func (r *reportRepository) GetProductReport(startDate, endDate time.Time, limit int) ([]ProductReport, error) {
	var reports []ProductReport

	query := r.db.Table("transaction_details").
		Select(`
			transaction_details.product_id,
			transaction_details.product_name,
			COALESCE(categories.name, 'Uncategorized') as category_name,
			SUM(transaction_details.quantity) as total_sold,
			SUM(transaction_details.sub_total) as total_revenue,
			COALESCE(products.stock, 0) as current_stock
		`).
		Joins("JOIN transactions ON transactions.id = transaction_details.transaction_id").
		Joins("LEFT JOIN products ON products.id = transaction_details.product_id").
		Joins("LEFT JOIN categories ON categories.id = products.category_id").
		Where("transactions.created_at >= ? AND transactions.created_at < ?", startDate, endDate.Add(24*time.Hour)).
		Group("transaction_details.product_id, transaction_details.product_name, categories.name, products.stock").
		Order("total_sold DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Scan(&reports).Error
	if err != nil {
		return nil, err
	}

	return reports, nil
}

// GetSalesSummary retrieves the overall summary for a date range
func (r *reportRepository) GetSalesSummary(startDate, endDate time.Time) (*SalesSummary, error) {
	var summary SalesSummary

	err := r.db.Table("transactions").
		Select("COALESCE(SUM(grand_total), 0) as total_sales, COUNT(*) as total_transactions").
		Where("created_at >= ? AND created_at < ?", startDate, endDate.Add(24*time.Hour)).
		Scan(&summary).Error
	if err != nil {
		return nil, err
	}

	// Get total items sold
	var itemsSold int64
	r.db.Table("transaction_details").
		Joins("JOIN transactions ON transactions.id = transaction_details.transaction_id").
		Where("transactions.created_at >= ? AND transactions.created_at < ?", startDate, endDate.Add(24*time.Hour)).
		Select("COALESCE(SUM(transaction_details.quantity), 0)").
		Scan(&itemsSold)
	summary.TotalItemsSold = itemsSold

	// Calculate days in range
	days := endDate.Sub(startDate).Hours() / 24
	if days > 0 {
		summary.AveragePerDay = summary.TotalSales / days
	}

	return &summary, nil
}
