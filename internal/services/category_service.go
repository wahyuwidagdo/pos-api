package services

import (
	"errors"
	"fmt"

	"pos-api/internal/models"
	"pos-api/internal/repositories"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// CategoryRequest mendefinisikan DTO untuk kategori.
type CategoryRequest struct {
	Name string `json:"name" validate:"required,min=3,max=50"`
}

// CategoryService mendefinisikan kontrak untuk logika bisnis kategori.
type CategoryService interface {
	CreateCategory(req CategoryRequest) (*models.Category, error)
	GetCategory(id uint) (*models.Category, error)
	ListCategories() ([]models.Category, error)
	UpdateCategory(id uint, req CategoryRequest) (*models.Category, error)
	DeleteCategory(id uint) error
}

type categoryService struct {
	repo      repositories.CategoryRepository
	validator *validator.Validate
}

func NewCategoryService(repo repositories.CategoryRepository) CategoryService {
	return &categoryService{
		repo:      repo,
		validator: validator.New(),
	}
}

func (s *categoryService) CreateCategory(req CategoryRequest) (*models.Category, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, errors.New("validasi gagal: Nama kategori wajib diisi (min 3 karakter)")
	}

	category := models.Category{
		Name: req.Name,
	}

	if err := s.repo.CreateCategory(&category); err != nil {
		// Asumsi: Jika error adalah duplikat/constraint, GORM akan memberikan error.
		return nil, fmt.Errorf("gagal membuat kategori: %w", err)
	}

	return &category, nil
}

func (s *categoryService) GetCategory(id uint) (*models.Category, error) {
	category, err := s.repo.GetCategoryByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("kategori tidak ditemukan")
		}

		return nil, fmt.Errorf("gagal mengambil kategori: %w", err)
	}

	return category, nil
}

func (s *categoryService) ListCategories() ([]models.Category, error) {
	categories, err := s.repo.GetAllCategories()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar kategori: %w", err)
	}

	return categories, nil
}

func (s *categoryService) UpdateCategory(id uint, req CategoryRequest) (*models.Category, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, errors.New("validasi gagal: Nama kategori wajib diisi (min 3 karakter)")
	}

	category, err := s.repo.GetCategoryByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("kategori tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal mengambil kategori untuk diupdate: %w", err)
	}

	category.Name = req.Name

	if err := s.repo.UpdateCategory(category); err != nil {
		return nil, fmt.Errorf("gagal mengupdate kategori: %w", err)
	}

	return category, nil
}

func (s *categoryService) DeleteCategory(id uint) error {
	// Catatan: Dalam aplikasi riil, Anda harus cek apakah kategori ini digunakan oleh produk lain sebelum dihapus.
	// Untuk MVP, kita biarkan database constraint (foreign key) yang menangani errornya.
	err := s.repo.DeleteCategory(id)
	if err != nil {
		return fmt.Errorf("gagal menghapus kategori: %w", err)
	}

	return nil
}
