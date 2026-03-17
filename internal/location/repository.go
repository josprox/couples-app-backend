package location

import (
	"github.com/jlhal/parejas/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	SaveLocation(loc *models.Location) error
	GetLastLocation(userID string) (*models.Location, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) SaveLocation(loc *models.Location) error {
	return r.db.Create(loc).Error
}

func (r *repository) GetLastLocation(userID string) (*models.Location, error) {
	var loc models.Location
	err := r.db.Where("user_id = ?", userID).Order("timestamp desc").First(&loc).Error
	if err != nil {
		return nil, err
	}
	return &loc, nil
}
