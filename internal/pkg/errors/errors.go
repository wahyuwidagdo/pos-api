package errors

import "errors"

// Definisi Custom Errors untuk lapisan Service
var (
	ErrNotFound             = errors.New("not found")                        // 404
	ErrConflict             = errors.New("conflict")                         // 409 (Data duplikat, dll.)
	ErrInsufficientStock    = errors.New("insufficient stock")               // 400
	ErrPaymantRequired      = errors.New("payment required")                 // 402 (Uang kurang)
	ErrForeignKeyConstraint = errors.New("foreign key constraint violation") // 400/409
)

// Gunakan fungsi ini di Service Layer
func Is(err, target error) bool {
	return errors.Is(err, target)
}
