package models

import "time"

type Product struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	SKU         string    `json:"sku" gorm:"unique"` // Stock Keeping Unit (kode unik)
	Description string    `json:"description"`
	Price       float64   `json:"price" gorm:"type:numeric;not null"` // Harga Jual
	Cost        float64   `json:"cost" gorm:"type:numeric"`           // Harga Modal (penting untuk menghitung profit)
	Stock       int       `json:"stock" gorm:"not null"`
	CategoryID  uint      `json:"category_id"`
	Category    Category  `json:"category" gorm:"foreignKey:CategoryID"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
