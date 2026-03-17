package relationships

import (
	"errors"

	"github.com/jlhal/parejas/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	CreateRequest(req *models.RelationshipRequest) error
	GetRequest(id string) (*models.RelationshipRequest, error)
	UpdateRequestStatus(id string, status string) error
	CreateRelationship(rel *models.Relationship) error
	GetRelationshipByUserID(userID string) (*models.Relationship, error)
	UpdateRelationship(rel *models.Relationship) error
	ListPendingRequests(receiverID string) ([]models.RelationshipRequest, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) CreateRequest(req *models.RelationshipRequest) error {
	return r.db.Create(req).Error
}

func (r *repository) GetRequest(id string) (*models.RelationshipRequest, error) {
	var req models.RelationshipRequest
	err := r.db.Where("id = ?", id).First(&req).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("request not found")
		}
		return nil, err
	}
	return &req, nil
}

func (r *repository) UpdateRequestStatus(id string, status string) error {
	return r.db.Model(&models.RelationshipRequest{}).Where("id = ?", id).Update("status", status).Error
}

func (r *repository) CreateRelationship(rel *models.Relationship) error {
	return r.db.Create(rel).Error
}

func (r *repository) GetRelationshipByUserID(userID string) (*models.Relationship, error) {
	var rel models.Relationship
	err := r.db.Preload("User1").Preload("User2").
		Where("user1_id = ? OR user2_id = ?", userID, userID).
		First(&rel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &rel, nil
}

func (r *repository) UpdateRelationship(rel *models.Relationship) error {
	// Use a blank model pointer with a WHERE clause to prevent GORM from
	// including association fields (user1_id, user2_id) in the UPDATE statement,
	// which would cause FK constraint errors.
	return r.db.Model(&models.Relationship{}).
		Where("id = ?", rel.ID).
		Updates(map[string]interface{}{
			"start_date": rel.StartDate,
			"how_we_met": rel.HowWeMet,
		}).Error
}

func (r *repository) ListPendingRequests(receiverID string) ([]models.RelationshipRequest, error) {
	var requests []models.RelationshipRequest
	err := r.db.Preload("Sender").Where("receiver_id = ? AND status = ?", receiverID, "pending").Find(&requests).Error
	return requests, err
}
