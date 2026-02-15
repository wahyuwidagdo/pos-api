package services

import (
	"pos-api/internal/models"
	"pos-api/internal/repositories"
)

type StoreSettingService interface {
	GetSettings() (*models.StoreSetting, error)
	UpdateSettings(settings *models.StoreSetting) (*models.StoreSetting, error)
}

type storeSettingService struct {
	repo repositories.StoreSettingRepository
}

func NewStoreSettingService(repo repositories.StoreSettingRepository) StoreSettingService {
	return &storeSettingService{repo: repo}
}

func (s *storeSettingService) GetSettings() (*models.StoreSetting, error) {
	return s.repo.GetSettings()
}

func (s *storeSettingService) UpdateSettings(settings *models.StoreSetting) (*models.StoreSetting, error) {
	return s.repo.UpsertSettings(settings)
}
