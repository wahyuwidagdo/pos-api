package models

import (
	"time"
)

// Shift represents a cashier shift with cash drawer management
type Shift struct {
	ID            uint       `json:"id" gorm:"primaryKey"`
	UserID        uint       `json:"user_id" gorm:"not null"`
	User          User       `json:"user" gorm:"foreignKey:UserID"`
	StartingCash  float64    `json:"starting_cash" gorm:"type:numeric;not null"`  // Modal awal kasir
	EndingCash    float64    `json:"ending_cash" gorm:"type:numeric"`              // Uang di laci saat tutup
	ExpectedCash  float64    `json:"expected_cash" gorm:"type:numeric"`            // Yang seharusnya (hitung dari transaksi)
	CashDiff      float64    `json:"cash_difference" gorm:"type:numeric"`          // Selisih (bisa + atau -)
	TotalSales    float64    `json:"total_sales" gorm:"type:numeric"`              // Total penjualan shift ini
	TotalTx       int        `json:"total_transactions" gorm:"default:0"`          // Jumlah transaksi
	Status        string     `json:"status" gorm:"default:'open'"`                 // open, closed
	Notes         string     `json:"notes"`                                        // Catatan kasir
	OpenedAt      time.Time  `json:"opened_at" gorm:"not null"`
	ClosedAt      *time.Time `json:"closed_at"`
}
