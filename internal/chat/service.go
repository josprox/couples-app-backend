package chat

import (
	"errors"
	"log"

	"github.com/jlhal/parejas/internal/models"
	"github.com/jlhal/parejas/internal/relationships"
	"github.com/jlhal/parejas/pkg/notifications"
)

type Service interface {
	SaveMessage(msg models.Message) (*models.Message, error)
	GetMessages(userID string, limit int, beforeTime string) ([]models.Message, error)
	UpdateMessageStatus(msgID string, status string) error
}

type service struct {
	repo       Repository
	relService relationships.Service
}

func NewService(repo Repository, relService relationships.Service) Service {
	return &service{repo, relService}
}

func (s *service) SaveMessage(msg models.Message) (*models.Message, error) {
	// The WS layer already attached UserID and RelationshipID to msg
	err := s.repo.SaveMessage(&msg)
	if err != nil {
		return nil, err
	}

	// Trigger Push Notification if sender is not null (User message)
	if msg.SenderID != nil {
		go func() {
			token, err := s.relService.GetPartnerFCMToken(*msg.SenderID)
			if err == nil && token != "" {
				notifications.SendPushNotification(
					token,
					"Nuevo mensaje ❤️",
					msg.Content,
					map[string]string{"type": "chat", "relationship_id": msg.RelationshipID},
				)
			} else if err != nil {
				log.Printf("Error getting partner FCM token: %v", err)
			}
		}()
	}

	return &msg, nil
}

func (s *service) GetMessages(userID string, limit int, beforeTime string) ([]models.Message, error) {
	rel, err := s.relService.GetMyRelationship(userID)
	if err != nil {
		return nil, err
	}
	if rel == nil {
		return nil, errors.New("you are not in a relationship")
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	messages, err := s.repo.GetMessages(rel.ID, limit, beforeTime)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (s *service) UpdateMessageStatus(msgID string, status string) error {
	return s.repo.UpdateMessageStatus(msgID, status)
}
