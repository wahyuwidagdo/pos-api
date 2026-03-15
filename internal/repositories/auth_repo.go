package repositories

import (
	"context"
	"pos-api/internal/models"

	"gorm.io/gorm"
)

// AuthRepository mendefinisikan kontrak untuk interaksi database otentikasi.
type AuthRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByID(ctx context.Context, id uint) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
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
func (r *authRepository) CreateUser(ctx context.Context, user *models.User) error {
	result := r.DB.WithContext(ctx).Create(user)
	return result.Error
}

// GetUserByUsername mencari pengguna berdasarkan username.
func (r *authRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	result := r.DB.WithContext(ctx).Where("username = ?", username).First(&user)
	return &user, result.Error
}

// GetUserByID mencari pengguna berdasarkan ID.
func (r *authRepository) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	result := r.DB.WithContext(ctx).First(&user, id)
	return &user, result.Error
}

// UpdateUser memperbarui data pengguna.
func (r *authRepository) UpdateUser(ctx context.Context, user *models.User) error {
	result := r.DB.WithContext(ctx).Save(user)
	return result.Error
}
