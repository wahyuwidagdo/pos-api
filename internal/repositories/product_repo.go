package repositories

import (
	"context"
	"pos-api/internal/models"

	"gorm.io/gorm"
)

// ProductRepository mendefinisikan kontrak untuk interaksi database produk.
type ProductRepository interface {
	CreateProduct(ctx context.Context, product *models.Product) error
	GetProductByID(ctx context.Context, id uint) (*models.Product, error)
	GetAllProducts(ctx context.Context, limit, offset int, search string, stockFilter string, sortBy string, sortOrder string, onlyTrashed bool) ([]models.Product, int64, error)
	GetLowStockProducts(ctx context.Context, threshold int) ([]models.Product, error)
	GetStockCounts(ctx context.Context) (map[string]int64, error)
	UpdateProduct(ctx context.Context, product *models.Product) error
	DeleteProduct(ctx context.Context, id uint) error
	RestoreProduct(ctx context.Context, id uint) error
	ForceDeleteProduct(ctx context.Context, id uint) error
}

type productRepository struct {
	DB *gorm.DB
}

// NewProductRepository membuat instance ProductRepository baru.
func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{
		DB: db,
	}
}

func (r *productRepository) CreateProduct(ctx context.Context, product *models.Product) error {
	result := r.DB.WithContext(ctx).Create(product)
	return result.Error
}

func (r *productRepository) GetProductByID(ctx context.Context, id uint) (*models.Product, error) {
	var product models.Product
	// Preload Category untuk mendapatkan data kategori sekalian
	result := r.DB.WithContext(ctx).Preload("Category").First(&product, id)
	return &product, result.Error
}

func (r *productRepository) GetAllProducts(ctx context.Context, limit, offset int, search string, stockFilter string, sortBy string, sortOrder string, onlyTrashed bool) ([]models.Product, int64, error) {
	var products []models.Product
	var totalItems int64

	// Base query for counting and fetching
	query := r.DB.WithContext(ctx).Model(&models.Product{})

	if onlyTrashed {
		query = query.Unscoped().Where("deleted_at IS NOT NULL")
	}

	if search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where("name ILIKE ? OR sku ILIKE ?", searchTerm, searchTerm)
	}

	// Apply stock filter
	switch stockFilter {
	case "low":
		query = query.Where("stock > 0 AND stock < 10")
	case "out":
		query = query.Where("stock <= 0")
	case "high":
		query = query.Where("stock >= 10")
	}

	// Hitung total items sebelum limit/offset
	// Gunakan Session(&gorm.Session{}) untuk clone query agar tidak mengubah state query asli
	if err := query.Session(&gorm.Session{}).Count(&totalItems).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	allowedSortColumns := map[string]bool{
		"name": true, "stock": true, "price": true, "cost": true, "created_at": true,
	}
	if allowedSortColumns[sortBy] {
		if sortOrder != "asc" && sortOrder != "desc" {
			sortOrder = "asc"
		}
		query = query.Order(sortBy + " " + sortOrder)
	} else {
		query = query.Order("created_at DESC")
	}

	// Fetch data dengan limit/offset dan preload
	err := query.Limit(limit).Offset(offset).Preload("Category").Find(&products).Error
	return products, totalItems, err
}

func (r *productRepository) UpdateProduct(ctx context.Context, product *models.Product) error {
	// Save akan mengupdate semua field, termasuk CategoryID
	result := r.DB.WithContext(ctx).Save(product)
	return result.Error
}

func (r *productRepository) GetLowStockProducts(ctx context.Context, threshold int) ([]models.Product, error) {
	var products []models.Product
	result := r.DB.WithContext(ctx).Preload("Category").Where("stock <= ?", threshold).Order("stock ASC").Find(&products)
	return products, result.Error
}

func (r *productRepository) DeleteProduct(ctx context.Context, id uint) error {
	result := r.DB.WithContext(ctx).Delete(&models.Product{}, id)
	return result.Error
}

func (r *productRepository) RestoreProduct(ctx context.Context, id uint) error {
	result := r.DB.WithContext(ctx).Unscoped().Model(&models.Product{}).Where("id = ?", id).Update("deleted_at", nil)
	return result.Error
}

func (r *productRepository) ForceDeleteProduct(ctx context.Context, id uint) error {
	result := r.DB.WithContext(ctx).Unscoped().Delete(&models.Product{}, id)
	return result.Error
}

func (r *productRepository) GetStockCounts(ctx context.Context) (map[string]int64, error) {
	var all, high, low, out int64

	r.DB.WithContext(ctx).Model(&models.Product{}).Count(&all)
	r.DB.WithContext(ctx).Model(&models.Product{}).Where("stock >= 10").Count(&high)
	r.DB.WithContext(ctx).Model(&models.Product{}).Where("stock > 0 AND stock < 10").Count(&low)
	r.DB.WithContext(ctx).Model(&models.Product{}).Where("stock = 0").Count(&out)

	return map[string]int64{
		"all":  all,
		"high": high,
		"low":  low,
		"out":  out,
	}, nil
}
