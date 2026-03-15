package services

import (
	"context"
	"errors"
	"fmt"
	"pos-api/internal/models"
	"pos-api/internal/repositories"
	"time"
)

type CreateCashFlowRequest struct {
	Type   string  `json:"type" validate:"required,oneof=income expense"`
	Source string  `json:"source" validate:"required"`
	Amount float64 `json:"amount" validate:"required,gt=0"`
	Date   string  `json:"date" validate:"required"` // "2026-02-16"
	Notes  string  `json:"notes"`
}

type UpdateCashFlowRequest struct {
	Type   string  `json:"type" validate:"omitempty,oneof=income expense"`
	Source string  `json:"source"`
	Amount float64 `json:"amount" validate:"omitempty,gt=0"`
	Date   string  `json:"date"`
	Notes  string  `json:"notes"`
}

type CashFlowSummary struct {
	TotalCapital float64 `json:"total_capital"`
	TotalIncome  float64 `json:"total_income"`
	TotalExpense float64 `json:"total_expense"`
	NetProfit    float64 `json:"net_profit"`
}

type CashFlowService interface {
	Create(ctx context.Context, req CreateCashFlowRequest, userID uint) (*models.CashFlow, error)
	Update(ctx context.Context, id uint, req UpdateCashFlowRequest) (*models.CashFlow, error)
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*models.CashFlow, error)
	GetAll(ctx context.Context, page, pageSize int, cfType, source string, startDate, endDate *time.Time) ([]models.CashFlow, int64, error)
	GetSummary(ctx context.Context, startDate, endDate time.Time) (*CashFlowSummary, error)
}

type cashFlowService struct {
	repo repositories.CashFlowRepository
}

func NewCashFlowService(repo repositories.CashFlowRepository) CashFlowService {
	return &cashFlowService{repo: repo}
}

func (s *cashFlowService) Create(ctx context.Context, req CreateCashFlowRequest, userID uint) (*models.CashFlow, error) {
	if req.Type != "income" && req.Type != "expense" {
		return nil, errors.New("type must be 'income' or 'expense'")
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format, use YYYY-MM-DD: %w", err)
	}

	cf := &models.CashFlow{
		Type:   req.Type,
		Source: req.Source,
		Amount: req.Amount,
		Date:   date,
		Notes:  req.Notes,
		UserID: userID,
	}

	if err := s.repo.Create(ctx, cf); err != nil {
		return nil, fmt.Errorf("failed to create cash flow entry: %w", err)
	}

	return cf, nil
}

func (s *cashFlowService) Update(ctx context.Context, id uint, req UpdateCashFlowRequest) (*models.CashFlow, error) {
	cf, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("cash flow entry with ID %d not found", id)
	}

	if req.Type != "" {
		cf.Type = req.Type
	}
	if req.Source != "" {
		cf.Source = req.Source
	}
	if req.Amount > 0 {
		cf.Amount = req.Amount
	}
	if req.Date != "" {
		date, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %w", err)
		}
		cf.Date = date
	}
	if req.Notes != "" {
		cf.Notes = req.Notes
	}

	if err := s.repo.Update(ctx, cf); err != nil {
		return nil, fmt.Errorf("failed to update cash flow entry: %w", err)
	}

	return cf, nil
}

func (s *cashFlowService) Delete(ctx context.Context, id uint) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("cash flow entry with ID %d not found", id)
	}
	return s.repo.Delete(ctx, id)
}

func (s *cashFlowService) GetByID(ctx context.Context, id uint) (*models.CashFlow, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *cashFlowService) GetAll(ctx context.Context, page, pageSize int, cfType, source string, startDate, endDate *time.Time) ([]models.CashFlow, int64, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.GetAll(ctx, pageSize, offset, cfType, source, startDate, endDate)
}

func (s *cashFlowService) GetSummary(ctx context.Context, startDate, endDate time.Time) (*CashFlowSummary, error) {
	capital, income, expense, err := s.repo.GetSummary(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}
	return &CashFlowSummary{
		TotalCapital: capital,
		TotalIncome:  income,
		TotalExpense: expense,
		NetProfit:    income - expense,
	}, nil
}
