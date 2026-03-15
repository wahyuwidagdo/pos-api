package database

import (
	"log"
	"pos-api/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB(cfg *config.Config) {
	dsn := cfg.DBUrl

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Opsi untuk melihat query SQL di console
	})

	if err != nil {
		log.Fatal("Gagal terhubung ke database:", err)
	}

	log.Println("Koneksi ke Database berhasil!")
	DB = db

	log.Println("Koneksi ke Database berhasil!")
	DB = db
}

// Get DB mengembalikan instance koneksi database.
func GetDB() *gorm.DB {
	return DB
}
