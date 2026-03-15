package services

import (
	"context"
	"errors"
	"fmt"
	"pos-api/internal/models"
	"pos-api/internal/repositories"
)

type CreatePaymentMethodRequest struct {
	Name      string `json:"name" validate:"required,min=1"`
	IsCash    bool   `json:"is_cash"`
	IsActive  bool   `json:"is_active"`
	SortOrder int    `json:"sort_order"`
}

type UpdatePaymentMethodRequest struct {
	Name      string `json:"name"`
	IsCash    *bool  `json:"is_cash"`
	IsActive  *bool  `json:"is_active"`
	SortOrder *int   `json:"sort_order"`
}

type PaymentMethodService interface {
	Create(ctx context.Context, req CreatePaymentMethodRequest) (*models.PaymentMethod, error)
	Update(ctx context.Context, id uint, req UpdatePaymentMethodRequest) (*models.PaymentMethod, error)
	Delete(ctx context.Context, id uint) error
	GetAll(ctx context.Context) ([]models.PaymentMethod, error)
	GetActive(ctx context.Context) ([]models.PaymentMethod, error)
}

type paymentMethodService struct {
	repo repositories.PaymentMethodRepository
}

func NewPaymentMethodService(repo repositories.PaymentMethodRepository) PaymentMethodService {
	return &paymentMethodService{repo: repo}
}

func (s *paymentMethodService) Create(ctx context.Context, req CreatePaymentMethodRequest) (*models.PaymentMethod, error) {
	if req.Name == "" {
		return nil, errors.New("name is required")
	}

	pm := &models.PaymentMethod{
		Name:      req.Name,
		IsCash:    req.IsCash,
		IsActive:  req.IsActive,
		SortOrder: req.SortOrder,
	}

	if err := s.repo.Create(ctx, pm); err != nil {
		return nil, fmt.Errorf("failed to create payment method: %w", err)
	}

	return pm, nil
}

func (s *paymentMethodService) Update(ctx context.Context, id uint, req UpdatePaymentMethodRequest) (*models.PaymentMethod, error) {
	pm, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("payment method with ID %d not found", id)
	}

	if req.Name != "" {
		pm.Name = req.Name
	}
	if req.IsCash != nil {
		pm.IsCash = *req.IsCash
	}
	if req.IsActive != nil {
		pm.IsActive = *req.IsActive
	}
	if req.SortOrder != nil {
		pm.SortOrder = *req.SortOrder
	}

	if err := s.repo.Update(ctx, pm); err != nil {
		return nil, fmt.Errorf("failed to update payment method: %w", err)
	}

	return pm, nil
}

func (s *paymentMethodService) Delete(ctx context.Context, id uint) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("payment method with ID %d not found", id)
	}
	return s.repo.Delete(ctx, id)
}

func (s *paymentMethodService) GetAll(ctx context.Context) ([]models.PaymentMethod, error) {
	return s.repo.GetAll(ctx)
}

func (s *paymentMethodService) GetActive(ctx context.Context) ([]models.PaymentMethod, error) {
	return s.repo.GetActive(ctx)
}
