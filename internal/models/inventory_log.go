package models

import (
	"time"

	"gorm.io/gorm"
)

// InventoryLog tracks all stock movements (In, Out, Adjustment)
type InventoryLog struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	ProductID   uint           `json:"product_id" gorm:"not null;index"`
	Product     Product        `json:"product" gorm:"foreignKey:ProductID"`
	Type        string         `json:"type" gorm:"not null"` // "in", "out", "adjustment"
	Source      string         `json:"source"`               // "purchase", "return", "damage", "expired", "opname", "sale", "audit"
	Quantity    int            `json:"quantity" gorm:"not null"`
	CostPrice   float64        `json:"cost_price" gorm:"type:numeric"` // Cost per unit at time of entry
	TotalCost   float64        `json:"total_cost" gorm:"type:numeric"` // quantity * cost_price
	StockBefore int            `json:"stock_before" gorm:"not null"`   // Stock level before this entry
	StockAfter  int            `json:"stock_after" gorm:"not null"`    // Stock level after this entry
	Notes       string         `json:"notes"`
	UserID      uint           `json:"user_id" gorm:"not null;index"`
	User        User           `json:"user" gorm:"foreignKey:UserID"`
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}
