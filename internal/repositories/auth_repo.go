package repositories

import (
	"pos-api/internal/models"

	"gorm.io/gorm"
)

// AuthRepository mendefinisikan kontrak untuk interaksi database otentikasi.
type AuthRepository interface {
	CreateUser(user *models.User) error
	GetUserByUsername(username string) (*models.User, error)
}

type authRepository struct {
	DB *gorm.DB
}

// NewAuthRepository membuat instance AuthRepository baru.
func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{
		DB: db,
	}
}

// CreateUser menyimpan pengguna baru ke database.
func (r *authRepository) CreateUser(user *models.User) error {
	// database.GetDB() mengambil instance GORM yang sudah kita buat.
	result := r.DB.Create(user)
	return result.Error
}

// GetUserByUsername mencari pengguna berdasarkan username.
func (r *authRepository) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	// First mencari satu record.
	result := r.DB.Where("username = ?", username).First(&user)

	// GORM mengembalikan error gorm.ErrRecordNotFound jika tidak ada
	return &user, result.Error
}
