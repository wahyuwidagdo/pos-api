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

	category, err := h.service.CreateCategory(req)
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
// @Summary      List All Categories
// @Description  Retrieve a list of all product categories. Requires Admin or Manager role.
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200 {object} utils.SuccessResponse{data=[]models.Category} "List of categories"
// @Failure      401 {object} utils.ErrorResponse "Authentication required"
// @Failure      403 {object} utils.ErrorResponse "Insufficient permissions"
// @Failure      500 {object} utils.ErrorResponse "Internal server error"
// @Router       /categories [get]
func (h *CategoryHandler) ListCategories(c *fiber.Ctx) error {
	categories, err := h.service.ListCategories()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil daftar kategory"})
	}

	return utils.JSONSuccess(c, fiber.StatusOK, "Daftar kategori berhasil dimuat", categories)
}

// UpdateCategory handles PUT /categories/{id}
// @Summary      Update Category
// @Description  Update an existing category by its ID. Requires Admin or Manager role.
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "Category ID"
// @Param        request body services.CategoryRequest true "Updated category data"
// @Success      200 {object} utils.SuccessResponse{data=models.Category} "Category updated successfully"
// @Failure      400 {object} utils.ErrorResponse "Invalid input or validation error"
// @Failure      401 {object} utils.ErrorResponse "Authentication required"
// @Failure      403 {object} utils.ErrorResponse "Insufficient permissions"
// @Failure      404 {object} utils.ErrorResponse "Category not found"
// @Failure      409 {object} utils.ErrorResponse "Category name conflicts with existing category"
// @Router       /categories/{id} [put]
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

// DeleteCategory handles DELETE /categories/{id}
// @Summary      Delete Category
// @Description  Delete a category by its ID. Requires Admin or Manager role. Cannot delete if category is used by products.
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "Category ID"
// @Success      200 {object} utils.SuccessResponse "Category deleted successfully"
// @Failure      400 {object} utils.ErrorResponse "Invalid category ID"
// @Failure      401 {object} utils.ErrorResponse "Authentication required"
// @Failure      403 {object} utils.ErrorResponse "Insufficient permissions"
// @Failure      404 {object} utils.ErrorResponse "Category not found"
// @Failure      409 {object} utils.ErrorResponse "Category is used by products and cannot be deleted"
// @Router       /categories/{id} [delete]
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
