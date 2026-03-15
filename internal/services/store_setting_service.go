package services

import (
	"context"
	"pos-api/internal/models"
	"pos-api/internal/repositories"
)

type StoreSettingService interface {
	GetSettings(ctx context.Context) (*models.StoreSetting, error)
	UpdateSettings(ctx context.Context, settings *models.StoreSetting) (*models.StoreSetting, error)
}

type storeSettingService struct {
	repo repositories.StoreSettingRepository
}

func NewStoreSettingService(repo repositories.StoreSettingRepository) StoreSettingService {
	return &storeSettingService{repo: repo}
}

func (s *storeSettingService) GetSettings(ctx context.Context) (*models.StoreSetting, error) {
	return s.repo.GetSettings(ctx)
}

func (s *storeSettingService) UpdateSettings(ctx context.Context, settings *models.StoreSetting) (*models.StoreSetting, error) {
	return s.repo.UpsertSettings(ctx, settings)
}
