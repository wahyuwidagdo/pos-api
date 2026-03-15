package main

import (
	"log"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	_ "pos-api/docs"
	"pos-api/internal/config"
	"pos-api/internal/handlers"
	"pos-api/internal/listeners"
	appLogger "pos-api/internal/logger"
	"pos-api/internal/pkg/events"
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
	// 1. Muat konfigurasi
	cfg := config.LoadConfig()

	// Initialize Logger
	appLogger.InitLogger(cfg.Environment)

	// Validate JWT secret length
	if len(cfg.JWTSecret) < 32 {
		slog.Warn("JWT_SECRET is shorter than 32 characters. Use a longer secret for production!")
	}

	// 2. Inisiasi koneksi Database (GORM & Migration)
	database.ConnectDB(cfg)

	// Run Database Migrations (golang-migrate)
	database.RunMigrations(database.DB)

	// 3. Inisiasi Fiber App
	app := fiber.New()
	app.Use(logger.New())

	// CORS Middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSOrigins,
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
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

	// --- INITIALIZE EVENT BUS ---
	eventBus := events.NewMemoryEventBus()
	eventBus.Subscribe(events.EventTransactionCreated, listeners.HandleCashFlowOnTransaction)
	eventBus.Subscribe(events.EventTransactionCreated, listeners.HandleInventoryOnTransaction)

	eventBus.Subscribe(events.EventTransactionReturned, listeners.HandleCashFlowOnTransactionReverted)
	eventBus.Subscribe(events.EventTransactionReturned, listeners.HandleInventoryOnTransactionReverted)

	eventBus.Subscribe(events.EventTransactionCancelled, listeners.HandleCashFlowOnTransactionReverted)
	eventBus.Subscribe(events.EventTransactionCancelled, listeners.HandleInventoryOnTransactionReverted)

	eventBus.Subscribe(events.EventInventoryAdjusted, listeners.HandleCashFlowOnInventoryAdjusted)

	// --- TRANSACTION Module ---
	transactionRepo := repositories.NewTransactionRepository(database.DB, eventBus)
	transactionService := services.NewTransactionService(transactionRepo, productRepo)
	transactionHandler := handlers.NewTransactionHandler(transactionService)

	// --- CATEGORY Module ---
	categoryRepo := repositories.NewCategoryRepository(database.DB)
	categoryService := services.NewCategoryService(categoryRepo)
	categoryHandler := handlers.NewCategoryHandler(categoryService)

	// --- CASH FLOW Module ---
	cashFlowRepo := repositories.NewCashFlowRepository(database.DB)
	cashFlowService := services.NewCashFlowService(cashFlowRepo)
	cashFlowHandler := handlers.NewCashFlowHandler(cashFlowService)

	// --- DASHBOARD Module ---
	dashboardRepo := repositories.NewDashboardRepository(database.DB)
	dashboardService := services.NewDashboardService(dashboardRepo, cashFlowRepo)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)

	// --- REPORT Module ---
	reportRepo := repositories.NewReportRepository(database.DB)
	reportService := services.NewReportService(reportRepo)
	reportHandler := handlers.NewReportHandler(reportService)

	// --- STORE SETTINGS Module ---
	storeSettingRepo := repositories.NewStoreSettingRepository(database.DB)
	storeSettingService := services.NewStoreSettingService(storeSettingRepo)
	storeSettingHandler := handlers.NewStoreSettingHandler(storeSettingService)

	// --- EXPORT Module ---
	exportHandler := handlers.NewExportHandler(productService, transactionService)

	// --- BARCODE Module ---
	barcodeHandler := handlers.NewBarcodeHandler(productService)

	// --- INVENTORY LOG Module ---
	inventoryLogRepo := repositories.NewInventoryLogRepository(database.DB, eventBus)
	inventoryLogService := services.NewInventoryLogService(inventoryLogRepo, productRepo)
	inventoryLogHandler := handlers.NewInventoryLogHandler(inventoryLogService)

	// --- PAYMENT METHOD Module ---
	paymentMethodRepo := repositories.NewPaymentMethodRepository(database.DB)
	paymentMethodService := services.NewPaymentMethodService(paymentMethodRepo)
	paymentMethodHandler := handlers.NewPaymentMethodHandler(paymentMethodService)

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
		authHandler,
		storeSettingHandler,
		exportHandler,
		barcodeHandler,
		inventoryLogHandler,
		cashFlowHandler,
		paymentMethodHandler,
	)

	// 6. Jalankan Server
	slog.Info("Starting server on port " + cfg.AppPort)
	log.Fatal(app.Listen(":" + cfg.AppPort))
}
