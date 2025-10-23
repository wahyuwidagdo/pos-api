package handlers

import (
	"pos-api/internal/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// ProductHandler menangani request HTTP terkait produk.
type ProductHandler struct {
	service services.ProductService
}

// NewProductHandler membuat instance ProductHandler baru.
func NewProductHandler(s services.ProductService) *ProductHandler {
	return &ProductHandler{service: s}
}

// CreateProduct menangani POST /api/v1/products
func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	var req services.ProductRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Input JSON tidak valid"})
	}

	product, err := h.service.CreateProduct(req)
	if err != nil {
		// Asumsikan error dari service adalah error bisnis (validasi, duplikat)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Produk berhasil dibuat",
		"product": product,
	})
}

// GetProduct menangani GET /api/v1/products/:id
func (h *ProductHandler) GetProduct(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID produk tidak valid"})
	}

	product, err := h.service.GetProduct(uint(id))
	if err != nil {
		// Error spesifik dari service (tidak ditemukan)
		if err.Error() == "produk tidak ditemukan" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil produk"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"product": product,
	})
}

// ListProducts menangani GET /api/v1/products
func (h *ProductHandler) ListProducts(c *fiber.Ctx) error {
	// Ambil query parameter 'page' dan 'pageSize'
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.Query("pageSize", "10"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	products, err := h.service.ListProducts(page, pageSize)
	if err != nil {
		// Asumsikan error adalah masalah internal/database
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil daftar produk: " + err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"page":     page,
		"pageSize": pageSize,
		"products": products,
	})
}

// UpdateProduct menangani PUT /api/v1/products/:id
func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {
	// 1. Parse ID dari parameter
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID produk tidak valid"})
	}

	// 2. Parse body request
	var req services.ProductRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Input JSON tidak valid"})
	}

	// 3. Panggil service untuk update
	updatedProduct, err := h.service.UpdateProduct(uint(id), req)
	if err != nil {
		// Penanganan error
		if err.Error() == "produk tidak ditemukan" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		// Asumsikan error lain adalah validasi/bisnis/database
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Gagal mengupdate produk: " + err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Produk berhasil diupdate",
		"product": updatedProduct,
	})
}

// DeleteProduct menangani DELETE /api/v1/products/:id
func (h *ProductHandler) DeleteProduct(c *fiber.Ctx) error {
	// 1. Parse ID dari parameter
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID produk tidak valid"})
	}

	// 2. Panggil service untuk delete
	err = h.service.DeleteProduct(uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menghapus produk: " + err.Error()})
	}

	// 3. Beri respons sukses
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Produk berhasil dihapus",
	})
}
