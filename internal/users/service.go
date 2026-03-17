package users

import (
	"time"
)

type Service interface {
	GetMyProfile(userID string) (*UserProfileResponse, error)
	UpdateFCMToken(userID string, token string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo}
}

type UserProfileResponse struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Email             string    `json:"email"`
	BirthDate         time.Time `json:"birth_date"`
	ProfilePictureURL string    `json:"profile_picture_url"`
	ZodiacSign        string    `json:"zodiac_sign"`
	HasRelationship   bool      `json:"has_relationship"`
	PartnerID         *string   `json:"partner_id,omitempty"`
}

func (s *service) GetMyProfile(userID string) (*UserProfileResponse, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	relationship, err := s.repo.GetRelationshipByUserID(userID)
	if err != nil {
		return nil, err
	}

	res := &UserProfileResponse{
		ID:                user.ID,
		Name:              user.Name,
		Email:             user.Email,
		BirthDate:         user.BirthDate,
		ProfilePictureURL: user.ProfilePictureURL,
		ZodiacSign:        user.ZodiacSign,
		HasRelationship:   relationship != nil,
	}

	if relationship != nil {
		if relationship.User1ID == userID {
			res.PartnerID = &relationship.User2ID
		} else {
			res.PartnerID = &relationship.User1ID
		}
	}

	return res, nil
}

func (s *service) UpdateFCMToken(userID string, token string) error {
	return s.repo.UpdateFCMToken(userID, token)
}
