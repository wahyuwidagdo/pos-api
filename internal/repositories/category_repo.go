package repositories

import (
	"context"
	"pos-api/internal/models"

	"gorm.io/gorm"
)

// CategoryRepository mendefinisikan kontrak untuk interaksi database kategori.
type CategoryRepository interface {
	CreateCategory(ctx context.Context, category *models.Category) error
	GetCategoryByID(ctx context.Context, id uint) (*models.Category, error)
	GetAllCategories(ctx context.Context, onlyTrashed bool) ([]models.Category, error)
	UpdateCategory(ctx context.Context, category *models.Category) error
	DeleteCategory(ctx context.Context, id uint) error
	RestoreCategory(ctx context.Context, id uint) error
	ForceDeleteCategory(ctx context.Context, id uint) error
}

type categoryRepository struct {
	DB *gorm.DB
}

// NewCategoryRepository membuat instance CategoryRepository baru.
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{DB: db}
}

func (r *categoryRepository) CreateCategory(ctx context.Context, category *models.Category) error {
	result := r.DB.WithContext(ctx).Create(category)
	return result.Error
}

func (r *categoryRepository) GetCategoryByID(ctx context.Context, id uint) (*models.Category, error) {
	var category models.Category
	result := r.DB.WithContext(ctx).First(&category, id)
	return &category, result.Error
}

func (r *categoryRepository) GetAllCategories(ctx context.Context, onlyTrashed bool) ([]models.Category, error) {
	var categories []models.Category
	query := r.DB.WithContext(ctx)

	if onlyTrashed {
		query = query.Unscoped().Where("deleted_at IS NOT NULL")
	}

	if err := query.Find(&categories).Error; err != nil {
		return nil, err
	}

	// Manually populate product_count since gorm:"-" prevents scanning subqueries
	type countResult struct {
		CategoryID   uint  `gorm:"column:category_id"`
		ProductCount int64 `gorm:"column:product_count"`
	}
	var counts []countResult
	r.DB.WithContext(ctx).Model(&models.Product{}).
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

func (r *categoryRepository) UpdateCategory(ctx context.Context, category *models.Category) error {
	result := r.DB.WithContext(ctx).Save(category)
	return result.Error
}

func (r *categoryRepository) DeleteCategory(ctx context.Context, id uint) error {
	// PENTING: Gunakan Unscoped().Delete() jika Anda ingin menghapus permanen,
	// atau gunakan Delete() jika ingin soft delete (rekomendasi default GORM)
	result := r.DB.WithContext(ctx).Delete(&models.Category{}, id)
	return result.Error
}

func (r *categoryRepository) RestoreCategory(ctx context.Context, id uint) error {
	return r.DB.WithContext(ctx).Unscoped().Model(&models.Category{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func (r *categoryRepository) ForceDeleteCategory(ctx context.Context, id uint) error {
	return r.DB.WithContext(ctx).Unscoped().Delete(&models.Category{}, id).Error
}
