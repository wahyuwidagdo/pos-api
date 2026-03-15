package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"pos-api/internal/config"
	"pos-api/internal/models"
	"pos-api/pkg/database"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	// 1. Load Configuration
	cfg := config.LoadConfig()

	// 2. Connect to Database
	database.ConnectDB(cfg)
	db := database.DB

	log.Println("🌱 Starting Database Seeding...")

	seedStoreSettings(db)
	seedUsers(db)
	seedPaymentMethods(db)
	categories := seedCategories(db)
	products := seedProducts(db, categories)
	seedInventory(db, products)
	seedCashFlow(db)
	seedTransactions(db, products)

	log.Println("✅ Database Seeding Completed Successfully!")
}

func seedStoreSettings(db *gorm.DB) {
	var setting models.StoreSetting
	if err := db.First(&setting).Error; err == nil {
		log.Println("Store Settings already exist, skipping...")
		return
	}

	setting = models.StoreSetting{
		StoreName:  "Toko Serba Ada",
		Address:    "Jl. Merdeka No. 45, Jakarta Selatan",
		Phone:      "081234567890",
		FooterText: "Terima kasih telah berbelanja di Toko Serba Ada!",
	}

	if err := db.Create(&setting).Error; err != nil {
		log.Fatalf("Failed to seed store settings: %v", err)
	}
	log.Println("Store Settings seeded.")
}

func seedUsers(db *gorm.DB) {
	users := []models.User{
		{
			Username: "admin",
			FullName: "Administrator",
			Role:     "admin",
		},
		{
			Username: "kasir",
			FullName: "Kasir Utama",
			Role:     "kasir",
		},
	}

	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	for _, u := range users {
		var existingUser models.User
		if err := db.Where("username = ?", u.Username).First(&existingUser).Error; err == nil {
			log.Printf("User %s already exists, skipping...", u.Username)
			continue
		}

		u.Password = string(hashedPassword)
		if err := db.Create(&u).Error; err != nil {
			log.Fatalf("Failed to seed user %s: %v", u.Username, err)
		}
		log.Printf("User %s seeded.", u.Username)
	}
}

func seedCategories(db *gorm.DB) []models.Category {
	categoryNames := []string{
		"Makanan", "Minuman", "Snack", "Sembako", "Elektronik",
		"Pakaian Pria", "Pakaian Wanita", "Aksesoris HP", "Alat Tulis", "Kesehatan",
		"Dapur", "Otomotif", "Mainan", "Hobi", "Perabotan",
	}
	var categories []models.Category

	for _, name := range categoryNames {
		var cat models.Category
		if err := db.Where("name = ?", name).First(&cat).Error; err != nil {
			cat = models.Category{Name: name}
			if err := db.Create(&cat).Error; err != nil {
				log.Fatalf("Failed to seed category %s: %v", name, err)
			}
			log.Printf("Category %s seeded.", name)
		} else {
			// log.Printf("Category %s already exists.", name)
		}
		categories = append(categories, cat)
	}
	return categories
}

func seedPaymentMethods(db *gorm.DB) {
	methods := []models.PaymentMethod{
		{Name: "Cash", IsCash: true, IsActive: true, SortOrder: 1},
		{Name: "QRIS", IsCash: false, IsActive: true, SortOrder: 2},
		{Name: "Debit Card", IsCash: false, IsActive: true, SortOrder: 3},
	}

	for _, m := range methods {
		var existing models.PaymentMethod
		if err := db.Where("name = ?", m.Name).First(&existing).Error; err == nil {
			log.Printf("Payment Method %s already exists.", m.Name)
			continue
		}

		if err := db.Create(&m).Error; err != nil {
			log.Fatalf("Failed to seed payment method %s: %v", m.Name, err)
		}
		log.Printf("Payment Method %s seeded.", m.Name)
	}
}

