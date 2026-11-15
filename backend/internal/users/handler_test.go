package users

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

type mockUserService struct {
	user *User
	err  error
}

func (m *mockUserService) GetByID(
	ctx context.Context,
	id string,
) (*User, error) {
	return m.user, m.err
}

func TestHandler_GetByID(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name       string
		service    *mockUserService
		wantStatus int
	}{
		{
			name: "success",
			service: &mockUserService{
				user: &User{
					ID:        id,
					FirstName: "Alice",
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "invalid id",
			service: &mockUserService{
				err: ErrInvalidUserID,
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "user not found",
			service: &mockUserService{
				err: ErrUserNotFound,
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "internal error",
			service: &mockUserService{
				err: errors.New("database unavailable"),
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(tt.service)

			mux := http.NewServeMux()
			handler.RegisterRoutes(mux)

			req := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/users/"+id.String(),
				nil,
			)

			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf(
					"expected status %d, got %d",
					tt.wantStatus,
					rec.Code,
				)
			}

			if tt.wantStatus == http.StatusOK {
				var response map[string]any

				if err := json.Unmarshal(
					rec.Body.Bytes(),
					&response,
				); err != nil {
					t.Fatal(err)
				}

				if response["first_name"] != "Alice" {
					t.Errorf("unexpected response: %v", response)
				}

				if _, exists := response["telegram_id"]; exists {
					t.Error("telegram_id must not be public")
				}
			}
		})
	}
}
