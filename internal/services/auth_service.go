package services

import (
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

// AuthService mendefinisikan kontrak untuk logika otentikasi.
type AuthService interface {
	Register(req AuthRequest) (*models.User, error)
	Login(req AuthRequest) (string, error) // Mengembalikan token JWT
}

type authService struct {
	repo repositories.AuthRepository
}

// NewAuthService membuat instance AuthService baru.
func NewAuthService(repo repositories.AuthRepository) AuthService {
	return &authService{repo: repo}
}

// Register menangani logika pendaftaran pengguna baru.
func (s *authService) Register(req AuthRequest) (*models.User, error) {
	// 1. Hash Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("gagal meng-hash password")
	}

	// 2. Buat objek User
	user := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		FullName: req.FullName,
		Role:     "admin", // User pertama/utama adalah admin
	}

	// 3. Simpan ke Database via Repository
	err = s.repo.CreateUser(&user)
	if err != nil {
		// GORM mungkin mengembalikan error unik (username sudah ada)
		return nil, errors.New("username sudah terdaftar")
	}

	return &user, nil
}

// Login menangani logika login dan pembuatan token JWT.
func (s *authService) Login(req AuthRequest) (string, error) {
	// 1. Cari User di Database
	user, err := s.repo.GetUserByUsername(req.Username)
	if err != nil {
		return "", errors.New("username atau password salah")
	}

	// 2. Bandingkan Password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return "", errors.New("password salah")
	}

	// 3. Buat Token JWT
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 72).Unix(), //Token berlaku 72 jam
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Ambil secret key dari .env
	t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", errors.New("gagal membuat token")
	}

	return t, nil
}
