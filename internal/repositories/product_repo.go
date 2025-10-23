package repositories

import (
	"pos-api/internal/models"

	"gorm.io/gorm"
)

// ProductRepository mendefinisikan kontrak untuk interaksi database produk.
type ProductRepository interface {
	CreateProduct(product *models.Product) error
	GetProductByID(id uint) (*models.Product, error)
	GetAllProducts(limit, offset int) ([]models.Product, error)
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

func (r *productRepository) GetAllProducts(limit, offset int) ([]models.Product, error) {
	var products []models.Product
	// Preload Category dan terapkan Pagination
	result := r.DB.Limit(limit).Offset(offset).Preload("Category").Find(&products)
	return products, result.Error
}

func (r *productRepository) UpdateProduct(product *models.Product) error {
	// Save akan mengupdate semua field, termasuk CategoryID
	result := r.DB.Save(product)
	return result.Error
}

func (r *productRepository) DeleteProduct(id uint) error {
	result := r.DB.Delete(&models.Product{}, id)
	// gorm.ErrRecordNotFound akan dikembalikan jika tidak ada record yang dihapus
	return result.Error
}
