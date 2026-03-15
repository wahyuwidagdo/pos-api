package models

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null"`
	SKU         string         `json:"sku" gorm:"unique"` // Stock Keeping Unit (kode unik)
	Barcode     string         `json:"barcode"`           // Unique index handled by migration (partial index where != '')
	Description string         `json:"description"`
	Price       float64        `json:"price" gorm:"type:numeric;not null"` // Harga Jual
	Cost        float64        `json:"cost" gorm:"type:numeric"`           // Harga Modal (penting untuk menghitung profit)
	Stock       int            `json:"stock" gorm:"not null"`
	CategoryID  uint           `json:"category_id"`
	Category    Category       `json:"category" gorm:"foreignKey:CategoryID"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}
