package handlers

import (
	"pos-api/internal/services"

	"github.com/gofiber/fiber/v2"
)

// DashboardHandler handles dashboard-related HTTP requests
type DashboardHandler struct {
	service services.DashboardService
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(s services.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: s}
}

// GetDashboard handles GET /dashboard
// @Summary      Get Dashboard Statistics
// @Description  Get sales summary including total sales, transaction count, items sold, low stock alerts, and top-selling products. Supports date range filtering via 'range' query param (today, this_week, this_month, this_year). Requires Admin or Manager role.
// @Tags         Dashboard
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        range query string false "Date range filter" Enums(today, this_week, this_month, this_year) default(today)
// @Success      200 {object} utils.SuccessResponse{data=services.DashboardResponse} "Dashboard data retrieved successfully"
// @Failure      401 {object} utils.ErrorResponse "Authentication required"
// @Failure      403 {object} utils.ErrorResponse "Insufficient permissions"
// @Failure      500 {object} utils.ErrorResponse "Internal server error"
// @Router       /dashboard [get]
func (h *DashboardHandler) GetDashboard(c *fiber.Ctx) error {
	dateRange := c.Query("range", "today")

	dashboard, err := h.service.GetDashboard(c.UserContext(), dateRange)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal mengambil data dashboard: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Data dashboard berhasil dimuat",
		"data":    dashboard,
	})
}
