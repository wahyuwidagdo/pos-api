package repositories

import (
	"pos-api/internal/models"

	"gorm.io/gorm"
)

type StoreSettingRepository interface {
	GetSettings() (*models.StoreSetting, error)
	UpsertSettings(settings *models.StoreSetting) (*models.StoreSetting, error)
}

type storeSettingRepository struct {
	db *gorm.DB
}

func NewStoreSettingRepository(db *gorm.DB) StoreSettingRepository {
	return &storeSettingRepository{db: db}
}

func (r *storeSettingRepository) GetSettings() (*models.StoreSetting, error) {
	var settings models.StoreSetting
	err := r.db.First(&settings).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Return default settings
			return &models.StoreSetting{
				StoreName:  "My Store",
				Address:    "",
				Phone:      "",
				FooterText: "Thank you for your purchase!",
			}, nil
		}
		return nil, err
	}
	return &settings, nil
}

func (r *storeSettingRepository) UpsertSettings(settings *models.StoreSetting) (*models.StoreSetting, error) {
	var existing models.StoreSetting
	err := r.db.First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create
		if err := r.db.Create(settings).Error; err != nil {
			return nil, err
		}
		return settings, nil
	} else if err != nil {
		return nil, err
	}

	// Update existing
	existing.StoreName = settings.StoreName
	existing.Address = settings.Address
	existing.Phone = settings.Phone
	existing.FooterText = settings.FooterText

	if err := r.db.Save(&existing).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}
