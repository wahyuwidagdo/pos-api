package handlers

import (
	"bytes"
	"fmt"
	"image/png"
	"pos-api/internal/services"
	"strconv"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"github.com/boombuler/barcode/ean"
	"github.com/gofiber/fiber/v2"
)

// BarcodeHandler handles barcode generation requests
type BarcodeHandler struct {
	productService services.ProductService
}

// NewBarcodeHandler creates a new BarcodeHandler instance.
func NewBarcodeHandler(ps services.ProductService) *BarcodeHandler {
	return &BarcodeHandler{productService: ps}
}

// GenerateBarcode generates a barcode PNG for a product's SKU
// @Summary      Generate Barcode
// @Description  Generate a barcode image (PNG) for a product's SKU
// @Tags         Barcode
// @Produce      image/png
// @Security     ApiKeyAuth
// @Param        id path int true "Product ID"
// @Param        type query string false "Barcode type: code128 or ean13" default(code128)
// @Param        width query int false "Width in pixels" default(300)
// @Param        height query int false "Height in pixels" default(100)
// @Success      200 {file} file "PNG barcode image"
// @Failure      404 {object} map[string]string "Product not found"
// @Router       /barcode/{id} [get]
func (h *BarcodeHandler) GenerateBarcode(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid product ID"})
	}

	product, err := h.productService.GetProduct(c.UserContext(), uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
	}

	sku := product.SKU
	if sku == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Product has no SKU"})
	}

	barcodeType := c.Query("type", "code128")
	width := c.QueryInt("width", 300)
	height := c.QueryInt("height", 100)

	var bc barcode.Barcode

	switch barcodeType {
	case "ean13":
		// EAN-13 requires exactly 12 or 13 digits
		bc, err = ean.Encode(sku)
	default:
		// Code128 supports any ASCII string
		bc, err = code128.Encode(sku)
	}

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to encode barcode: %s", err.Error()),
		})
	}

	// Scale barcode to requested dimensions
	bc, err = barcode.Scale(bc, width, height)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to scale barcode",
		})
	}

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, bc); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to encode barcode image",
		})
	}

	c.Set("Content-Type", "image/png")
	c.Set("Content-Disposition", fmt.Sprintf("inline; filename=barcode_%s.png", sku))
	return c.Send(buf.Bytes())
}

// GenerateBatchBarcodes generates barcodes for multiple products
// @Summary      Generate Batch Barcodes
// @Description  Generate barcodes for multiple products as a JSON array of base64 images
// @Tags         Barcode
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        body body []int true "Array of product IDs"
// @Success      200 {object} map[string]interface{} "Batch barcode data"
// @Router       /barcode/batch [post]
func (h *BarcodeHandler) GenerateBatchBarcodes(c *fiber.Ctx) error {
	var productIDs []uint
	if err := c.BodyParser(&productIDs); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if len(productIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No product IDs provided"})
	}

	type BarcodeItem struct {
		ProductID   uint    `json:"product_id"`
		ProductName string  `json:"product_name"`
		SKU         string  `json:"sku"`
		Price       float64 `json:"price"`
		BarcodeURL  string  `json:"barcode_url"`
	}

	var results []BarcodeItem
	for _, pid := range productIDs {
		product, err := h.productService.GetProduct(c.UserContext(), pid)
		if err != nil {
			continue // Skip products that don't exist
		}
		if product.SKU == "" {
			continue
		}

		results = append(results, BarcodeItem{
			ProductID:   product.ID,
			ProductName: product.Name,
			SKU:         product.SKU,
			Price:       product.Price,
			BarcodeURL:  fmt.Sprintf("/api/v1/barcode/%d?type=code128&width=300&height=80", product.ID),
		})
	}

	return c.JSON(fiber.Map{
		"data":  results,
		"count": len(results),
	})
}
