package handlers

import (
	"encoding/csv"
	"fmt"
	"pos-api/internal/models"
	"pos-api/internal/services"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ExportHandler handles data export requests
type ExportHandler struct {
	productService     services.ProductService
	transactionService services.TransactionService
}

// NewExportHandler creates a new ExportHandler instance.
func NewExportHandler(ps services.ProductService, ts services.TransactionService) *ExportHandler {
	return &ExportHandler{
		productService:     ps,
		transactionService: ts,
	}
}

// ExportProductsCSV exports all products as CSV
// @Summary      Export Products CSV
// @Description  Export all products to CSV format
// @Tags         Export
// @Produce      text/csv
// @Security     ApiKeyAuth
// @Success      200 {file} file "CSV file"
// @Router       /export/products/csv [get]
func (h *ExportHandler) ExportProductsCSV(c *fiber.Ctx) error {
	products, _, err := h.productService.ListProducts(c.UserContext(), 1, 10000, "", "", "name", "asc", false)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch products",
		})
	}

	filename := fmt.Sprintf("products_%s.csv", time.Now().Format("2006-01-02"))
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	writer := csv.NewWriter(c.Response().BodyWriter())

	// BOM for Excel UTF-8 compatibility
	c.Response().BodyWriter().Write([]byte{0xEF, 0xBB, 0xBF})

	// Header row
	writer.Write([]string{
		"ID", "Name", "SKU", "Description", "Price", "Cost", "Stock", "Category", "Created At",
	})

	// Data rows
	for _, p := range products {
		writer.Write([]string{
			strconv.FormatUint(uint64(p.ID), 10),
			p.Name,
			p.SKU,
			p.Description,
			fmt.Sprintf("%.2f", p.Price),
			fmt.Sprintf("%.2f", p.Cost),
			strconv.Itoa(p.Stock),
			p.Category.Name,
			p.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	writer.Flush()
	return nil
}

// ExportTransactionsCSV exports transactions as CSV
// @Summary      Export Transactions CSV
// @Description  Export all transactions to CSV format
// @Tags         Export
// @Produce      text/csv
// @Security     ApiKeyAuth
// @Success      200 {file} file "CSV file"
// @Router       /export/transactions/csv [get]
func (h *ExportHandler) ExportTransactionsCSV(c *fiber.Ctx) error {
	// Fetch all transactions for export (page 1, large limit)
	paginationData, err := h.transactionService.ListTransactions(c.UserContext(), 1, 10000, "", "", "")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch transactions",
		})
	}

	transactions, ok := paginationData.Data.([]models.Transaction)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid data format",
		})
	}

	filename := fmt.Sprintf("transactions_%s.csv", time.Now().Format("2006-01-02"))
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	writer := csv.NewWriter(c.Response().BodyWriter())

	// BOM for Excel UTF-8 compatibility
	c.Response().BodyWriter().Write([]byte{0xEF, 0xBB, 0xBF})

	writer.Write([]string{
		"ID", "Transaction Code", "Total Amount", "Discount",
		"Grand Total", "Cash", "Change", "Payment Method", "Date",
	})

	for _, t := range transactions {
		writer.Write([]string{
			strconv.FormatUint(uint64(t.ID), 10),
			t.TransactionCode,
			fmt.Sprintf("%.2f", t.TotalAmount),
			fmt.Sprintf("%.2f", t.Discount),
			fmt.Sprintf("%.2f", t.GrandTotal),
			fmt.Sprintf("%.2f", t.Cash),
			fmt.Sprintf("%.2f", t.Change),
			t.PaymentMethod,
			t.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	writer.Flush()
	return nil
}
