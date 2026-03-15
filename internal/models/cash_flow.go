package models

import (
	"time"

	"gorm.io/gorm"
)

// CashFlow tracks income and expenses beyond sales
type CashFlow struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Type      string         `json:"type" gorm:"not null"`   // "income" or "expense"
	Source    string         `json:"source" gorm:"not null"` // e.g., "sales", "rent", "electricity", "supplies", "salary", "other_income", "other_expense"
	Amount    float64        `json:"amount" gorm:"type:numeric;not null"`
	Date      time.Time      `json:"date" gorm:"not null;index"`
	Notes     string         `json:"notes"`
	UserID    uint           `json:"user_id" gorm:"not null;index"`
	User      User           `json:"user" gorm:"foreignKey:UserID"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}
