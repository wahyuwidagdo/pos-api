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

// Categories handles POST /categories
// @Summary Create Kategori
// @Description Membuat kategori baru.
// @Tags Categories
// @Accept json
// @Produce json
// @Param category body services.CategoryRequest true "Kredensial Category"
// @Success 200 {object} map[string]string "Berhasil Create Category"
// @Failure 400 {object} map[string]string "Validasi/Input Invalid"
// @Failure 401 {object} map[string]string "Kredensial Tidak Valid"
// @Router /categories [post]
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

// Categories handles GET /categories/:id
// @Summary Get Kategori
// @Description Mengambil 1 kategori berdasarkan id.
// @Tags Categories
// @Accept json
// @Produce json
// @Param category body services.CategoryRequest true "Kredensial Category"
// @Success 200 {object} map[string]string "Berhasil Get Category"
// @Failure 400 {object} map[string]string "Validasi/Input Invalid"
// @Failure 401 {object} map[string]string "Kredensial Tidak Valid"
// @Router /categories/:id [get]
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

// Categories handles GET /categories
// @Summary Get All Kategori
// @Description Mengambil semua daftar kategori.
// @Tags Categories
// @Accept json
// @Produce json
// @Param category body services.CategoryRequest true "Kredensial Category"
// @Success 200 {object} map[string]string "Berhasil Get All Categories"
// @Failure 400 {object} map[string]string "Validasi/Input Invalid"
// @Failure 401 {object} map[string]string "Kredensial Tidak Valid"
// @Router /categories [get]
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

// Categories handles PUT /categories/:id
// @Summary Update Kategori
// @Description Update 1 kategori berdasarkan id.
// @Tags Categories
// @Accept json
// @Produce json
// @Param category body services.CategoryRequest true "Kredensial Category"
// @Success 200 {object} map[string]string "Berhasil Create Category"
// @Failure 400 {object} map[string]string "Validasi/Input Invalid"
// @Failure 401 {object} map[string]string "Kredensial Tidak Valid"
// @Router /categories/:id [put]
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

// Categories handles DELETE /categories/:id
// @Summary Delete Kategori
// @Description Delete 1 kategori berdasarkan id.
// @Tags Categories
// @Accept json
// @Produce json
// @Param category body services.CategoryRequest true "Kredensial Category"
// @Success 200 {object} map[string]string "Berhasil Delete Category"
// @Failure 400 {object} map[string]string "Validasi/Input Invalid"
// @Failure 401 {object} map[string]string "Kredensial Tidak Valid"
// @Router /categories/:id [delete]
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