func seedProducts(db *gorm.DB, categories []models.Category) []models.Product {
	// Helper to find category ID by name
	getCatID := func(name string) uint {
		for _, c := range categories {
			if c.Name == name {
				return c.ID
			}
		}
		return categories[0].ID // Fallback
	}

	getBarcode := func(idx int) string {
		return fmt.Sprintf("8%d%04d", time.Now().UnixMicro(), idx)
	}

	products := []models.Product{
		// Makanan & Minuman
		{Name: "Nasi Goreng Spesial", SKU: "FOOD-001", Barcode: getBarcode(1), Price: 25000, Cost: 15000, Stock: 100, CategoryID: getCatID("Makanan")},
		{Name: "Ayam Bakar Madu", SKU: "FOOD-002", Barcode: getBarcode(2), Price: 30000, Cost: 18000, Stock: 80, CategoryID: getCatID("Makanan")},
		{Name: "Sate Ayam Madura", SKU: "FOOD-003", Barcode: getBarcode(3), Price: 22000, Cost: 12000, Stock: 50, CategoryID: getCatID("Makanan")},
		{Name: "Mie Goreng Jawa", SKU: "FOOD-004", Barcode: getBarcode(4), Price: 20000, Cost: 10000, Stock: 0, CategoryID: getCatID("Makanan")}, // Out of stock
		{Name: "Es Teh Manis", SKU: "DRINK-001", Barcode: getBarcode(5), Price: 5000, Cost: 2000, Stock: 200, CategoryID: getCatID("Minuman")},
		{Name: "Kopi Susu Gula Aren", SKU: "DRINK-002", Barcode: getBarcode(6), Price: 18000, Cost: 8000, Stock: 150, CategoryID: getCatID("Minuman")},
		{Name: "Jus Jeruk Segar", SKU: "DRINK-003", Barcode: getBarcode(7), Price: 15000, Cost: 7000, Stock: 5, CategoryID: getCatID("Minuman")}, // Low stock
		{Name: "Air Mineral 600ml", SKU: "DRINK-004", Barcode: getBarcode(8), Price: 5000, Cost: 2500, Stock: 300, CategoryID: getCatID("Minuman")},

		// Snack
		{Name: "Keripik Singkong", SKU: "SNACK-001", Barcode: getBarcode(9), Price: 10000, Cost: 5000, Stock: 50, CategoryID: getCatID("Snack")},
		{Name: "Chitato Lite", SKU: "SNACK-002", Barcode: getBarcode(10), Price: 12000, Cost: 9000, Stock: 40, CategoryID: getCatID("Snack")},
		{Name: "Oreo Original", SKU: "SNACK-003", Barcode: getBarcode(11), Price: 8000, Cost: 5000, Stock: 2, CategoryID: getCatID("Snack")}, // Low stock
		{Name: "Silverqueen Chunky", SKU: "SNACK-004", Barcode: getBarcode(12), Price: 25000, Cost: 18000, Stock: 60, CategoryID: getCatID("Snack")},

		// Sembako
		{Name: "Beras 5kg", SKU: "SEMBAKO-001", Barcode: getBarcode(13), Price: 70000, Cost: 60000, Stock: 20, CategoryID: getCatID("Sembako")},
		{Name: "Minyak Goreng 2L", SKU: "SEMBAKO-002", Barcode: getBarcode(14), Price: 35000, Cost: 30000, Stock: 30, CategoryID: getCatID("Sembako")},
		{Name: "Gula Pasir 1kg", SKU: "SEMBAKO-003", Barcode: getBarcode(15), Price: 16000, Cost: 13000, Stock: 0, CategoryID: getCatID("Sembako")}, // Out of stock
		{Name: "Telur Ayam 1kg", SKU: "SEMBAKO-004", Barcode: getBarcode(16), Price: 28000, Cost: 24000, Stock: 15, CategoryID: getCatID("Sembako")},

		// Elektronik & Aksesoris HP
		{Name: "Kabel Data Type-C", SKU: "ELEC-001", Barcode: getBarcode(17), Price: 25000, Cost: 10000, Stock: 50, CategoryID: getCatID("Aksesoris HP")},
		{Name: "Charger Samsung 25W", SKU: "ELEC-002", Barcode: getBarcode(18), Price: 150000, Cost: 100000, Stock: 10, CategoryID: getCatID("Aksesoris HP")},
		{Name: "Powerbank 10000mAh", SKU: "ELEC-003", Barcode: getBarcode(19), Price: 200000, Cost: 150000, Stock: 5, CategoryID: getCatID("Aksesoris HP")}, // Low stock
		{Name: "Earphone Bluetooth", SKU: "ELEC-004", Barcode: getBarcode(20), Price: 120000, Cost: 80000, Stock: 25, CategoryID: getCatID("Elektronik")},
		{Name: "Mouse Wireless", SKU: "ELEC-005", Barcode: getBarcode(21), Price: 85000, Cost: 50000, Stock: 30, CategoryID: getCatID("Elektronik")},

		// Alat Tulis
		{Name: "Pulpen Pilot", SKU: "ATK-001", Price: 3000, Cost: 1500, Stock: 200, CategoryID: getCatID("Alat Tulis")},
		{Name: "Buku Tulis Sidu", SKU: "ATK-002", Price: 5000, Cost: 3000, Stock: 150, CategoryID: getCatID("Alat Tulis")},
		{Name: "Pensil 2B", SKU: "ATK-003", Price: 2000, Cost: 800, Stock: 3, CategoryID: getCatID("Alat Tulis")}, // Low stock
		{Name: "Penghapus", SKU: "ATK-004", Price: 1000, Cost: 300, Stock: 100, CategoryID: getCatID("Alat Tulis")},

		// Dapur & Perabotan
		{Name: "Piring Keramik", SKU: "HOME-001", Price: 15000, Cost: 8000, Stock: 40, CategoryID: getCatID("Dapur")},
		{Name: "Gelas Kaca", SKU: "HOME-002", Price: 8000, Cost: 4000, Stock: 60, CategoryID: getCatID("Dapur")},
		{Name: "Sapu Ijuk", SKU: "HOME-003", Price: 25000, Cost: 15000, Stock: 10, CategoryID: getCatID("Perabotan")},
		{Name: "Kain Pel", SKU: "HOME-004", Price: 30000, Cost: 18000, Stock: 0, CategoryID: getCatID("Perabotan")}, // Out of Stock

		// Kesehatan
		{Name: "Masker Medis 1 Box", SKU: "HEALTH-001", Price: 25000, Cost: 15000, Stock: 100, CategoryID: getCatID("Kesehatan")},
		{Name: "Hand Sanitizer 100ml", SKU: "HEALTH-002", Price: 15000, Cost: 8000, Stock: 50, CategoryID: getCatID("Kesehatan")},
		{Name: "Vitamin C 500mg", SKU: "HEALTH-003", Price: 45000, Cost: 30000, Stock: 20, CategoryID: getCatID("Kesehatan")},
		{Name: "Paracetamol Strip", SKU: "HEALTH-004", Price: 5000, Cost: 3000, Stock: 200, CategoryID: getCatID("Kesehatan")},

		// Pakaian
		{Name: "Kaos Polos Hitam XL", SKU: "FASHION-001", Price: 50000, Cost: 30000, Stock: 30, CategoryID: getCatID("Pakaian Pria")},
		{Name: "Kemeja Flanel L", SKU: "FASHION-002", Price: 120000, Cost: 80000, Stock: 15, CategoryID: getCatID("Pakaian Pria")},
		{Name: "Rok Plisket", SKU: "FASHION-003", Price: 75000, Cost: 50000, Stock: 20, CategoryID: getCatID("Pakaian Wanita")},
		{Name: "Jilbab Pashmina", SKU: "FASHION-004", Price: 35000, Cost: 20000, Stock: 40, CategoryID: getCatID("Pakaian Wanita")},

		// Otomotif
		{Name: "Oli Mesin Matic", SKU: "AUTO-001", Price: 55000, Cost: 40000, Stock: 20, CategoryID: getCatID("Otomotif")},
		{Name: "Kanebo", SKU: "AUTO-002", Price: 15000, Cost: 8000, Stock: 30, CategoryID: getCatID("Otomotif")},
		{Name: "Pewangi Mobil", SKU: "AUTO-003", Price: 20000, Cost: 12000, Stock: 3, CategoryID: getCatID("Otomotif")}, // Low stock

		// Mainan & Hobi
		{Name: "Hot Wheels", SKU: "TOY-001", Price: 30000, Cost: 20000, Stock: 50, CategoryID: getCatID("Mainan")},
		{Name: "Lego Minifigure", SKU: "TOY-002", Price: 45000, Cost: 30000, Stock: 15, CategoryID: getCatID("Mainan")},
		{Name: "Benang Rajut", SKU: "HOBBY-001", Price: 12000, Cost: 7000, Stock: 40, CategoryID: getCatID("Hobi")},
		{Name: "Cat Akrilik 12 Warna", SKU: "HOBBY-002", Price: 35000, Cost: 22000, Stock: 10, CategoryID: getCatID("Hobi")},

		// Fillers to reach 50+
		{Name: "Roti Tawar Kupas", SKU: "FOOD-005", Price: 18000, Cost: 12000, Stock: 10, CategoryID: getCatID("Makanan")},
		{Name: "Susu UHT 1L", SKU: "DRINK-005", Price: 22000, Cost: 18000, Stock: 24, CategoryID: getCatID("Minuman")},
		{Name: "Coklat Batang", SKU: "SNACK-005", Price: 15000, Cost: 10000, Stock: 4, CategoryID: getCatID("Snack")},        // Low
		{Name: "Tepung Terigu 1kg", SKU: "SEMBAKO-005", Price: 12000, Cost: 9000, Stock: 0, CategoryID: getCatID("Sembako")}, // Out
		{Name: "Kabel HDMI", SKU: "ELEC-006", Price: 45000, Cost: 25000, Stock: 15, CategoryID: getCatID("Elektronik")},
		{Name: "Spidol Papan Tulis", SKU: "ATK-005", Price: 8000, Cost: 4000, Stock: 30, CategoryID: getCatID("Alat Tulis")},
		{Name: "Wajan Teflon 24cm", SKU: "HOME-005", Price: 85000, Cost: 60000, Stock: 12, CategoryID: getCatID("Dapur")},
		{Name: "Sabun Cuci Piring", SKU: "HOME-006", Price: 15000, Cost: 10000, Stock: 40, CategoryID: getCatID("Perabotan")},
		{Name: "Minyak Kayu Putih 60ml", SKU: "HEALTH-005", Price: 18000, Cost: 12000, Stock: 25, CategoryID: getCatID("Kesehatan")},
	}

	var seededProducts []models.Product
	for _, p := range products {
		var existing models.Product
		if err := db.Where("sku = ?", p.SKU).First(&existing).Error; err == nil {
			log.Printf("Product %s already exists.", p.Name)
			seededProducts = append(seededProducts, existing)
			continue
		}

		if err := db.Create(&p).Error; err != nil {
			log.Fatalf("Failed to seed product %s: %v", p.Name, err)
		}
		log.Printf("Product %s seeded.", p.Name)
		seededProducts = append(seededProducts, p)
	}
	return seededProducts
}

