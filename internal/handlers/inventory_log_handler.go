package handlers

import (
	"pos-api/internal/services"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type InventoryLogHandler struct {
	service services.InventoryLogService
}

func NewInventoryLogHandler(s services.InventoryLogService) *InventoryLogHandler {
	return &InventoryLogHandler{service: s}
}

// AdjustStock handles POST /inventory
func (h *InventoryLogHandler) AdjustStock(c *fiber.Ctx) error {
	var req services.StockAdjustmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Get user ID from JWT token
	// Get user ID from Locals (set by JWTMiddleware)
	userIDFloat, ok := c.Locals("userID").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User ID not found in token"})
	}
	userID := uint(userIDFloat)

	log, err := h.service.AdjustStock(c.UserContext(), req, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Stock adjusted successfully",
		"data":    log,
	})
}

// GetLogsByProduct handles GET /inventory/product/:id
func (h *InventoryLogHandler) GetLogsByProduct(c *fiber.Ctx) error {
	productID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid product ID"})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize", "20"))

	logs, total, err := h.service.GetLogsByProduct(c.UserContext(), uint(productID), page, pageSize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":     "Inventory logs retrieved",
		"data":        logs,
		"total_items": total,
		"page":        page,
		"page_size":   pageSize,
	})
}

// GetAllLogs handles GET /inventory
func (h *InventoryLogHandler) GetAllLogs(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize", "20"))
	logType := c.Query("type", "")
	source := c.Query("source", "")

	var startDate, endDate *time.Time
	if sd := c.Query("start_date"); sd != "" {
		t, err := time.Parse("2006-01-02", sd)
		if err == nil {
			startDate = &t
		}
	}
	if ed := c.Query("end_date"); ed != "" {
		t, err := time.Parse("2006-01-02", ed)
		if err == nil {
			// Include the full day until 23:59:59
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endDate = &t
		}
	}

	logs, total, err := h.service.GetAllLogs(c.UserContext(), page, pageSize, logType, source, startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":     "Inventory logs retrieved",
		"data":        logs,
		"total_items": total,
		"page":        page,
		"page_size":   pageSize,
	})
}

func (h *InventoryLogHandler) GetInventoryStats(c *fiber.Ctx) error {
	var startDate, endDate *time.Time
	if sd := c.Query("start_date"); sd != "" {
		t, err := time.Parse("2006-01-02", sd)
		if err == nil {
			startDate = &t
		}
	}
	if ed := c.Query("end_date"); ed != "" {
		t, err := time.Parse("2006-01-02", ed)
		if err == nil {
			// Include the full day until 23:59:59
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endDate = &t
		}
	}

	stats, err := h.service.GetInventoryStats(c.UserContext(), startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Inventory stats retrieved",
		"data":    stats,
	})
}
