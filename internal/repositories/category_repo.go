package repositories

import (
	"pos-api/internal/models"

	"gorm.io/gorm"
)

// CategoryRepository mendefinisikan kontrak untuk interaksi database kategori.
type CategoryRepository interface {
	CreateCategory(category *models.Category) error
	GetCategoryByID(id uint) (*models.Category, error)
	GetAllCategories() ([]models.Category, error)
	UpdateCategory(category *models.Category) error
	DeleteCategory(id uint) error
}

type categoryRepository struct {
	DB *gorm.DB
}

// NewCategoryRepository membuat instance CategoryRepository baru.
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{DB: db}
}

func (r *categoryRepository) CreateCategory(category *models.Category) error {
	result := r.DB.Create(category)
	return result.Error
}

func (r *categoryRepository) GetCategoryByID(id uint) (*models.Category, error) {
	var category models.Category
	result := r.DB.First(&category, id)
	return &category, result.Error
}

func (r *categoryRepository) GetAllCategories() ([]models.Category, error) {
	var categories []models.Category
	if err := r.DB.Find(&categories).Error; err != nil {
		return nil, err
	}

	// Manually populate product_count since gorm:"-" prevents scanning subqueries
	type countResult struct {
		CategoryID   uint  `gorm:"column:category_id"`
		ProductCount int64 `gorm:"column:product_count"`
	}
	var counts []countResult
	r.DB.Model(&models.Product{}).
		Select("category_id, COUNT(*) as product_count").
		Group("category_id").
		Find(&counts)

	countMap := make(map[uint]int64)
	for _, c := range counts {
		countMap[c.CategoryID] = c.ProductCount
	}

	for i := range categories {
		categories[i].ProductCount = countMap[categories[i].ID]
	}

	return categories, nil
}

func (r *categoryRepository) UpdateCategory(category *models.Category) error {
	result := r.DB.Save(category)
	return result.Error
}

func (r *categoryRepository) DeleteCategory(id uint) error {
	// PENTING: Gunakan Unscoped().Delete() jika Anda ingin menghapus permanen,
	// atau gunakan Delete() jika ingin soft delete (rekomendasi default GORM)
	result := r.DB.Delete(&models.Category{}, id)
	return result.Error
}
