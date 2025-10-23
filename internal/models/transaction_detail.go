package models

type TransactionDetail struct {
	ID            uint    `json:"id" gorm:"primaryKey"`
	TransactionID uint    `json:"transaction_id"`
	ProductID     uint    `json:"product_id"`
	ProductName   string  `json:"product_name"` // Cache nama produk (jika produk diubah, histori transaksi tetap benar)
	Quantity      int     `json:"quantity" gorm:"not null"`
	PriceAtSale   float64 `json:"price_at_sale" gorm:"type:numeric;not null"` // Harga jual saat transaksi terjadi
	SubTotal      float64 `json:"subtotal" gorm:"type:numeric;not null"`      // Quantity * PriceAtSale
	Product       Product `json:"product" gorm:"foreignKey:ProductID"`
}
