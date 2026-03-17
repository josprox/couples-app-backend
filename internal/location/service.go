package location

import (
	"github.com/jlhal/parejas/internal/models"
)

type Service interface {
	SaveLocation(userID, relID string, lat, lng float64) error
	GetPartnerLocation(userID string, relID string) (*models.Location, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo}
}

func (s *service) SaveLocation(userID, relID string, lat, lng float64) error {
	loc := &models.Location{
		UserID:         userID,
		RelationshipID: relID,
		Latitude:       lat,
		Longitude:      lng,
	}
	return s.repo.SaveLocation(loc)
}

func (s *service) GetPartnerLocation(userID string, relID string) (*models.Location, error) {
	return s.repo.GetLastLocation(userID)
}
