package handlers

import (
	"pos-api/internal/services"

	"github.com/gofiber/fiber/v2"

	customErrors "pos-api/internal/pkg/errors" // Import custom errors
)

type CategoryHandler struct {
	service services.CategoryService
}

func NewCategoryHandler(s services.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: s}
}

// CreateCategory handles POST /categories
func (h *CategoryHandler) CreateCategory(c *fiber.Ctx) error {
	var req services.CategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Input JSON tidak valid"})
	}

	category, err := h.service.CreateCategory(req)
	if err != nil {
		if customErrors.Is(err, customErrors.ErrConflict) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Nama kategori sudah ada (duplikat)."}) // 409
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Gagal membuat kategori: " + err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Kategori berhasil dibuat",
		"data":    category,
	})
}

// GetCategory handles GET /categories/:id
func (h *CategoryHandler) GetCategory(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID kategori tidak valid"})
	}

	category, err := h.service.GetCategory(uint(id))
	if err != nil {
		if customErrors.Is(err, customErrors.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Kategori tidak ditemukan"}) // 404
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil kategori: " + err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": category})
}

// ListCategories handles GET /categories
func (h *CategoryHandler) ListCategories(c *fiber.Ctx) error {
	categories, err := h.service.ListCategories()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil daftar kategory"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Daftar kategori berhasil dibuat",
		"data":    categories,
	})
}

// UpdateCategory handles PUT /categories/:id
func (h *CategoryHandler) UpdateCategory(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID kategori tidak valid"})
	}

	var req services.CategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Input JSON tidak valid"})
	}

	updatedCategory, err := h.service.UpdateCategory(uint(id), req)
	if err != nil {
		// --- Custom Error Handling ---
		if customErrors.Is(err, customErrors.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Kategori yang ingin diupdate tidak ditemukan."}) // 404
		}
		if customErrors.Is(err, customErrors.ErrConflict) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Nama kategori duplikat dengan kategori lain."}) // 409
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Gagal mengupdate kategori: " + err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Kategori berhasil diupdate",
		"data":    updatedCategory,
	})
}

// DeleteCategory handles DELETE /categories/:id
func (h *CategoryHandler) DeleteCategory(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID karegori tidak valid"})
	}

	if err := h.service.DeleteCategory(uint(id)); err != nil {
		if customErrors.Is(err, customErrors.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Kategori yang ingin dihapus tidak ditemukan."}) // 404
		}
		if customErrors.Is(err, customErrors.ErrForeignKeyConstraint) {
			// Kategori tidak bisa dihapus karena masih digunakan oleh produk
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Kategori tidak dapat dihapus karena masih digunakan oleh beberapa produk."}) // 409
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menghapus kategori: " + err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Kategori berhasil dihapus",
	})
}
