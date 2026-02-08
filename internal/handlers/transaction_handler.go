package handlers

import (
	"fmt"
	"pos-api/internal/services"

	"github.com/gofiber/fiber/v2"
)

// TransactionHandler menyimpan dependensi ke TransactionService
type TransactionHandler struct {
	service services.TransactionService
}

// NewTransactionHandler membuat instance baru dari TransactionHandler
func NewTransactionHandler(s services.TransactionService) *TransactionHandler {
	return &TransactionHandler{service: s}
}

// GetTransaction handles GET /transactions/{id}
// @Summary      Get Transaction by ID
// @Description  Retrieve a single transaction with its details. Requires Admin or Manager role.
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "Transaction ID"
// @Success      200 {object} utils.SuccessResponse{data=models.Transaction} "Transaction found"
// @Failure      400 {object} utils.ErrorResponse "Invalid transaction ID"
// @Failure      401 {object} utils.ErrorResponse "Authentication required"
// @Failure      403 {object} utils.ErrorResponse "Insufficient permissions"
// @Failure      404 {object} utils.ErrorResponse "Transaction not found"
// @Failure      500 {object} utils.ErrorResponse "Internal server error"
// @Router       /transactions/{id} [get]
func (h *TransactionHandler) GetTransaction(c *fiber.Ctx) error {
	// 1. Ambil dan parse ID dari URL parameter
	id, err := c.ParamsInt("id")
	if err != nil || id < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID transaksi tidak valid"})
	}

	// 2. Panggil Service Layer
	transaction, err := h.service.GetTransaction(uint(id))

	// 3. Handle Error dari Service Layer
	if err != nil {
		// Cek apakah error adalah 'Not Found'
		if err.Error() == fmt.Sprintf("transaksi dengan ID %d tidak ditemukan", id) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		// Error lainnya (mis. DB error)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil detail transaksi"})
	}

	// 4. Kirim Respons Sukses (200 OK)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Detail transaksi ditemukan",
		"data":    transaction,
	})
}

// ListTransactions handles GET /transactions
// @Summary      List All Transactions
// @Description  Retrieve a list of all sales transactions. Requires Admin or Manager role.
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200 {object} utils.SuccessResponse{data=[]models.Transaction} "List of transactions"
// @Failure      401 {object} utils.ErrorResponse "Authentication required"
// @Failure      403 {object} utils.ErrorResponse "Insufficient permissions"
// @Failure      500 {object} utils.ErrorResponse "Internal server error"
// @Router       /transactions [get]
func (h *TransactionHandler) ListTransactions(c *fiber.Ctx) error {
	// 1. Panggil Service Layer
	transactions, err := h.service.ListTransactions()

	// 2. Handle Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil daftar transaksi"})
	}

	// 3. Kirim Respons Sukses (200 OK)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Daftar transaksi berhasil diambil",
		"data":    transactions,
	})
}

// CreateTransaction handles POST /transactions
// @Summary      Create New Transaction (Sales)
// @Description  Process a new sales transaction. Automatically deducts stock, calculates totals, and generates invoice code. Accessible by all authenticated roles (Admin, Manager, Cashier).
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request body services.TransactionRequest true "Transaction data with items"
// @Success      201 {object} utils.SuccessResponse{data=models.Transaction} "Transaction processed successfully"
// @Failure      400 {object} utils.ErrorResponse "Invalid input, insufficient stock, or insufficient payment"
// @Failure      401 {object} utils.ErrorResponse "Authentication required"
// @Failure      403 {object} utils.ErrorResponse "Insufficient permissions"
// @Failure      404 {object} utils.ErrorResponse "Product not found"
// @Failure      500 {object} utils.ErrorResponse "Internal server error"
// @Router       /transactions [post]
func (h *TransactionHandler) CreateTransaction(c *fiber.Ctx) error {
	// 1. Inisiasi DTO dan Binding Request Body
	var req services.TransactionRequest
	if err := c.BodyParser(&req); err != nil {
		// Jika gagal parse JSON, kembalikan Bad Request
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Permintaan tidak valid",
			"detail": err.Error(),
		})
	}

	// 2. Panggil Service Layer untuk memproses logika bisnis
	// Validasi input sudah di-handle di dalam service.ProcessTransaction
	transaction, err := h.service.ProcessTransaction(req)

	// 3. Handle Error dari Service Layer
	if err != nil {
		// Logika bisnis gagal (validasi, stok kurang, uang kurang, DB error, dll.)
		// Asumsikan semua error dari service adalah Bad Request (400) atau Internal Server Error (500)
		// Kita bisa melakukan pengecekan error yang lebih detail di sini, tapi untuk sementara,
		// kita kembalikan 400 untuk error yang berasal dari input/logika bisnis.
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(), // Menggunakan pesan error dari service
		})
	}

	// 4. Kirim Respons Sukses (201 Created)
	// Kita kembalikan struct transaksi yang sudah terisi ID dan data lainnya
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Transaksi berhasil diproses",
		"data":    transaction,
	})
}
