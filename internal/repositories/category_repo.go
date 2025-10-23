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
	result := r.DB.Find(&categories)
	return categories, result.Error
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
