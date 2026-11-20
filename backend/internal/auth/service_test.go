package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/qu0ta/BookChain/internal/users"
)

type mockValidator struct {
	user *TelegramUser
	err  error
}

func (m *mockValidator) Validate(
	initData string,
) (*TelegramUser, error) {
	return m.user, m.err
}

type mockAuthUserRepo struct {
	user   *users.User
	err    error
	called bool
}

func (m *mockAuthUserRepo) UpsertTelegramUser(
	ctx context.Context,
	telegramID int64,
	username *string,
	firstName string,
	lastName *string,
) (*users.User, error) {
	m.called = true

	if m.err != nil {
		return nil, m.err
	}

	if m.user != nil {
		return m.user, nil
	}

	return &users.User{
		ID:         uuid.New(),
		TelegramID: telegramID,
		Username:   username,
		FirstName:  firstName,
		LastName:   lastName,
	}, nil
}

func TestService_Authenticate(t *testing.T) {
	t.Run("successful authentication", func(t *testing.T) {
		validator := &mockValidator{
			user: &TelegramUser{
				ID:        123456789,
				FirstName: "Alice",
				Username:  "alice",
			},
		}

		repo := &mockAuthUserRepo{}
		service := NewService(validator, repo)

		user, err := service.Authenticate(
			context.Background(),
			"signed-init-data",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !repo.called {
			t.Fatal("repository was not called")
		}

		if user.ID == uuid.Nil {
			t.Fatal("expected non-empty user UUID")
		}

		if user.TelegramID != 123456789 {
			t.Fatalf("unexpected telegram ID: %d", user.TelegramID)
		}

		if user.Username == nil || *user.Username != "alice" {
			t.Fatalf("unexpected username: %v", user.Username)
		}
	})

	t.Run("invalid initData", func(t *testing.T) {
		validator := &mockValidator{
			err: ErrInvalidInitData,
		}

		repo := &mockAuthUserRepo{}
		service := NewService(validator, repo)

		_, err := service.Authenticate(
			context.Background(),
			"invalid",
		)

		if !errors.Is(err, ErrInvalidInitData) {
			t.Fatalf("expected ErrInvalidInitData, got %v", err)
		}

		if repo.called {
			t.Fatal("repository must not be called")
		}
	})

	t.Run("database error", func(t *testing.T) {
		dbErr := errors.New("database unavailable")

		validator := &mockValidator{
			user: &TelegramUser{
				ID:        123456789,
				FirstName: "Alice",
			},
		}

		repo := &mockAuthUserRepo{
			err: dbErr,
		}

		service := NewService(validator, repo)

		_, err := service.Authenticate(
			context.Background(),
			"signed-init-data",
		)

		if !errors.Is(err, dbErr) {
			t.Fatalf("expected database error, got %v", err)
		}
	})
}
