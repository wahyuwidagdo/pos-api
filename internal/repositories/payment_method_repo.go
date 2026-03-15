package repositories

import (
	"context"
	"pos-api/internal/models"

	"gorm.io/gorm"
)

type PaymentMethodRepository interface {
	Create(ctx context.Context, pm *models.PaymentMethod) error
	Update(ctx context.Context, pm *models.PaymentMethod) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*models.PaymentMethod, error)
	GetAll(ctx context.Context) ([]models.PaymentMethod, error)
	GetActive(ctx context.Context) ([]models.PaymentMethod, error)
}

type paymentMethodRepository struct {
	DB *gorm.DB
}

func NewPaymentMethodRepository(db *gorm.DB) PaymentMethodRepository {
	return &paymentMethodRepository{DB: db}
}

func (r *paymentMethodRepository) Create(ctx context.Context, pm *models.PaymentMethod) error {
	return r.DB.WithContext(ctx).Create(pm).Error
}

func (r *paymentMethodRepository) Update(ctx context.Context, pm *models.PaymentMethod) error {
	return r.DB.WithContext(ctx).Save(pm).Error
}

func (r *paymentMethodRepository) Delete(ctx context.Context, id uint) error {
	return r.DB.WithContext(ctx).Delete(&models.PaymentMethod{}, id).Error
}

func (r *paymentMethodRepository) GetByID(ctx context.Context, id uint) (*models.PaymentMethod, error) {
	var pm models.PaymentMethod
	err := r.DB.WithContext(ctx).First(&pm, id).Error
	return &pm, err
}

func (r *paymentMethodRepository) GetAll(ctx context.Context) ([]models.PaymentMethod, error) {
	var methods []models.PaymentMethod
	err := r.DB.WithContext(ctx).Order("sort_order ASC, name ASC").Find(&methods).Error
	return methods, err
}

func (r *paymentMethodRepository) GetActive(ctx context.Context) ([]models.PaymentMethod, error) {
	var methods []models.PaymentMethod
	err := r.DB.WithContext(ctx).Where("is_active = ?", true).Order("sort_order ASC, name ASC").Find(&methods).Error
	return methods, err
}
