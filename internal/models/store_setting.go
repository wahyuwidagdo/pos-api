package models

import "time"

type StoreSetting struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	StoreName  string    `json:"store_name" gorm:"not null;default:'My Store'"`
	Address    string    `json:"address"`
	Phone      string    `json:"phone"`
	FooterText string    `json:"footer_text" gorm:"default:'Thank you for your purchase!'"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
