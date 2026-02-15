package repositories

import (
	"pos-api/internal/models"

	"gorm.io/gorm"
)

// AuthRepository mendefinisikan kontrak untuk interaksi database otentikasi.
type AuthRepository interface {
	CreateUser(user *models.User) error
	GetUserByUsername(username string) (*models.User, error)
	GetUserByID(id uint) (*models.User, error)
	UpdateUser(user *models.User) error
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
	result := r.DB.Create(user)
	return result.Error
}

// GetUserByUsername mencari pengguna berdasarkan username.
func (r *authRepository) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	result := r.DB.Where("username = ?", username).First(&user)
	return &user, result.Error
}

// GetUserByID mencari pengguna berdasarkan ID.
func (r *authRepository) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	result := r.DB.First(&user, id)
	return &user, result.Error
}

// UpdateUser memperbarui data pengguna.
func (r *authRepository) UpdateUser(user *models.User) error {
	result := r.DB.Save(user)
	return result.Error
}
