package handlers

import (
	"pos-api/internal/services"
	"strconv"

	customErrors "pos-api/internal/pkg/errors" // Import custom errors

	"pos-api/internal/pkg/utils"

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

// CreateProduct handles POST /products
// @Summary      Create New Product
// @Description  Create a new product in the inventory. Requires Admin or Manager role.
// @Tags         Products
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request body services.ProductRequest true "Product data"
// @Success      201 {object} utils.SuccessResponse{data=models.Product} "Product created successfully"
// @Failure      400 {object} utils.ErrorResponse "Invalid input or validation error"
// @Failure      401 {object} utils.ErrorResponse "Authentication required"
// @Failure      403 {object} utils.ErrorResponse "Insufficient permissions"
// @Failure      409 {object} utils.ErrorResponse "Product SKU or name already exists"
// @Router       /products [post]
func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	var req services.ProductRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Input JSON tidak valid"})
	}

	product, err := h.service.CreateProduct(req)
	if err != nil {
		if customErrors.Is(err, customErrors.ErrConflict) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "SKU atau nama produk sudah ada (duplikat)."}) // 409
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Gagal membuat produk: " + err.Error()})
	}

	return utils.JSONSuccess(c, fiber.StatusCreated, "Produk berhasil dibuat", product)
}

// GetProduct handles GET /products/{id}
// @Summary      Get Product by ID
// @Description  Retrieve a single product by its ID. Requires Admin or Manager role.
// @Tags         Products
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "Product ID"
// @Success      200 {object} utils.SuccessResponse{data=models.Product} "Product found"
// @Failure      400 {object} utils.ErrorResponse "Invalid product ID"
// @Failure      401 {object} utils.ErrorResponse "Authentication required"
// @Failure      403 {object} utils.ErrorResponse "Insufficient permissions"
// @Failure      404 {object} utils.ErrorResponse "Product not found"
// @Router       /products/{id} [get]
func (h *ProductHandler) GetProduct(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID produk tidak valid"})
	}

	product, err := h.service.GetProduct(uint(id))
	if err != nil {
		if customErrors.Is(err, customErrors.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Produk tidak ditemukan."}) // 404
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil produk"})
	}

	return utils.JSONSuccess(c, fiber.StatusOK, "Detail produk ditemukan", product)
}

// ListProducts handles GET /products
// @Summary      List All Products
// @Description  Retrieve a paginated list of all products. Requires Admin or Manager role.
// @Tags         Products
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        page query int false "Page number (default: 1)" default(1)
// @Param        pageSize query int false "Number of items per page (default: 10)" default(10)
// @Success      200 {object} utils.PagedResponse{data=[]models.Product} "List of products"
// @Failure      401 {object} utils.ErrorResponse "Authentication required"
// @Failure      403 {object} utils.ErrorResponse "Insufficient permissions"
// @Failure      500 {object} utils.ErrorResponse "Internal server error"
// @Router       /products [get]
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

	search := c.Query("search", "")

	products, err := h.service.ListProducts(page, pageSize, search)
	if err != nil {
		// Asumsikan error adalah masalah internal/database
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil daftar produk: " + err.Error()})
	}

	return utils.JSONPaged(c, "Daftar produk berhasil dimuat", products, page, pageSize)
}

// GetLowStockProducts handles GET /products/low-stock
// @Summary      Get Low Stock Products
// @Description  Retrieve products with stock at or below the threshold. Useful for inventory alerts. Requires Admin or Manager role.
// @Tags         Products
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        threshold query int false "Stock threshold (default: 10)" default(10)
// @Success      200 {object} utils.SuccessResponse{data=[]models.Product} "Low stock products"
// @Failure      401 {object} utils.ErrorResponse "Authentication required"
// @Failure      403 {object} utils.ErrorResponse "Insufficient permissions"
// @Failure      500 {object} utils.ErrorResponse "Internal server error"
// @Router       /products/low-stock [get]
func (h *ProductHandler) GetLowStockProducts(c *fiber.Ctx) error {
	threshold, err := strconv.Atoi(c.Query("threshold", "10"))
	if err != nil || threshold <= 0 {
		threshold = 10
	}

	products, err := h.service.GetLowStockProducts(threshold)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal mengambil produk stok rendah: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":   "Daftar produk stok rendah berhasil dimuat",
		"data":      products,
		"threshold": threshold,
		"count":     len(products),
	})
}

// UpdateProduct handles PUT /products/{id}
// @Summary      Update Product
// @Description  Update an existing product by its ID. Requires Admin or Manager role.
// @Tags         Products
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "Product ID"
// @Param        request body services.ProductRequest true "Updated product data"
// @Success      200 {object} utils.SuccessResponse{data=models.Product} "Product updated successfully"
// @Failure      400 {object} utils.ErrorResponse "Invalid input or validation error"
// @Failure      401 {object} utils.ErrorResponse "Authentication required"
// @Failure      403 {object} utils.ErrorResponse "Insufficient permissions"
// @Failure      404 {object} utils.ErrorResponse "Product not found"
// @Failure      409 {object} utils.ErrorResponse "Product SKU or name conflicts with existing product"
// @Router       /products/{id} [put]
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
		// --- Custom Error Handling ---
		if customErrors.Is(err, customErrors.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Produk tidak ditemukan"}) // 404
		}
		if customErrors.Is(err, customErrors.ErrConflict) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "SKU atau nama produk sudah ada."}) // 409
		}
		// --- End Custom Error Handling ---

		// Default Error
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Gagal mengupdate produk: " + err.Error()})
	}

	return utils.JSONSuccess(c, fiber.StatusOK, "Produk berhasil diupdate", updatedProduct)
}

// DeleteProduct handles DELETE /products/{id}
// @Summary      Delete Product
// @Description  Delete a product by its ID. Requires Admin or Manager role. Cannot delete if product is used in transactions.
// @Tags         Products
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "Product ID"
// @Success      200 {object} utils.SuccessResponse "Product deleted successfully"
// @Failure      400 {object} utils.ErrorResponse "Invalid product ID"
// @Failure      401 {object} utils.ErrorResponse "Authentication required"
// @Failure      403 {object} utils.ErrorResponse "Insufficient permissions"
// @Failure      404 {object} utils.ErrorResponse "Product not found"
// @Failure      409 {object} utils.ErrorResponse "Product is used in transactions and cannot be deleted"
// @Router       /products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c *fiber.Ctx) error {
	// 1. Parse ID dari parameter
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID produk tidak valid"})
	}

	// 2. Panggil service untuk delete
	if err := h.service.DeleteProduct(uint(id)); err != nil {
		if customErrors.Is(err, customErrors.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Produk yang ingin dihapus tidak ditemukan."}) // 404
		}
		if customErrors.Is(err, customErrors.ErrForeignKeyConstraint) {
			// Produk tidak bisa dihapus karena masih terkait dengan entitas lain (misalnya transaksi)
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Produk tidak dapat dihapus karena sudah digunakan dalam transaksi."}) // 409
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menghapus produk: " + err.Error()})
	}

	// 3. Beri respons sukses
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Produk berhasil dihapus",
	})
}
