package auth

import (
	"errors"

	"github.com/jlhal/parejas/config"
	"github.com/jlhal/parejas/internal/models"
	"github.com/jlhal/parejas/pkg/utils"
)

type Service interface {
	Register(req RegisterRequest) (string, error)
	Login(req LoginRequest) (string, error)
}

type service struct {
	repo Repository
	cfg  *config.Config
}

func NewService(repo Repository, cfg *config.Config) Service {
	return &service{repo, cfg}
}

// DTOs
type RegisterRequest struct {
	Name      string    `json:"name" binding:"required"`
	Email     string    `json:"email" binding:"required,email"`
	Password  string    `json:"password" binding:"required,min=6"`
	BirthDate string    `json:"birth_date" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (s *service) Register(req RegisterRequest) (string, error) {
	// Check if exists
	_, err := s.repo.GetUserByEmail(req.Email)
	if err == nil {
		return "", errors.New("email already registered")
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return "", err
	}

	parsedDate, err := utils.ParseFlexibleDate(req.BirthDate)
	if err != nil {
		return "", err
	}

	zodiac := utils.CalculateZodiac(parsedDate)

	user := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hash,
		BirthDate:    parsedDate,
		ZodiacSign:   zodiac,
	}

	err = s.repo.CreateUser(user)
	if err != nil {
		return "", err
	}

	// Generate Token
	token, err := utils.GenerateToken(user.ID, s.cfg)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *service) Login(req LoginRequest) (string, error) {
	user, err := s.repo.GetUserByEmail(req.Email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return "", errors.New("invalid credentials")
	}

	token, err := utils.GenerateToken(user.ID, s.cfg)
	if err != nil {
		return "", err
	}

	return token, nil
}
