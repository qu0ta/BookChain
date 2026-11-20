package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/qu0ta/BookChain/internal/users"
)

type mockLoginService struct {
	user *users.User
	err  error
}

func (m *mockLoginService) Authenticate(
	ctx context.Context,
	initData string,
) (*users.User, error) {
	return m.user, m.err
}

func TestHandler_Login(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		service    *mockLoginService
		wantStatus int
	}{
		{
			name: "successful login",
			body: `{"init_data":"signed-data"}`,
			service: &mockLoginService{
				user: &users.User{
					ID:         uuid.New(),
					TelegramID: 123456789,
					FirstName:  "Alice",
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing init_data",
			body:       `{}`,
			service:    &mockLoginService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid signature",
			body: `{"init_data":"invalid"}`,
			service: &mockLoginService{
				err: ErrInvalidInitData,
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "expired initData",
			body: `{"init_data":"expired"}`,
			service: &mockLoginService{
				err: ErrExpiredInitData,
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "database error",
			body: `{"init_data":"signed-data"}`,
			service: &mockLoginService{
				err: errors.New("database unavailable"),
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "invalid JSON",
			body:       `{"init_data":`,
			service:    &mockLoginService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "multiple JSON values",
			body:       `{"init_data":"a"} {"init_data":"b"}`,
			service:    &mockLoginService{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(tt.service)

			mux := http.NewServeMux()
			handler.RegisterRoutes(mux)

			req := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/auth/telegram",
				bytes.NewBufferString(tt.body),
			)

			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf(
					"expected status %d, got %d; body: %s",
					tt.wantStatus,
					rec.Code,
					rec.Body.String(),
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

				userData, ok := response["user"].(map[string]any)
				if !ok {
					t.Fatalf("missing user in response: %v", response)
				}

				if userData["first_name"] != "Alice" {
					t.Fatalf("unexpected user: %v", userData)
				}

				if _, exists := userData["telegram_id"]; exists {
					t.Fatal("telegram_id must not be public")
				}
			}
		})
	}
}
