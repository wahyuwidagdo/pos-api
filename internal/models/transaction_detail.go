package models

type TransactionDetail struct {
	ID            uint    `json:"id" gorm:"primaryKey"`
	TransactionID uint    `json:"transaction_id"`
	ProductID     uint    `json:"product_id"`
	ProductName   string  `json:"product_name"` // Cache nama produk (jika produk diubah, histori transaksi tetap benar)
	Quantity      int     `json:"quantity" gorm:"not null"`
	PriceAtSale   float64 `json:"price_at_sale" gorm:"type:numeric;not null"` // Harga jual saat transaksi terjadi
	CostAtSale    float64 `json:"cost_at_sale" gorm:"type:numeric;default:0"` // Harga beli saat transaksi (untuk laporan laba)
	SubTotal      float64 `json:"subtotal" gorm:"type:numeric;not null"`      // Quantity * PriceAtSale
	Product       Product `json:"product" gorm:"foreignKey:ProductID"`
}
