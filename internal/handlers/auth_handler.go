package handlers

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"pos-api/internal/pkg/utils"
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

// Register handles POST /auth/register
// @Summary      Register New User
// @Description  Register a new user account for the POS system
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body services.AuthRequest true "User registration credentials"
// @Success      201 {object} utils.SuccessResponse{data=models.User} "User registered successfully"
// @Failure      400 {object} utils.ErrorResponse "Invalid input or validation error"
// @Failure      409 {object} utils.ErrorResponse "Username already exists"
// @Router       /auth/register [post]
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

	return utils.JSONSuccess(c, fiber.StatusCreated, "Pendaftaran berhasil", user)
}

// Login handles POST /auth/login
// @Summary      User Login
// @Description  Authenticate user and return JWT token for API access
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body services.AuthRequest true "User login credentials"
// @Success      200 {object} utils.SuccessResponse{data=object{token=string}} "Login successful, returns JWT token"
// @Failure      400 {object} utils.ErrorResponse "Invalid input format"
// @Failure      401 {object} utils.ErrorResponse "Invalid username or password"
// @Router       /auth/login [post]
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

	return utils.JSONSuccess(c, fiber.StatusOK, "Login berhasil", fiber.Map{"token": token})
}
