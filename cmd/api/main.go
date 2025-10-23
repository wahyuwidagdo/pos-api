package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"

	"pos-api/internal/handlers"
	"pos-api/internal/repositories"
	"pos-api/internal/routes"
	"pos-api/internal/services"
	"pos-api/pkg/database"
)

func main() {
	// 1. Muat variable lingkungan
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	port := os.Getenv("PORT")

	// 2. Inisiasi koneksi Database (GORM & Migration)
	database.ConnectDB()

	// 3. Inisiasi Fiber App
	app := fiber.New()
	app.Use(logger.New())

	// 4. Inisiasi Dependency Injection (DI)
	// --- AUTH Module ---
	authRepo := repositories.NewAuthRepository(database.DB)
	authService := services.NewAuthService(authRepo)
	authHandler := handlers.NewAuthHandler(authService)

	// --- PRODUCT Module ---
	productRepo := repositories.NewProductRepository(database.DB)
	productService := services.NewProductService(productRepo)
	productHandler := handlers.NewProductHandler(productService)

	// --- TRANSACTION Module ---
	transactionRepo := repositories.NewTransactionRepository(database.DB)
	transactionService := services.NewTransactionService(transactionRepo, productRepo)
	transactionHandler := handlers.NewTransactionHandler(transactionService)

	// --- CATEGORY Module ---
	categoryRepo := repositories.NewCategoryRepository(database.DB)
	categoryService := services.NewCategoryService(categoryRepo)
	categoryHandler := handlers.NewCategoryHandler(categoryService)

	// 5. Definisi Route
	// Health Check
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": "POS API is running smoothly!",
		})
	})

	apiV1 := app.Group("/api/v1")
	// Route Publik (Auth)
	routes.AuthRoutes(apiV1.Group("/auth"), authHandler)

	// Route Terproteksi (Perlu Token)
	// Pass router group utama (apiV1) dan semua handlers yang dibutuhkan.
	routes.ProtectedRoutes(apiV1, productHandler, transactionHandler, categoryHandler) // Group /api/v1/products

	// 6. Jalankan Server
	log.Fatal(app.Listen(":" + port))
}