func seedInventory(db *gorm.DB, products []models.Product) {
	// Assuming first user is admin/system
	var user models.User
	db.First(&user)

	for _, p := range products {
		var count int64
		db.Model(&models.InventoryLog{}).Where("product_id = ? AND type = ? AND source = ?", p.ID, "in", "purchase").Count(&count)

		if count > 0 {
			// Log already exists for this product
			continue
		}

		logEntry := models.InventoryLog{
			ProductID:   p.ID,
			Type:        "in",
			Source:      "purchase",
			Quantity:    p.Stock, // Initial stock from product definition
			CostPrice:   p.Cost,
			TotalCost:   float64(p.Stock) * p.Cost,
			StockBefore: 0,
			StockAfter:  p.Stock,
			Notes:       "Initial Stock Seed",
			UserID:      user.ID,
			CreatedAt:   time.Now().Add(-24 * 30 * time.Hour), // 30 days ago
		}

		if err := db.Create(&logEntry).Error; err != nil {
			log.Printf("Failed to seed inventory for %s: %v", p.Name, err)
		} else {
			log.Printf("Inventory log seeded for %s.", p.Name)
		}
	}
	log.Println("Inventory logs check completed.")
}

func seedCashFlow(db *gorm.DB) {
	var count int64
	db.Model(&models.CashFlow{}).Count(&count)
	if count > 0 {
		log.Println("Cash Flow data already exist, skipping...")
		return
	}

	var user models.User
	db.First(&user)

	flows := []models.CashFlow{
		{Type: "income", Source: "modal_awal", Amount: 10000000, Notes: "Modal Awal Toko", Date: time.Now().Add(-30 * 24 * time.Hour)},
		{Type: "income", Source: "modal_tambahan", Amount: 5000000, Notes: "Tambahan Modal dari Pemilik", Date: time.Now().Add(-15 * 24 * time.Hour)},
		{Type: "expense", Source: "sewa", Amount: 2000000, Notes: "Sewa Ruko Bulan Ini", Date: time.Now().Add(-25 * 24 * time.Hour)},
		{Type: "expense", Source: "listrik", Amount: 500000, Notes: "Listrik Bulan Ini", Date: time.Now().Add(-5 * 24 * time.Hour)},
		{Type: "expense", Source: "gaji_karyawan", Amount: 3000000, Notes: "Gaji Kasir Utama", Date: time.Now().Add(-3 * 24 * time.Hour)},
	}

	for _, f := range flows {
		f.UserID = user.ID
		if err := db.Create(&f).Error; err != nil {
			log.Printf("Failed to seed cash flow: %v", err)
		}
	}
	log.Println("Cash Flow seeded.")
}

