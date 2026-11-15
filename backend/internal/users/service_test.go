package users

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type mockUserRepository struct {
	user *User
	err  error
}

func (m *mockUserRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*User, error) {
	return m.user, m.err
}

func (m *mockUserRepository) GetByTelegramID(
	ctx context.Context,
	telegramID int64,
) (*User, error) {
	return m.user, m.err
}

func (m *mockUserRepository) Create(
	ctx context.Context,
	user *User,
) error {
	if m.err != nil {
		return m.err
	}

	user.ID = uuid.New()
	m.user = user
	return nil
}

func TestService_GetByID(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name    string
		id      string
		repo    *mockUserRepository
		wantErr error
	}{
		{
			name: "user found",
			id:   id.String(),
			repo: &mockUserRepository{
				user: &User{ID: id, FirstName: "Alice"},
			},
		},
		{
			name:    "invalid UUID",
			id:      "invalid-id",
			repo:    &mockUserRepository{},
			wantErr: ErrInvalidUserID,
		},
		{
			name: "user not found",
			id:   id.String(),
			repo: &mockUserRepository{
				err: ErrUserNotFound,
			},
			wantErr: ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(tt.repo)

			user, err := service.GetById(context.Background(), tt.id)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if user == nil || user.ID != id {
				t.Fatalf("unexpected user: %+v", user)
			}
		})
	}
}

func TestService_Create(t *testing.T) {
	tests := []struct {
		name       string
		telegramID int64
		firstName  string
		wantErr    error
	}{
		{
			name:       "success",
			telegramID: 123456789,
			firstName:  "Alice",
		},
		{
			name:       "invalid telegram ID",
			telegramID: -1,
			firstName:  "Alice",
			wantErr:    ErrInvalidTelegramID,
		},
		{
			name:       "empty first name",
			telegramID: 123456789,
			firstName:  "",
			wantErr:    ErrInvalidFirstName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepository{}
			service := NewService(repo)

			user, err := service.Create(
				context.Background(),
				tt.telegramID,
				nil,
				tt.firstName,
				nil,
			)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if user == nil || user.ID == uuid.Nil {
				t.Fatalf("expected created user, got %+v", user)
			}
		})
	}
}
