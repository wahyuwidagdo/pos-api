package handlers

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"pos-api/internal/services"
)

// Inisiasi validator
var validate = validator.New()

// AuthHandler menangani request HTTP terkait otentikasi.
type AuthHandler struct {
	service services.AuthService
}

// NewAuthHandler membuat instance AuthHandler baru.
func NewAuthHandler(s services.AuthService) *AuthHandler {
	return &AuthHandler{service: s}
}

// Register menangani permintaan pendaftaran pengguna (POST /auth/register).
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req services.AuthRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Input JSON tidak valid"})
	}

	// --- LOGIKA VALIDASI BARU ---
	if err := validate.Struct(req); err != nil {
		// Logika untuk menampilkan error validasi spesifik
		validationErrors := err.(validator.ValidationErrors)

		// Cek apakah ada error required
		for _, fieldErr := range validationErrors {
			if fieldErr.Tag() == "required" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fieldErr.Field() + " harus diisi"})
			}
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Validasi gagal: " + err.Error()})
	}
	// --- AKHIR LOGIKA VALIDASI BARU ---

	// Panggil Service Layer
	user, err := h.service.Register(req)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	}

	// Hindari mengembalikan password yang di-hash
	user.Password = ""

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Pendaftaran berhasil",
		"user":    user,
	})
}

// Login menangani permintaan login pengguna (POST /auth/login).
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req services.AuthRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Input tidak valid"})
	}

	// Panggil Service Layer untuk login dan mendapatkan token
	token, err := h.service.Login(req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login berhasil",
		"token":   token,
	})
}