func seedTransactions(db *gorm.DB, products []models.Product) {
	var count int64
	db.Model(&models.Transaction{}).Count(&count)
	if count > 0 {
		log.Println("Transactions already exist, skipping...")
		return
	}

	log.Println("Seeding Transactions (this might take a moment)...")

	// Generate 50 dummy transactions over the last 30 days
	startDate := time.Now().Add(-30 * 24 * time.Hour)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	paymentMethods := []string{"Cash", "QRIS", "Debit Card"}

	for i := 0; i < 50; i++ {
		txDate := startDate.Add(time.Duration(r.Intn(30*24)) * time.Hour)

		// Pick 1-5 random products for this transaction
		numItems := r.Intn(5) + 1
		var details []models.TransactionDetail
		var totalAmount float64

		for j := 0; j < numItems; j++ {
			prod := products[r.Intn(len(products))]
			qty := r.Intn(3) + 1
			subTotal := float64(qty) * prod.Price

			details = append(details, models.TransactionDetail{
				ProductID:   prod.ID,
				ProductName: prod.Name,
				Quantity:    qty,
				PriceAtSale: prod.Price,
				CostAtSale:  prod.Cost,
				SubTotal:    subTotal,
			})
			totalAmount += subTotal
		}

		paymentMethod := paymentMethods[r.Intn(len(paymentMethods))]

		tx := models.Transaction{
			TransactionCode:    fmt.Sprintf("INV-%s-%04d", txDate.Format("20060102"), i+1),
			TotalAmount:        totalAmount,
			Discount:           0,
			GrandTotal:         totalAmount,
			Cash:               totalAmount, // Assume exact payment for simplicity
			Change:             0,
			PaymentMethod:      paymentMethod,
			TransactionDetails: details,
			CreatedAt:          txDate,
		}

		if err := db.Create(&tx).Error; err != nil {
			log.Printf("Failed to seed transaction %d: %v", i, err)
		}
	}
	log.Println("Transactions seeded.")
}
