package database

import (
	"fmt"
	"log"
	"os"
	"pos-api/internal/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSL_MODE"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Opsi untuk melihat query SQL di console
	})

	if err != nil {
		log.Fatal("Gagal terhubung ke database:", err)
	}

	log.Println("Koneksi ke Database berhasil!")
	DB = db

	// Auto Migration - Membuat tabel jika belum ada
	err = db.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.Category{},
		&models.Transaction{},
		models.TransactionDetail{},
	)

	if err != nil {
		log.Fatal("Gagal menjalankan migrasi:", err)
	}
	log.Println("Migrasi tabel berhasil dijalankan.")
}

// Get DB mengembalikan instance koneksi database.
func GetDB() *gorm.DB {
	return DB
}
