package chat

import (
	"github.com/jlhal/parejas/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	SaveMessage(msg *models.Message) error
	GetMessages(relationshipID string, limit int, beforeTime string) ([]models.Message, error)
	UpdateMessageStatus(msgID string, status string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) SaveMessage(msg *models.Message) error {
	// MariaDB validates JSON column content — empty string is not valid JSON
	if msg.Metadata == "" {
		msg.Metadata = "{}"
	}
	return r.db.Create(msg).Error
}

func (r *repository) GetMessages(relationshipID string, limit int, beforeTime string) ([]models.Message, error) {
	var messages []models.Message
	query := r.db.Where("relationship_id = ?", relationshipID)
	
	if beforeTime != "" {
		query = query.Where("created_at < ?", beforeTime)
	}

	err := query.Order("created_at DESC").Limit(limit).Find(&messages).Error
	return messages, err
}

func (r *repository) UpdateMessageStatus(msgID string, status string) error {
	return r.db.Model(&models.Message{}).Where("id = ?", msgID).Update("status", status).Error
}
