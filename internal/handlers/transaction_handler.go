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

// GetTransaction menangani request GET /transactions/:id
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

// ListTransactions menangani request GET /transactions
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

// CreateTransaction menangani request POST /transactions
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
