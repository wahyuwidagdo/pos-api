package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"pos-api/internal/models"
	"pos-api/internal/repositories"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	customErrors "pos-api/internal/pkg/errors" // Import custom errors
)

// ProductRequest mendefinisikan DTO untuk membuat atau mengupdate produk.
type ProductRequest struct {
	Name        string  `json:"name" validate:"required,min=3,max=100"`
	SKU         string  `json:"sku"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,gt=0"` // Harus lebih besar dari 0
	Cost        float64 `json:"cost" validate:"gt=0"`           // Harus lebih besar atau sama dengan 0
	Stock       int     `json:"stock" validate:"required,gt=0"`
	CategoryID  uint    `json:"category_id" validate:"required"`
}

// ProductService mendefinisikan kontrak untuk logika bisnis produk.
type ProductService interface {
	CreateProduct(req ProductRequest) (*models.Product, error)
	GetProduct(id uint) (*models.Product, error)
	ListProducts(page, pageSize int) ([]models.Product, error)
	UpdateProduct(id uint, req ProductRequest) (*models.Product, error)
	DeleteProduct(id uint) error
}

type productService struct {
	repo      repositories.ProductRepository
	validator *validator.Validate
}

// NewProductService membuat instance ProductService baru.
func NewProductService(repo repositories.ProductRepository) ProductService {
	return &productService{
		repo:      repo,
		validator: validator.New(),
	}
}

// CreateProduct menangani pembuatan produk baru.
func (s *productService) CreateProduct(req ProductRequest) (*models.Product, error) {
	// 1. Validasi Request DTO
	if err := s.validator.Struct(req); err != nil {
		return nil, errors.New("validasi gagal: " + err.Error())
	}

	// 2. Logika Bisnis: Generate SKU jika kosong
	if req.SKU == "" {
		req.SKU = fmt.Sprintf("SKU-%d", time.Now().UnixNano())
	}

	product := models.Product{
		Name:        req.Name,
		SKU:         req.SKU,
		Description: req.Description,
		Price:       req.Price,
		Cost:        req.Cost,
		Stock:       req.Stock,
		CategoryID:  req.CategoryID,
	}

	// 3. Simpan ke Repository
	if err := s.repo.CreateProduct(&product); err != nil {
		// Pengecekan Duplikat Key (Constraint Conflict)
		if strings.Contains(err.Error(), "unique constrain") || strings.Contains(err.Error(), "duplicate key") {
			return nil, customErrors.ErrConflict // <-- Mengembalikan Custom Error 409
		}
		return nil, errors.New("gagal membuat produk: " + err.Error())
	}

	return &product, nil
}

// GetProduct mendapatkan detail produk berdasarkan ID.
func (s *productService) GetProduct(id uint) (*models.Product, error) {
	product, err := s.repo.GetProductByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErrors.ErrNotFound // <-- Mengembalikan Custom Error 404
		}
		return nil, errors.New("gagal mengambil produk")
	}

	return product, nil
}

func (s *productService) ListProducts(page, pageSize int) ([]models.Product, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	products, err := s.repo.GetAllProducts(pageSize, offset)
	if err != nil {
		return nil, errors.New("gagal mengambil produk")
	}

	return products, nil
}

func (s *productService) UpdateProduct(id uint, req ProductRequest) (*models.Product, error) {
	// 1. Validasi Request DTO
	if err := s.validator.Struct(req); err != nil {
		return nil, errors.New("validasi gagal: " + err.Error())
	}

	// 2. Ambil produk yang akan diupdate
	product, err := s.repo.GetProductByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErrors.ErrNotFound // <-- Mengembalikan Custom Error 404
		}
		return nil, errors.New("gagal mengambil produk untuk di update")
	}

	// 3. Update field-field produk dengan data baru dari request
	product.Name = req.Name
	if req.SKU == "" {
		fmt.Println("INFO: SKU dari request kosong, menggunakan SKU existing: ", product.SKU)
	} else {
		product.SKU = req.SKU
	}
	product.Description = req.Description
	product.Price = req.Price
	product.Cost = req.Cost
	product.Stock = req.Stock
	product.CategoryID = req.CategoryID

	// 4. Simpan perubahan ke repository
	if err := s.repo.UpdateProduct(product); err != nil {
		// PERHATIAN: Di sini kita harus memeriksa apakah error GORM adalah Conflict (Duplikat)
		// Pemeriksaan ini biasanya melibatkan pengecekan string error database (tergantung driver)
		// Asumsi error bisa karena CategoryID tidak valid atau SKU duplikat
		if strings.Contains(err.Error(), "duplikat key value") || strings.Contains(err.Error(), "unique constraint") {
			return nil, customErrors.ErrConflict // <-- Mengembalikan Custom Error 409
		}
		return nil, fmt.Errorf("gagal mengupdate produk: %w", err)
	}

	// Ambil kembali produk dengan relasi yang sudah ter-preload (Category)
	// Agar response-nya lengkap
	updateProduct, err := s.repo.GetProductByID(id)
	if err != nil {
		return product, nil // Kembalikan yang lama jika gagal ambil, tapi log error ini
	}

	return updateProduct, nil
}

func (s *productService) DeleteProduct(id uint) error {
	err := s.repo.DeleteProduct(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return customErrors.ErrNotFound // <-- Mengembalikan Custom Error 404
		}
		// PENTING: Cek Foreign Key Constraint Violation (misal, produk sudah ada di transaksi)
		if strings.Contains(err.Error(), "foreign key constraint") {
			return customErrors.ErrForeignKeyConstraint // <-- Mengembalikan Custom Error
		}
		return errors.New("gagal menghapus produk: " + err.Error())
	}

	return nil
}
