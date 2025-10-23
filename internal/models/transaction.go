package models

import (
	"time"
)

type Transaction struct {
	ID                 uint                `json:"id" gorm:"primaryKey"`
	TransactionCode    string              `json:"transaction_code" gorm:"unique;not null"`   // Contoh: INV-20231016-0001
	TotalAmount        float64             `json:"total_amount" gorm:"type:numeric;not null"` // Total sebelum diskon/pajak
	Discount           float64             `json:"discount" gorm:"type:numeric"`
	GrandTotal         float64             `json:"grand_total" gorm:"type:numeric;not null"`            // Total akhir yang harus dibayar
	Cash               float64             `json:"cash" gorm:"type:numeric;not null"`                   // Uang tunai yang dibayarkan pelanggan
	Change             float64             `json:"change" gorm:"type:numeric;not null"`                 // Uang kembalian
	PaymentMethod      string              `json:"payment_method"`                                      // e.g., "Cash", "QRIS"
	TransactionDetails []TransactionDetail `json:"transaction_details" gorm:"foreignKey:TransactionID"` // Relasi ke detail
	CreatedAt          time.Time           `json:"created_at"`
}
