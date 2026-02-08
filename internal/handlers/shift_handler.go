package handlers

import (
	"strconv"

	"pos-api/internal/services"

	"github.com/gofiber/fiber/v2"
)

// ShiftHandler handles shift-related HTTP requests
type ShiftHandler struct {
	service services.ShiftService
}

// NewShiftHandler creates a new shift handler
func NewShiftHandler(s services.ShiftService) *ShiftHandler {
	return &ShiftHandler{service: s}
}

// OpenShift handles POST /shifts/open
// @Summary      Open New Shift
// @Description  Open a new cashier shift with starting cash amount. Only one shift can be open at a time. Requires authentication.
// @Tags         Shifts
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request body services.OpenShiftRequest true "Starting cash amount"
// @Success      201 {object} utils.SuccessResponse{data=models.Shift} "Shift opened successfully"
// @Failure      400 {object} utils.ErrorResponse "Invalid input or shift already open"
// @Failure      401 {object} utils.ErrorResponse "Authentication required"
// @Failure      409 {object} utils.ErrorResponse "A shift is already open"
// @Router       /shifts/open [post]
func (h *ShiftHandler) OpenShift(c *fiber.Ctx) error {
	// Get user ID from JWT context
	userID, ok := c.Locals("userID").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Tidak dapat mengambil informasi pengguna",
		})
	}

	var req services.OpenShiftRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Input tidak valid",
		})
	}

	shift, err := h.service.OpenShift(uint(userID), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Shift berhasil dibuka",
		"data":    shift,
	})
}

// CloseShift handles POST /shifts/close
// @Summary      Close Current Shift
// @Description  Close the currently open shift with ending cash count. Calculates expected vs actual cash and difference. Requires authentication.
// @Tags         Shifts
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request body services.CloseShiftRequest true "Ending cash amount and notes"
// @Success      200 {object} utils.SuccessResponse{data=models.Shift} "Shift closed successfully with reconciliation"
// @Failure      400 {object} utils.ErrorResponse "Invalid input or no open shift"
// @Failure      401 {object} utils.ErrorResponse "Authentication required"
// @Router       /shifts/close [post]
func (h *ShiftHandler) CloseShift(c *fiber.Ctx) error {
	// Get user ID from JWT context
	userID, ok := c.Locals("userID").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Tidak dapat mengambil informasi pengguna",
		})
	}

	var req services.CloseShiftRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Input tidak valid",
		})
	}

	shift, err := h.service.CloseShift(uint(userID), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Shift berhasil ditutup",
		"data":    shift,
	})
}

// GetCurrentShift handles GET /shifts/current
// @Summary      Get Current Open Shift
// @Description  Get the currently open shift with real-time transaction summary. Requires authentication.
// @Tags         Shifts
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200 {object} utils.SuccessResponse{data=models.Shift} "Current shift data"
// @Failure      401 {object} utils.ErrorResponse "Authentication required"
// @Failure      404 {object} utils.ErrorResponse "No open shift found"
// @Router       /shifts/current [get]
func (h *ShiftHandler) GetCurrentShift(c *fiber.Ctx) error {
	shift, err := h.service.GetCurrentShift()
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Data shift ditemukan",
		"data":    shift,
	})
}

// GetShift handles GET /shifts/:id
// @Summary      Get Shift by ID
// @Description  Get shift details by ID including reconciliation data. Requires Admin or Manager role.
// @Tags         Shifts
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "Shift ID"
// @Success      200 {object} utils.SuccessResponse{data=models.Shift} "Shift data"
// @Failure      400 {object} utils.ErrorResponse "Invalid shift ID"
// @Failure      401 {object} utils.ErrorResponse "Authentication required"
// @Failure      403 {object} utils.ErrorResponse "Insufficient permissions"
// @Failure      404 {object} utils.ErrorResponse "Shift not found"
// @Router       /shifts/{id} [get]
func (h *ShiftHandler) GetShift(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID shift tidak valid",
		})
	}

	shift, err := h.service.GetShiftByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Shift tidak ditemukan",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Data shift ditemukan",
		"data":    shift,
	})
}

// ListShifts handles GET /shifts
// @Summary      List All Shifts
// @Description  Get a paginated list of all shifts (open and closed). Requires Admin or Manager role.
// @Tags         Shifts
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        page query int false "Page number (default: 1)" default(1)
// @Param        pageSize query int false "Items per page (default: 10)" default(10)
// @Success      200 {object} utils.PagedResponse{data=[]models.Shift} "List of shifts"
// @Failure      401 {object} utils.ErrorResponse "Authentication required"
// @Failure      403 {object} utils.ErrorResponse "Insufficient permissions"
// @Failure      500 {object} utils.ErrorResponse "Internal server error"
// @Router       /shifts [get]
func (h *ShiftHandler) ListShifts(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize", "10"))

	shifts, err := h.service.ListShifts(page, pageSize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal mengambil daftar shift",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "Daftar shift berhasil dimuat",
		"data":     shifts,
		"page":     page,
		"pageSize": pageSize,
	})
}
