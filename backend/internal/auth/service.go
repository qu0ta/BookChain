package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/qu0ta/BookChain/internal/users"
)

type UserRepository interface {
	UpsertTelegramUser(ctx context.Context, telegramID int64, username *string, firstName string, lastName *string) (*users.User, error)
}

type Service struct {
	validator InitDataValidator
	users     UserRepository
}

type InitDataValidator interface {
	Validate(initData string) (*TelegramUser, error)
}

func NewService(validator InitDataValidator, users UserRepository) *Service {
	return &Service{
		validator: validator,
		users:     users,
	}
}

func nullableString(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

func (s *Service) Authenticate(ctx context.Context, initData string) (*users.User, error) {
	telegramUser, err := s.validator.Validate(initData)
	if err != nil {
		return nil, fmt.Errorf("validate telegram initData: %w", err)
	}

	username := nullableString(telegramUser.Username)
	lastName := nullableString(telegramUser.LastName)

	user, err := s.users.UpsertTelegramUser(ctx, telegramUser.ID, username, telegramUser.FirstName, lastName)
	if err != nil {
		return nil, fmt.Errorf("upsert telegram user: %w", err)
	}

	return user, nil
}
