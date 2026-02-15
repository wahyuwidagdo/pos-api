package repositories

import (
	"pos-api/internal/models"

	"gorm.io/gorm"
)

// ProductRepository mendefinisikan kontrak untuk interaksi database produk.
type ProductRepository interface {
	CreateProduct(product *models.Product) error
	GetProductByID(id uint) (*models.Product, error)
	GetAllProducts(limit, offset int, search string, stockFilter string, sortBy string, sortOrder string) ([]models.Product, int64, error)
	GetLowStockProducts(threshold int) ([]models.Product, error)
	GetStockCounts() (map[string]int64, error)
	UpdateProduct(product *models.Product) error
	DeleteProduct(id uint) error
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

func (r *productRepository) CreateProduct(product *models.Product) error {
	result := r.DB.Create(product)
	return result.Error
}

func (r *productRepository) GetProductByID(id uint) (*models.Product, error) {
	var product models.Product
	// Preload Category untuk mendapatkan data kategori sekalian
	result := r.DB.Preload("Category").First(&product, id)
	return &product, result.Error
}

func (r *productRepository) GetAllProducts(limit, offset int, search string, stockFilter string, sortBy string, sortOrder string) ([]models.Product, int64, error) {
	var products []models.Product
	var totalItems int64

	// Base query for counting and fetching
	query := r.DB.Model(&models.Product{})

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
	if err := query.Count(&totalItems).Error; err != nil {
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

func (r *productRepository) UpdateProduct(product *models.Product) error {
	// Save akan mengupdate semua field, termasuk CategoryID
	result := r.DB.Save(product)
	return result.Error
}

func (r *productRepository) GetLowStockProducts(threshold int) ([]models.Product, error) {
	var products []models.Product
	result := r.DB.Preload("Category").Where("stock <= ?", threshold).Order("stock ASC").Find(&products)
	return products, result.Error
}

func (r *productRepository) DeleteProduct(id uint) error {
	result := r.DB.Delete(&models.Product{}, id)
	return result.Error
}

func (r *productRepository) GetStockCounts() (map[string]int64, error) {
	var all, high, low, out int64

	r.DB.Model(&models.Product{}).Count(&all)
	r.DB.Model(&models.Product{}).Where("stock >= 10").Count(&high)
	r.DB.Model(&models.Product{}).Where("stock > 0 AND stock < 10").Count(&low)
	r.DB.Model(&models.Product{}).Where("stock = 0").Count(&out)

	return map[string]int64{
		"all":  all,
		"high": high,
		"low":  low,
		"out":  out,
	}, nil
}
