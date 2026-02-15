package handlers

import (
	"strconv"

	"pos-api/internal/services"

	"github.com/gofiber/fiber/v2"
)

// ReportHandler handles report-related HTTP requests
type ReportHandler struct {
	service services.ReportService
}

// NewReportHandler creates a new report handler
func NewReportHandler(s services.ReportService) *ReportHandler {
	return &ReportHandler{service: s}
}

// GetSalesReport handles GET /reports/sales
// @Summary      Get Sales Report
// @Description  Get sales report for a date range including daily breakdown and summary. Requires Admin or Manager role.
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        start_date query string true "Start date (YYYY-MM-DD)" example(2026-02-01)
// @Param        end_date query string true "End date (YYYY-MM-DD)" example(2026-02-08)
// @Success      200 {object} utils.SuccessResponse{data=services.SalesReportResponse} "Sales report retrieved successfully"
// @Failure      400 {object} utils.ErrorResponse "Invalid date format or range"
// @Failure      401 {object} utils.ErrorResponse "Authentication required"
// @Failure      403 {object} utils.ErrorResponse "Insufficient permissions"
// @Failure      500 {object} utils.ErrorResponse "Internal server error"
// @Router       /reports/sales [get]
func (h *ReportHandler) GetSalesReport(c *fiber.Ctx) error {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Parameter start_date dan end_date harus diisi",
		})
	}

	report, err := h.service.GetSalesReport(startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Laporan penjualan berhasil dimuat",
		"data":    report,
	})
}

// GetProductReport handles GET /reports/products
// @Summary      Get Product Performance Report
// @Description  Get product sales performance for a date range, sorted by quantity sold. Requires Admin or Manager role.
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        start_date query string true "Start date (YYYY-MM-DD)" example(2026-02-01)
// @Param        end_date query string true "End date (YYYY-MM-DD)" example(2026-02-08)
// @Param        limit query int false "Limit results (default: 20)" default(20)
// @Success      200 {object} utils.SuccessResponse{data=services.ProductReportResponse} "Product report retrieved successfully"
// @Failure      400 {object} utils.ErrorResponse "Invalid date format or range"
// @Failure      401 {object} utils.ErrorResponse "Authentication required"
// @Failure      403 {object} utils.ErrorResponse "Insufficient permissions"
// @Failure      500 {object} utils.ErrorResponse "Internal server error"
// @Router       /reports/products [get]
func (h *ReportHandler) GetProductReport(c *fiber.Ctx) error {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	limitStr := c.Query("limit", "20")

	if startDate == "" || endDate == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Parameter start_date dan end_date harus diisi",
		})
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	report, err := h.service.GetProductReport(startDate, endDate, limit)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Laporan produk berhasil dimuat",
		"data":    report,
	})
}

// GetStockValue handles GET /reports/stock-value
func (h *ReportHandler) GetStockValue(c *fiber.Ctx) error {
	stockValue, err := h.service.GetStockValue()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal mengambil nilai stok",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Nilai stok berhasil dimuat",
		"data":    stockValue,
	})
}
