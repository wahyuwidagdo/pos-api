package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"

	_ "pos-api/docs"
	"pos-api/internal/handlers"
	"pos-api/internal/models"
	"pos-api/internal/repositories"
	"pos-api/internal/routes"
	"pos-api/internal/services"
	"pos-api/pkg/database"

	swagger "github.com/arsmn/fiber-swagger/v2" // Package untuk melayani UI Swagger
)

// @title POS MVP API Documentation
// @version 1.0
// @description Ini adalah dokumentasi untuk Backend API Point of Sale (POS) MVP.
// @termsOfService http://swagger.io/terms/

// @contact.name Support Tim
// @contact.email support@posmvp.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:3000
// @BasePath /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	// 1. Muat variable lingkungan
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	port := os.Getenv("PORT")

	// 2. Inisiasi koneksi Database (GORM & Migration)
	database.ConnectDB()

	// Auto-migrate Shift model
	database.DB.AutoMigrate(&models.Shift{})

	// 3. Inisiasi Fiber App
	app := fiber.New()
	app.Use(logger.New())

	// Terapkan CORS Middleware <-- HARUS ADA DI SINI
	app.Use(cors.New(cors.Config{
		// Izinkan request hanya dari domain/port frontend Anda (misalnya 5173 atau 3001)
		// Ganti 'http://localhost:5173' dengan port tempat React Anda berjalan
		AllowOrigins: "http://localhost:5173, http://127.0.0.1:5173",

		// Izinkan metode HTTP yang dibutuhkan
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",

		// Izinkan Header Kritis (termasuk Authorization untuk JWT)
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

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

	// --- DASHBOARD Module ---
	dashboardRepo := repositories.NewDashboardRepository(database.DB)
	dashboardService := services.NewDashboardService(dashboardRepo)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)

	// --- REPORT Module ---
	reportRepo := repositories.NewReportRepository(database.DB)
	reportService := services.NewReportService(reportRepo)
	reportHandler := handlers.NewReportHandler(reportService)

	// --- SHIFT Module ---
	shiftRepo := repositories.NewShiftRepository(database.DB)
	shiftService := services.NewShiftService(shiftRepo)
	shiftHandler := handlers.NewShiftHandler(shiftService)

	// 5. Definisi Route
	// Health Check
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": "POS API is running smoothly!",
		})
	})

	apiV1 := app.Group("/api/v1")

	// 5a. SWAGGER ROUTE: Akses di http://localhost:PORT/swagger/*
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Route Publik (Auth)
	routes.AuthRoutes(apiV1.Group("/auth"), authHandler)

	// Route Terproteksi (Perlu Token)
	// Pass router group utama (apiV1) dan semua handlers yang dibutuhkan.
	routes.ProtectedRoutes(
		apiV1,
		productHandler,
		transactionHandler,
		categoryHandler,
		dashboardHandler,
		reportHandler,
		shiftHandler,
	)

	// 6. Jalankan Server
	log.Fatal(app.Listen(":" + port))
}

