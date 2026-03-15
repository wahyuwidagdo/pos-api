package handlers

import (
	"pos-api/internal/models"
	"pos-api/internal/services"

	"github.com/gofiber/fiber/v2"
)

type StoreSettingHandler struct {
	service services.StoreSettingService
}

func NewStoreSettingHandler(s services.StoreSettingService) *StoreSettingHandler {
	return &StoreSettingHandler{service: s}
}

// GetStoreSettings handles GET /store-settings
func (h *StoreSettingHandler) GetStoreSettings(c *fiber.Ctx) error {
	settings, err := h.service.GetSettings(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal mengambil pengaturan toko",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Pengaturan toko berhasil dimuat",
		"data":    settings,
	})
}

// UpdateStoreSettings handles PUT /store-settings
func (h *StoreSettingHandler) UpdateStoreSettings(c *fiber.Ctx) error {
	var req models.StoreSetting
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Input tidak valid",
		})
	}

	settings, err := h.service.UpdateSettings(c.UserContext(), &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal menyimpan pengaturan toko",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Pengaturan toko berhasil disimpan",
		"data":    settings,
	})
}
