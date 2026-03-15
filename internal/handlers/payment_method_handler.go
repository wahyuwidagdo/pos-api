package handlers

import (
	"pos-api/internal/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type PaymentMethodHandler struct {
	service services.PaymentMethodService
}

func NewPaymentMethodHandler(s services.PaymentMethodService) *PaymentMethodHandler {
	return &PaymentMethodHandler{service: s}
}

// CreatePaymentMethod handles POST /payment-methods
func (h *PaymentMethodHandler) CreatePaymentMethod(c *fiber.Ctx) error {
	var req services.CreatePaymentMethodRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	pm, err := h.service.Create(c.UserContext(), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Payment method created",
		"data":    pm,
	})
}

// UpdatePaymentMethod handles PUT /payment-methods/:id
func (h *PaymentMethodHandler) UpdatePaymentMethod(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var req services.UpdatePaymentMethodRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	pm, err := h.service.Update(c.UserContext(), uint(id), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Payment method updated",
		"data":    pm,
	})
}

// DeletePaymentMethod handles DELETE /payment-methods/:id
func (h *PaymentMethodHandler) DeletePaymentMethod(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

	if err := h.service.Delete(c.UserContext(), uint(id)); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Payment method deleted"})
}

// ListPaymentMethods handles GET /payment-methods
func (h *PaymentMethodHandler) ListPaymentMethods(c *fiber.Ctx) error {
	methods, err := h.service.GetAll(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Payment methods retrieved",
		"data":    methods,
	})
}

// GetActivePaymentMethods handles GET /payment-methods/active
func (h *PaymentMethodHandler) GetActivePaymentMethods(c *fiber.Ctx) error {
	methods, err := h.service.GetActive(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Active payment methods retrieved",
		"data":    methods,
	})
}
