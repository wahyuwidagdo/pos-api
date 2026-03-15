package services

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"

	"pos-api/internal/models"
	"pos-api/internal/repositories"
)

// AuthRequest mendefinisikan DTO (Data Transfer Object) untuk login/register request
type AuthRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	FullName string `json:"full_name"` // Hanya untuk register
}

// UpdateProfileRequest untuk update profile
type UpdateProfileRequest struct {
	FullName string `json:"full_name"`
	Username string `json:"username"`
}

// ChangePasswordRequest untuk ganti password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required"`
}

// AuthService mendefinisikan kontrak untuk logika otentikasi.
type AuthService interface {
	Register(ctx context.Context, req AuthRequest) (*models.User, error)
	Login(ctx context.Context, req AuthRequest) (string, *models.User, error) // Mengembalikan token JWT + user
	GetProfile(ctx context.Context, userID uint) (*models.User, error)
	UpdateProfile(ctx context.Context, userID uint, req UpdateProfileRequest) (*models.User, error)
	ChangePassword(ctx context.Context, userID uint, req ChangePasswordRequest) error
}

type authService struct {
	repo repositories.AuthRepository
}

// NewAuthService membuat instance AuthService baru.
func NewAuthService(repo repositories.AuthRepository) AuthService {
	return &authService{repo: repo}
}

// Register menangani logika pendaftaran pengguna baru.
func (s *authService) Register(ctx context.Context, req AuthRequest) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("gagal meng-hash password")
	}

	user := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		FullName: req.FullName,
		Role:     "admin",
	}

	err = s.repo.CreateUser(ctx, &user)
	if err != nil {
		return nil, errors.New("username sudah terdaftar")
	}

	return &user, nil
}

// Login menangani logika login dan pembuatan token JWT.
func (s *authService) Login(ctx context.Context, req AuthRequest) (string, *models.User, error) {
	user, err := s.repo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return "", nil, errors.New("username atau password salah")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return "", nil, errors.New("password salah")
	}

	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", nil, errors.New("gagal membuat token")
	}

	return t, user, nil
}

// GetProfile mengembalikan data profil pengguna berdasarkan ID.
func (s *authService) GetProfile(ctx context.Context, userID uint) (*models.User, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("pengguna tidak ditemukan")
	}
	return user, nil
}

// UpdateProfile memperbarui profil pengguna.
func (s *authService) UpdateProfile(ctx context.Context, userID uint, req UpdateProfileRequest) (*models.User, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("pengguna tidak ditemukan")
	}

	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.Username != "" && req.Username != user.Username {
		// Check if new username is already taken
		existing, _ := s.repo.GetUserByUsername(ctx, req.Username)
		if existing != nil && existing.ID != userID {
			return nil, errors.New("username sudah digunakan")
		}
		user.Username = req.Username
	}

	err = s.repo.UpdateUser(ctx, user)
	if err != nil {
		return nil, errors.New("gagal memperbarui profil")
	}

	return user, nil
}

// ChangePassword mengganti password pengguna.
func (s *authService) ChangePassword(ctx context.Context, userID uint, req ChangePasswordRequest) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return errors.New("pengguna tidak ditemukan")
	}

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword))
	if err != nil {
		return errors.New("password saat ini salah")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("gagal meng-hash password baru")
	}

	user.Password = string(hashedPassword)
	err = s.repo.UpdateUser(ctx, user)
	if err != nil {
		return errors.New("gagal memperbarui password")
	}

	return nil
}
