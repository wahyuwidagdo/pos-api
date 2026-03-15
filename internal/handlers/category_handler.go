package handlers

import (
	"pos-api/internal/pkg/utils"
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
// @Summary      Create New Category
// @Description  Create a new product category. Requires Admin or Manager role.
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request body services.CategoryRequest true "Category data"
// @Success      201 {object} utils.SuccessResponse{data=models.Category} "Category created successfully"
// @Failure      400 {object} utils.ErrorResponse "Invalid input or validation error"
// @Failure      401 {object} utils.ErrorResponse "Authentication required"
// @Failure      403 {object} utils.ErrorResponse "Insufficient permissions"
// @Failure      409 {object} utils.ErrorResponse "Category name already exists"
// @Router       /categories [post]
func (h *CategoryHandler) CreateCategory(c *fiber.Ctx) error {
	var req services.CategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Input JSON tidak valid"})
	}

	category, err := h.service.CreateCategory(c.UserContext(), req)
	if err != nil {
		if customErrors.Is(err, customErrors.ErrConflict) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Nama kategori sudah ada (duplikat)."}) // 409
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Gagal membuat kategori: " + err.Error()})
	}

	return utils.JSONSuccess(c, fiber.StatusCreated, "Kategori berhasil dibuat", category)
}

// GetCategory handles GET /categories/{id}
// @Summary      Get Category by ID
// @Description  Retrieve a single category by its ID. Requires Admin or Manager role.
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "Category ID"
// @Success      200 {object} utils.SuccessResponse{data=models.Category} "Category found"
// @Failure      400 {object} utils.ErrorResponse "Invalid category ID"
// @Failure      401 {object} utils.ErrorResponse "Authentication required"
// @Failure      403 {object} utils.ErrorResponse "Insufficient permissions"
// @Failure      404 {object} utils.ErrorResponse "Category not found"
// @Router       /categories/{id} [get]
func (h *CategoryHandler) GetCategory(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID kategori tidak valid"})
	}

	category, err := h.service.GetCategory(c.UserContext(), uint(id))
	if err != nil {
		if customErrors.Is(err, customErrors.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Kategori tidak ditemukan"}) // 404
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil kategori: " + err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": category})
}

// ListCategories handles GET /categories
// @Summary      List All Categories
// @Description  Retrieve a list of all product categories. Requires Admin or Manager role.
func (h *CategoryHandler) ListCategories(c *fiber.Ctx) error {
	onlyTrashed := c.Query("trashed") == "true"
	categories, err := h.service.ListCategories(c.UserContext(), onlyTrashed)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil daftar kategory"})
	}

	return utils.JSONSuccess(c, fiber.StatusOK, "Daftar kategori berhasil dimuat", categories)
}

// UpdateCategory handles PUT /categories/{id}
func (h *CategoryHandler) UpdateCategory(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID kategori tidak valid"})
	}

	var req services.CategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Input JSON tidak valid"})
	}

	updatedCategory, err := h.service.UpdateCategory(c.UserContext(), uint(id), req)
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

// DeleteCategory handles DELETE /categories/{id}
func (h *CategoryHandler) DeleteCategory(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID karegori tidak valid"})
	}

	if err := h.service.DeleteCategory(c.UserContext(), uint(id)); err != nil {
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

// RestoreCategory handles POST /categories/{id}/restore
func (h *CategoryHandler) RestoreCategory(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID kategori tidak valid"})
	}

	if err := h.service.RestoreCategory(c.UserContext(), uint(id)); err != nil {
		if customErrors.Is(err, customErrors.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Kategori tidak ditemukan"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengembalikan kategori: " + err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Kategori berhasil dikembalikan",
	})
}

// ForceDeleteCategory handles DELETE /categories/{id}/force
func (h *CategoryHandler) ForceDeleteCategory(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID kategori tidak valid"})
	}

	if err := h.service.ForceDeleteCategory(c.UserContext(), uint(id)); err != nil {
		if customErrors.Is(err, customErrors.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Kategori tidak ditemukan"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menghapus kategori secara permanen: " + err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Kategori berhasil dihapus secara permanen",
	})
}
