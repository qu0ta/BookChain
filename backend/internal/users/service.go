package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrInvalidUserID     = errors.New("invalid user id")
	ErrInvalidTelegramID = errors.New("invalid telegram id")
	ErrInvalidFirstName  = errors.New("invalid first name")
)

type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByTelegramID(ctx context.Context, telegramID int64) (*User, error)
	Create(ctx context.Context, user *User) error
}

type Service struct {
	repo UserRepository
}

func NewService(repo UserRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetById(ctx context.Context, id string) (*User, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, ErrInvalidUserID
	}
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func (s *Service) GetByTelegramID(ctx context.Context, telegramID int64) (*User, error) {
	if telegramID <= 0 {
		return nil, ErrInvalidTelegramID
	}
	user, err := s.repo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, fmt.Errorf("get user by telegram id: %w", err)
	}
	return user, nil
}
func (s *Service) Create(ctx context.Context, telegramID int64, username *string, firstName string, lastName *string) (*User, error) {
	if telegramID <= 0 {
		return nil, ErrInvalidTelegramID
	}

	if firstName == "" {
		return nil, ErrInvalidFirstName
	}
	user := &User{
		TelegramID: telegramID,
		Username:   username,
		FirstName:  firstName,
		LastName:   lastName,
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}
