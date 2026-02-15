package routes

import (
	"pos-api/internal/handlers"
	"pos-api/internal/middlewares"

	"github.com/gofiber/fiber/v2"
)

// ProtectedRoutes mendaftarkan route yang memerlukan otentikasi JWT.
func ProtectedRoutes(
	router fiber.Router,
	productHandler *handlers.ProductHandler,
	transactionHandler *handlers.TransactionHandler,
	categoryHandler *handlers.CategoryHandler,
	dashboardHandler *handlers.DashboardHandler,
	reportHandler *handlers.ReportHandler,
	authHandler *handlers.AuthHandler,
	storeSettingHandler *handlers.StoreSettingHandler,
) {
	// Middleware JWT digunakan untuk semua route di bawah ini
	jwtMiddleware := middlewares.JWTMiddleware()

	// --- DEFINISI RBAC ROLES ---
	adminManager := middlewares.RBACMiddleware(middlewares.RoleAdmin, middlewares.RoleManager)
	allRoles := middlewares.RBACMiddleware(middlewares.RoleAdmin, middlewares.RoleManager, middlewares.RoleCashier)
	adminOnly := middlewares.RBACMiddleware(middlewares.RoleAdmin)
	// --- END DEFINISI RBAC ROLES ---

	// --- DASHBOARD Routes --- (Admin/Manager)
	dashboardGroup := router.Group("/dashboard", jwtMiddleware, adminManager)
	dashboardGroup.Get("/", dashboardHandler.GetDashboard) // GET /api/v1/dashboard

	// --- REPORTS Routes --- (Admin/Manager)
	reportGroup := router.Group("/reports", jwtMiddleware, adminManager)
	reportGroup.Get("/sales", reportHandler.GetSalesReport)      // GET /api/v1/reports/sales
	reportGroup.Get("/products", reportHandler.GetProductReport) // GET /api/v1/reports/products
	reportGroup.Get("/stock-value", reportHandler.GetStockValue) // GET /api/v1/reports/stock-value

	// --- STORE SETTINGS Routes ---
	storeSettingsGroup := router.Group("/store-settings", jwtMiddleware)
	storeSettingsGroup.Get("/", allRoles, storeSettingHandler.GetStoreSettings)     // GET /api/v1/store-settings
	storeSettingsGroup.Put("/", adminOnly, storeSettingHandler.UpdateStoreSettings) // PUT /api/v1/store-settings (admin only)

	// --- PRODUCT Routes ---
	// READ: All authenticated users can view products (cashiers need this for POS)
	productGroup := router.Group("/products", jwtMiddleware, allRoles)
	productGroup.Get("/", productHandler.ListProducts)
	productGroup.Get("/low-stock", productHandler.GetLowStockProducts) // GET /api/v1/products/low-stock
	productGroup.Get("/:id", productHandler.GetProduct)

	// WRITE: Only Admin/Manager can create, update, delete products
	productGroup.Post("/", adminManager, productHandler.CreateProduct)
	productGroup.Put("/:id", adminManager, productHandler.UpdateProduct)
	productGroup.Delete("/:id", adminManager, productHandler.DeleteProduct)

	// --- CATEGORY Routes ---
	// READ: All authenticated users can view categories (cashiers need this for POS filtering)
	categoryGroup := router.Group("/categories", jwtMiddleware, allRoles)
	categoryGroup.Get("/", categoryHandler.ListCategories)
	categoryGroup.Get("/:id", categoryHandler.GetCategory)

	// WRITE: Only Admin/Manager can create, update, delete categories
	categoryGroup.Post("/", adminManager, categoryHandler.CreateCategory)
	categoryGroup.Put("/:id", adminManager, categoryHandler.UpdateCategory)
	categoryGroup.Delete("/:id", adminManager, categoryHandler.DeleteCategory)

	// --- TRANSACTION Routes --- (Admin/Manager boleh CRUD)
	transactionGroup := router.Group("/transactions", jwtMiddleware) // Hanya JWT, RBAC diterapkan per endpoint

	// Endpoint Penjualan: Bisa diakses oleh KASIR
	transactionGroup.Post("/", allRoles, transactionHandler.CreateTransaction) // POST /api/v1/transactions

	// Endpoint Laporan: Hanya diakses oleh ADMIN/MANAGER
	transactionGroup.Get("/", adminManager, transactionHandler.ListTransactions)  // GET /api/v1/transactions
	transactionGroup.Get("/:id", adminManager, transactionHandler.GetTransaction) // GET /api/v1/transactions/:id

	// --- USER PROFILE Routes --- (All authenticated roles)
	profileGroup := router.Group("/auth", jwtMiddleware, allRoles)
	profileGroup.Get("/profile", authHandler.GetProfile)      // GET /api/v1/auth/profile
	profileGroup.Put("/profile", authHandler.UpdateProfile)   // PUT /api/v1/auth/profile
	profileGroup.Put("/password", authHandler.ChangePassword) // PUT /api/v1/auth/password
}
