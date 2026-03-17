package users

import (
	"errors"

	"github.com/jlhal/parejas/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	GetUserByID(id string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetRelationshipByUserID(userID string) (*models.Relationship, error)
	UpdateFCMToken(userID string, token string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) GetUserByID(id string) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *repository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *repository) GetRelationshipByUserID(userID string) (*models.Relationship, error) {
	var relationship models.Relationship
	err := r.db.Preload("User1").Preload("User2").
		Where("user1_id = ? OR user2_id = ?", userID, userID).
		First(&relationship).Error
		
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No relationship yet
		}
		return nil, err
	}
	return &relationship, nil
}

func (r *repository) UpdateFCMToken(userID string, token string) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("fcm_token", token).Error
}
