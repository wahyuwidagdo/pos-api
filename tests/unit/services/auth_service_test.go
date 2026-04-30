package services_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"pos-api/internal/models"
	"pos-api/internal/services"
	"pos-api/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// setupAuthTest creates a fresh mock and service for each test.
func setupAuthTest(t *testing.T) (*mocks.AuthRepository, services.AuthService) {
	mockRepo := mocks.NewAuthRepository(t)
	service := services.NewAuthService(mockRepo)
	os.Setenv("JWT_SECRET", "test-secret-key-for-unit-tests")
	return mockRepo, service
}

// --- Register ---

func TestAuthService_Register_Success(t *testing.T) {
	mockRepo, service := setupAuthTest(t)
	ctx := context.Background()

	mockRepo.On("CreateUser", ctx, mock.AnythingOfType("*models.User")).Return(nil).Once()

	user, err := service.Register(ctx, services.AuthRequest{
		Username: "newuser",
		Password: "password123",
		FullName: "New User",
	})

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "newuser", user.Username)
	assert.Equal(t, "admin", user.Role)
}

func TestAuthService_Register_DuplicateUsername(t *testing.T) {
	mockRepo, service := setupAuthTest(t)
	ctx := context.Background()

	mockRepo.On("CreateUser", ctx, mock.AnythingOfType("*models.User")).Return(errors.New("unique constraint")).Once()

	user, err := service.Register(ctx, services.AuthRequest{
		Username: "existing",
		Password: "password123",
	})

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "sudah terdaftar")
}

// --- Login ---

func TestAuthService_Login_Success(t *testing.T) {
	mockRepo, service := setupAuthTest(t)
	ctx := context.Background()

	hashedPw, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	existingUser := &models.User{
		ID:       1,
		Username: "admin",
		Password: string(hashedPw),
		Role:     "admin",
	}

	mockRepo.On("GetUserByUsername", ctx, "admin").Return(existingUser, nil).Once()

	token, user, err := service.Login(ctx, services.AuthRequest{
		Username: "admin",
		Password: "password123",
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotNil(t, user)
	assert.Equal(t, "admin", user.Username)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	mockRepo, service := setupAuthTest(t)
	ctx := context.Background()

	hashedPw, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	existingUser := &models.User{
		ID:       1,
		Username: "admin",
		Password: string(hashedPw),
	}

	mockRepo.On("GetUserByUsername", ctx, "admin").Return(existingUser, nil).Once()

	token, user, err := service.Login(ctx, services.AuthRequest{
		Username: "admin",
		Password: "wrongpassword",
	})

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "password salah")
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	mockRepo, service := setupAuthTest(t)
	ctx := context.Background()

	mockRepo.On("GetUserByUsername", ctx, "nonexistent").Return(nil, errors.New("record not found")).Once()

	token, user, err := service.Login(ctx, services.AuthRequest{
		Username: "nonexistent",
		Password: "password123",
	})

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "username atau password salah")
}

// --- GetProfile ---

func TestAuthService_GetProfile_Success(t *testing.T) {
	mockRepo, service := setupAuthTest(t)
	ctx := context.Background()

	expectedUser := &models.User{ID: 1, Username: "admin", FullName: "Administrator"}
	mockRepo.On("GetUserByID", ctx, uint(1)).Return(expectedUser, nil).Once()

	user, err := service.GetProfile(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
}

func TestAuthService_GetProfile_NotFound(t *testing.T) {
	mockRepo, service := setupAuthTest(t)
	ctx := context.Background()

	mockRepo.On("GetUserByID", ctx, uint(999)).Return(nil, errors.New("record not found")).Once()

	user, err := service.GetProfile(ctx, 999)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "tidak ditemukan")
}

// --- UpdateProfile ---

func TestAuthService_UpdateProfile_Success(t *testing.T) {
	mockRepo, service := setupAuthTest(t)
	ctx := context.Background()

	existingUser := &models.User{ID: 1, Username: "admin", FullName: "Old Name"}
	mockRepo.On("GetUserByID", ctx, uint(1)).Return(existingUser, nil).Once()
	mockRepo.On("UpdateUser", ctx, mock.AnythingOfType("*models.User")).Return(nil).Once()

	user, err := service.UpdateProfile(ctx, 1, services.UpdateProfileRequest{
		FullName: "New Name",
	})

	assert.NoError(t, err)
	assert.Equal(t, "New Name", user.FullName)
}

func TestAuthService_UpdateProfile_DuplicateUsername(t *testing.T) {
	mockRepo, service := setupAuthTest(t)
	ctx := context.Background()

	existingUser := &models.User{ID: 1, Username: "admin", FullName: "Admin"}
	otherUser := &models.User{ID: 2, Username: "taken"}

	mockRepo.On("GetUserByID", ctx, uint(1)).Return(existingUser, nil).Once()
	mockRepo.On("GetUserByUsername", ctx, "taken").Return(otherUser, nil).Once()

	user, err := service.UpdateProfile(ctx, 1, services.UpdateProfileRequest{
		Username: "taken",
	})

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "sudah digunakan")
}

// --- ChangePassword ---

func TestAuthService_ChangePassword_Success(t *testing.T) {
	mockRepo, service := setupAuthTest(t)
	ctx := context.Background()

	hashedPw, _ := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.DefaultCost)
	existingUser := &models.User{ID: 1, Password: string(hashedPw)}

	mockRepo.On("GetUserByID", ctx, uint(1)).Return(existingUser, nil).Once()
	mockRepo.On("UpdateUser", ctx, mock.AnythingOfType("*models.User")).Return(nil).Once()

	err := service.ChangePassword(ctx, 1, services.ChangePasswordRequest{
		CurrentPassword: "oldpassword",
		NewPassword:     "newpassword123",
	})

	assert.NoError(t, err)
}

func TestAuthService_ChangePassword_WrongCurrentPassword(t *testing.T) {
	mockRepo, service := setupAuthTest(t)
	ctx := context.Background()

	hashedPw, _ := bcrypt.GenerateFromPassword([]byte("realpassword"), bcrypt.DefaultCost)
	existingUser := &models.User{ID: 1, Password: string(hashedPw)}

	mockRepo.On("GetUserByID", ctx, uint(1)).Return(existingUser, nil).Once()

	err := service.ChangePassword(ctx, 1, services.ChangePasswordRequest{
		CurrentPassword: "wrongpassword",
		NewPassword:     "newpassword123",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password saat ini salah")
}
