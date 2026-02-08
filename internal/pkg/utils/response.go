package utils

import "github.com/gofiber/fiber/v2"

// SuccessResponse mendefinisikan format response untuk operasi yang sukses.
// Digunakan untuk pesan sukses sederhana, item tunggal, atau operasi auth (token).
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"` // Bisa berisi struct (item tunggal) atau string (token)
}

// PagedResponse mendefinisikan format response untuk daftar data (array) dengan informasi pagination.
type PagedResponse struct {
	Message  string      `json:"message"`
	Data     interface{} `json:"data,omitempty"` // Array data (misal: []Product)
	Page     int         `json:"page,omitempty"`
	PageSize int         `json:"pageSize,omitempty"`
	// Total    int         `json:"total,omitempty"`
}

// ErrorResponse mendefinisikan format response untuk kesalahan.
type ErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// ------------------- HELPER FUNCTIONS -------------------

// JSONSuccess digunakan untuk item tunggal (Create/GetByID/Update) atau pesan sederhana.
func JSONSuccess(c *fiber.Ctx, statusCode int, message string, data interface{}) error {
	return c.Status(statusCode).JSON(SuccessResponse{
		Message: message,
		Data:    data,
	})
}

// JSONPaged digunakan untuk daftar data (List).
func JSONPaged(c *fiber.Ctx, message string, data interface{}, page, pageSize int) error {
	return c.Status(fiber.StatusOK).JSON(PagedResponse{
		Message:  message,
		Data:     data,
		Page:     page,
		PageSize: pageSize,
		// Total:    total,
	})
}

// JSONError digunakan untuk respons error.
func JSONError(c *fiber.Ctx, statusCode int, message string, err string) error {
	return c.Status(statusCode).JSON(ErrorResponse{
		Message: message,
		Error:   err,
	})
}
