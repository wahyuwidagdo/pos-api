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

	if err := validate.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		for _, fieldErr := range validationErrors {
			if fieldErr.Tag() == "required" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fieldErr.Field() + " harus diisi"})
			}
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Validasi gagal: " + err.Error()})
	}

	user, err := h.service.Register(c.UserContext(), req)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	}

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

	token, user, err := h.service.Login(c.UserContext(), req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	// Return both token and user info
	return utils.JSONSuccess(c, fiber.StatusOK, "Login berhasil", fiber.Map{
		"token": token,
		"user": fiber.Map{
			"id":        user.ID,
			"username":  user.Username,
			"full_name": user.FullName,
			"role":      user.Role,
			"is_active": user.IsActive,
		},
	})
}

// GetProfile handles GET /auth/profile
func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	userIDFloat, ok := c.Locals("userID").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User ID not found"})
	}
	userID := uint(userIDFloat)

	user, err := h.service.GetProfile(c.UserContext(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	user.Password = ""
	return utils.JSONSuccess(c, fiber.StatusOK, "Profile retrieved", user)
}

// UpdateProfile handles PUT /auth/profile
func (h *AuthHandler) UpdateProfile(c *fiber.Ctx) error {
	userIDFloat, ok := c.Locals("userID").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User ID not found"})
	}
	userID := uint(userIDFloat)

	var req services.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Input tidak valid"})
	}

	user, err := h.service.UpdateProfile(c.UserContext(), userID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	user.Password = ""
	return utils.JSONSuccess(c, fiber.StatusOK, "Profile updated", user)
}

// ChangePassword handles PUT /auth/password
func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	userIDFloat, ok := c.Locals("userID").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User ID not found"})
	}
	userID := uint(userIDFloat)

	var req services.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Input tidak valid"})
	}

	if err := validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Password fields are required"})
	}

	err := h.service.ChangePassword(c.UserContext(), userID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return utils.JSONSuccess(c, fiber.StatusOK, "Password updated successfully", nil)
}
