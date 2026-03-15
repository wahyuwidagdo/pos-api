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
	exportHandler *handlers.ExportHandler,
	barcodeHandler *handlers.BarcodeHandler,
	inventoryLogHandler *handlers.InventoryLogHandler,
	cashFlowHandler *handlers.CashFlowHandler,
	paymentMethodHandler *handlers.PaymentMethodHandler,
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
	productGroup.Get("/stock-counts", productHandler.GetStockCounts)   // GET /api/v1/products/stock-counts
	productGroup.Get("/:id", productHandler.GetProduct)

	// WRITE: Only Admin/Manager can create, update, delete products
	productGroup.Post("/", adminManager, productHandler.CreateProduct)
	productGroup.Put("/:id", adminManager, productHandler.UpdateProduct)
	productGroup.Delete("/:id", adminManager, productHandler.DeleteProduct)
	productGroup.Post("/:id/restore", adminManager, productHandler.RestoreProduct)
	productGroup.Delete("/:id/force", adminManager, productHandler.ForceDeleteProduct)

	// --- CATEGORY Routes ---
	// READ: All authenticated users can view categories (cashiers need this for POS filtering)
	categoryGroup := router.Group("/categories", jwtMiddleware, allRoles)
	categoryGroup.Get("/", categoryHandler.ListCategories)
	categoryGroup.Get("/:id", categoryHandler.GetCategory)

	// WRITE: Only Admin/Manager can create, update, delete categories
	categoryGroup.Post("/", adminManager, categoryHandler.CreateCategory)
	categoryGroup.Put("/:id", adminManager, categoryHandler.UpdateCategory)
	categoryGroup.Delete("/:id", adminManager, categoryHandler.DeleteCategory)
	categoryGroup.Post("/:id/restore", adminManager, categoryHandler.RestoreCategory)     // POST /api/v1/categories/:id/restore
	categoryGroup.Delete("/:id/force", adminManager, categoryHandler.ForceDeleteCategory) // DELETE /api/v1/categories/:id/force

	// --- TRANSACTION Routes --- (Admin/Manager boleh CRUD)
	transactionGroup := router.Group("/transactions", jwtMiddleware) // Hanya JWT, RBAC diterapkan per endpoint

	// Endpoint Penjualan: Bisa diakses oleh KASIR
	transactionGroup.Post("/", allRoles, transactionHandler.CreateTransaction) // POST /api/v1/transactions

	// Endpoint Laporan: Hanya diakses oleh ADMIN/MANAGER
	transactionGroup.Get("/", adminManager, transactionHandler.ListTransactions)             // GET /api/v1/transactions
	transactionGroup.Get("/:id", adminManager, transactionHandler.GetTransaction)            // GET /api/v1/transactions/:id
	transactionGroup.Post("/:id/cancel", adminManager, transactionHandler.CancelTransaction) // POST /api/v1/transactions/:id/cancel
	transactionGroup.Post("/:id/return", adminManager, transactionHandler.ReturnTransaction) // POST /api/v1/transactions/:id/return

	// --- USER PROFILE Routes --- (All authenticated roles)
	profileGroup := router.Group("/auth", jwtMiddleware, allRoles)
	profileGroup.Get("/profile", authHandler.GetProfile)      // GET /api/v1/auth/profile
	profileGroup.Put("/profile", authHandler.UpdateProfile)   // PUT /api/v1/auth/profile
	profileGroup.Put("/password", authHandler.ChangePassword) // PUT /api/v1/auth/password

	// --- EXPORT Routes --- (Admin/Manager)
	exportGroup := router.Group("/export", jwtMiddleware, adminManager)
	exportGroup.Get("/products/csv", exportHandler.ExportProductsCSV)         // GET /api/v1/export/products/csv
	exportGroup.Get("/transactions/csv", exportHandler.ExportTransactionsCSV) // GET /api/v1/export/transactions/csv

	// --- BARCODE Routes --- (All authenticated roles)
	barcodeGroup := router.Group("/barcode", jwtMiddleware, allRoles)
	barcodeGroup.Get("/:id", barcodeHandler.GenerateBarcode)          // GET /api/v1/barcode/:id
	barcodeGroup.Post("/batch", barcodeHandler.GenerateBatchBarcodes) // POST /api/v1/barcode/batch

	// --- INVENTORY LOG Routes --- (Admin/Manager)
	inventoryGroup := router.Group("/inventory", jwtMiddleware, adminManager)
	inventoryGroup.Get("/stats", inventoryLogHandler.GetInventoryStats)      // GET /api/v1/inventory/stats
	inventoryGroup.Get("/", inventoryLogHandler.GetAllLogs)                  // GET /api/v1/inventory
	inventoryGroup.Post("/", inventoryLogHandler.AdjustStock)                // POST /api/v1/inventory
	inventoryGroup.Get("/product/:id", inventoryLogHandler.GetLogsByProduct) // GET /api/v1/inventory/product/:id

	// --- CASH FLOW Routes --- (Admin/Manager)
	cashFlowGroup := router.Group("/cash-flow", jwtMiddleware, adminManager)
	cashFlowGroup.Get("/", cashFlowHandler.ListCashFlows)        // GET /api/v1/cash-flow
	cashFlowGroup.Post("/", cashFlowHandler.CreateCashFlow)      // POST /api/v1/cash-flow
	cashFlowGroup.Get("/summary", cashFlowHandler.GetSummary)    // GET /api/v1/cash-flow/summary
	cashFlowGroup.Get("/:id", cashFlowHandler.GetCashFlow)       // GET /api/v1/cash-flow/:id
	cashFlowGroup.Put("/:id", cashFlowHandler.UpdateCashFlow)    // PUT /api/v1/cash-flow/:id
	cashFlowGroup.Delete("/:id", cashFlowHandler.DeleteCashFlow) // DELETE /api/v1/cash-flow/:id

	// --- PAYMENT METHOD Routes ---
	paymentMethodGroup := router.Group("/payment-methods", jwtMiddleware)
	paymentMethodGroup.Get("/", allRoles, paymentMethodHandler.ListPaymentMethods)            // GET /api/v1/payment-methods (all roles)
	paymentMethodGroup.Get("/active", allRoles, paymentMethodHandler.GetActivePaymentMethods) // GET /api/v1/payment-methods/active (for POS)
	paymentMethodGroup.Post("/", adminManager, paymentMethodHandler.CreatePaymentMethod)      // POST /api/v1/payment-methods
	paymentMethodGroup.Put("/:id", adminManager, paymentMethodHandler.UpdatePaymentMethod)    // PUT /api/v1/payment-methods/:id
	paymentMethodGroup.Delete("/:id", adminOnly, paymentMethodHandler.DeletePaymentMethod)    // DELETE /api/v1/payment-methods/:id
}
