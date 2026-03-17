package relationships

import (
	"errors"

	"github.com/jlhal/parejas/internal/models"
	"github.com/jlhal/parejas/internal/users"
	"github.com/jlhal/parejas/pkg/utils"
)

type Service interface {
	SendRequest(senderID, receiverEmail string) error
	AcceptRequest(receiverID, requestID string) error
	GetMyRelationship(userID string) (*models.Relationship, error)
	UpdateWizard(userID string, req UpdateWizardRequest) error
	GetPendingRequests(userID string) ([]models.RelationshipRequest, error)
	GetPartnerFCMToken(userID string) (string, error)
}

type service struct {
	repo      Repository
	usersRepo users.Repository
}

func NewService(repo Repository, usersRepo users.Repository) Service {
	return &service{repo, usersRepo}
}

type SendRequestDTO struct {
	ReceiverEmail string `json:"receiver_email" binding:"required,email"`
}

type UpdateWizardRequest struct {
	StartDate string `json:"start_date" binding:"required"`
	HowWeMet  string `json:"how_we_met" binding:"required"`
}

func (s *service) SendRequest(senderID, receiverEmail string) error {
	// 1. Get sender relationship to check if already in one
	rel, err := s.repo.GetRelationshipByUserID(senderID)
	if err != nil {
		return err
	}
	if rel != nil {
		return errors.New("you are already in a relationship")
	}

	// 2. Find receiver by email (We need a way to get user by email from usersRepo)
	// Since UsersRepo doesn't have GetUserByEmail, we could add it, or use AuthRepo.
	// Let's assume we can add GetUserByEmail to UsersRepo or AuthRepo. 
	// For simplicity in Clean Architecture cross-domain calls, I will ask for it in UsersRepo interfaces soon.
	// Actually, I will just call a method that I am about to add to UsersRepo.
	receiver, err := s.usersRepo.GetUserByEmail(receiverEmail)
	if err != nil {
		return errors.New("receiver user not found")
	}

	if senderID == receiver.ID {
		return errors.New("cannot send request to yourself")
	}

	// 3. Create request
	req := &models.RelationshipRequest{
		SenderID:   senderID,
		ReceiverID: receiver.ID,
		Status:     "pending",
	}

	return s.repo.CreateRequest(req)
}

func (s *service) AcceptRequest(receiverID, requestID string) error {
	req, err := s.repo.GetRequest(requestID)
	if err != nil {
		return err
	}

	if req.ReceiverID != receiverID {
		return errors.New("unauthorized to accept this request")
	}

	if req.Status != "pending" {
		return errors.New("request is not pending")
	}

	// Update request status
	err = s.repo.UpdateRequestStatus(requestID, "accepted")
	if err != nil {
		return err
	}

	// Create real relationship
	rel := &models.Relationship{
		User1ID: req.SenderID,
		User2ID: req.ReceiverID,
	}

	return s.repo.CreateRelationship(rel)
}

func (s *service) GetMyRelationship(userID string) (*models.Relationship, error) {
	rel, err := s.repo.GetRelationshipByUserID(userID)
	if err != nil {
		return nil, err
	}
	if rel == nil {
		return nil, errors.New("no relationship found")
	}
	return rel, nil
}

func (s *service) UpdateWizard(userID string, req UpdateWizardRequest) error {
	rel, err := s.repo.GetRelationshipByUserID(userID)
	if err != nil {
		return err
	}
	if rel == nil {
		return errors.New("no relationship found")
	}

	parsedDate, err := utils.ParseFlexibleDate(req.StartDate)
	if err != nil {
		return err
	}

	rel.StartDate = parsedDate
	rel.HowWeMet = req.HowWeMet

	return s.repo.UpdateRelationship(rel)
}

func (s *service) GetPartnerFCMToken(userID string) (string, error) {
	rel, err := s.repo.GetRelationshipByUserID(userID)
	if err != nil {
		return "", err
	}
	if rel == nil {
		return "", errors.New("no relationship found")
	}

	partnerID := rel.User1ID
	if partnerID == userID {
		partnerID = rel.User2ID
	}

	partner, err := s.usersRepo.GetUserByID(partnerID)
	if err != nil {
		return "", err
	}

	return partner.FCMToken, nil
}

func (s *service) GetPendingRequests(userID string) ([]models.RelationshipRequest, error) {
	return s.repo.ListPendingRequests(userID)
}
