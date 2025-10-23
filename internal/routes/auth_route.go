package routes

import (
	"pos-api/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

// AuthRoutes menerima router group dan mendaftarkan route Auth
func AuthRoutes(router fiber.Router, authHandler *handlers.AuthHandler) {
	router.Post("/register", authHandler.Register)
	router.Post("/login", authHandler.Login)
}
