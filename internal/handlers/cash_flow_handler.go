package handlers

import (
	"pos-api/internal/services"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type CashFlowHandler struct {
	service services.CashFlowService
}

func NewCashFlowHandler(s services.CashFlowService) *CashFlowHandler {
	return &CashFlowHandler{service: s}
}

// CreateCashFlow handles POST /cash-flow
func (h *CashFlowHandler) CreateCashFlow(c *fiber.Ctx) error {
	var req services.CreateCashFlowRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	userIDFloat, ok := c.Locals("userID").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User ID not found in token"})
	}
	userID := uint(userIDFloat)

	cf, err := h.service.Create(c.UserContext(), req, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Cash flow entry created",
		"data":    cf,
	})
}

// UpdateCashFlow handles PUT /cash-flow/:id
func (h *CashFlowHandler) UpdateCashFlow(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var req services.UpdateCashFlowRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	cf, err := h.service.Update(c.UserContext(), uint(id), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Cash flow entry updated",
		"data":    cf,
	})
}

// DeleteCashFlow handles DELETE /cash-flow/:id
func (h *CashFlowHandler) DeleteCashFlow(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	if err := h.service.Delete(c.UserContext(), uint(id)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Cash flow entry deleted"})
}

// GetCashFlow handles GET /cash-flow/:id
func (h *CashFlowHandler) GetCashFlow(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	cf, err := h.service.GetByID(c.UserContext(), uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Cash flow entry retrieved",
		"data":    cf,
	})
}

// ListCashFlows handles GET /cash-flow
func (h *CashFlowHandler) ListCashFlows(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize", "20"))
	cfType := c.Query("type", "")
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
			endDate = &t
		}
	}

	flows, total, err := h.service.GetAll(c.UserContext(), page, pageSize, cfType, source, startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":     "Cash flow entries retrieved",
		"data":        flows,
		"total_items": total,
		"page":        page,
		"page_size":   pageSize,
	})
}

// GetSummary handles GET /cash-flow/summary
func (h *CashFlowHandler) GetSummary(c *fiber.Ctx) error {
	// Default to current month if no dates provided
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	hasDateFilter := false
	if sd := c.Query("start_date"); sd != "" {
		t, err := time.Parse("2006-01-02", sd)
		if err == nil {
			startDate = t
			hasDateFilter = true
		}
	}
	if ed := c.Query("end_date"); ed != "" {
		t, err := time.Parse("2006-01-02", ed)
		if err == nil {
			endDate = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			hasDateFilter = true
		}
	}

	// If neither start nor end was provided, use default (current month)
	// This preserves backward compatibility
	_ = hasDateFilter

	summary, err := h.service.GetSummary(c.UserContext(), startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Cash flow summary retrieved",
		"data":    summary,
	})
}
