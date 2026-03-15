package models

import (
	"time"

	"gorm.io/gorm"
)

// PaymentMethod stores dynamic payment methods that can be managed by admin
type PaymentMethod struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"not null;unique"` // e.g., "Cash", "BCA", "OVO", "GoPay", "QRIS"
	IsCash    bool           `json:"is_cash" gorm:"default:false"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	SortOrder int            `json:"sort_order" gorm:"default:0"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}
