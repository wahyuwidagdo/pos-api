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
	shiftHandler *handlers.ShiftHandler,
) {
	// Middleware JWT digunakan untuk semua route di bawah ini
	jwtMiddleware := middlewares.JWTMiddleware()

	// --- DEFINISI RBAC ROLES ---
	adminManager := middlewares.RBACMiddleware(middlewares.RoleAdmin, middlewares.RoleManager)
	allRoles := middlewares.RBACMiddleware(middlewares.RoleAdmin, middlewares.RoleManager, middlewares.RoleCashier)
	// --- END DEFINISI RBAC ROLES ---

	// --- DASHBOARD Routes --- (Admin/Manager)
	dashboardGroup := router.Group("/dashboard", jwtMiddleware, adminManager)
	dashboardGroup.Get("/", dashboardHandler.GetDashboard) // GET /api/v1/dashboard

	// --- REPORTS Routes --- (Admin/Manager)
	reportGroup := router.Group("/reports", jwtMiddleware, adminManager)
	reportGroup.Get("/sales", reportHandler.GetSalesReport)       // GET /api/v1/reports/sales
	reportGroup.Get("/products", reportHandler.GetProductReport) // GET /api/v1/reports/products

	// --- SHIFT Routes --- (All authenticated users can open/close, Admin/Manager can view history)
	shiftGroup := router.Group("/shifts", jwtMiddleware)
	shiftGroup.Post("/open", allRoles, shiftHandler.OpenShift)      // POST /api/v1/shifts/open
	shiftGroup.Post("/close", allRoles, shiftHandler.CloseShift)    // POST /api/v1/shifts/close
	shiftGroup.Get("/current", allRoles, shiftHandler.GetCurrentShift) // GET /api/v1/shifts/current
	shiftGroup.Get("/", adminManager, shiftHandler.ListShifts)      // GET /api/v1/shifts
	shiftGroup.Get("/:id", adminManager, shiftHandler.GetShift)     // GET /api/v1/shifts/:id

	// --- PRODUCT Routes --- (Admin/Manager boleh CRUD)
	productGroup := router.Group("/products", jwtMiddleware, adminManager) // <-- JWT + RBAC diterapkan di sini
	productGroup.Post("/", productHandler.CreateProduct)                   // POST /api/v1/products
	productGroup.Get("/", productHandler.ListProducts)
	productGroup.Get("/low-stock", productHandler.GetLowStockProducts) // GET /api/v1/products/low-stock
	productGroup.Get("/:id", productHandler.GetProduct)
	productGroup.Put("/:id", productHandler.UpdateProduct)
	productGroup.Delete("/:id", productHandler.DeleteProduct)

	// --- CATEGORY Routes ---
	categoryGroup := router.Group("/categories", jwtMiddleware, adminManager) // <-- JWT + RBAC diterapkan di sini
	categoryGroup.Post("/", categoryHandler.CreateCategory)
	categoryGroup.Get("/", categoryHandler.ListCategories)
	categoryGroup.Get("/:id", categoryHandler.GetCategory)
	categoryGroup.Put("/:id", categoryHandler.UpdateCategory)
	categoryGroup.Delete("/:id", categoryHandler.DeleteCategory)

	// --- TRANSACTION Routes --- (Admin/Manager boleh CRUD)
	transactionGroup := router.Group("/transactions", jwtMiddleware) // Hanya JWT, RBAC diterapkan per endpoint

	// Endpoint Penjualan: Bisa diakses oleh KASIR
	transactionGroup.Post("/", allRoles, transactionHandler.CreateTransaction) // POST /api/v1/transactions

	// Endpoint Laporan: Hanya diakses oleh ADMIN/MANAGER
	transactionGroup.Get("/", adminManager, transactionHandler.ListTransactions)  // GET /api/v1/transactions
	transactionGroup.Get("/:id", adminManager, transactionHandler.GetTransaction) // GET /api/v1/transactions/:id
}

