package auth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/qu0ta/BookChain/internal/users"
)

type AuthService interface {
	Authenticate(ctx context.Context, initData string) (*users.User, error)
}
type Handler struct {
	service AuthService
}

const maxBytes = 16 << 10

func NewHandler(service AuthService) *Handler {
	return &Handler{service: service}
}

type LoginRequest struct {
	InitData string `json:"init_data"`
}

type LoginResponse struct {
	User users.PublicUser `json:"user"`
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/telegram", h.Login)
}

func matchAuthServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidInitData), errors.Is(err, ErrExpiredInitData):
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	default:
		log.Printf("telegram authentication failed: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)

	}
	return
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var req LoginRequest

	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.InitData == "" {
		http.Error(w, "init_data is required", http.StatusBadRequest)
		return
	}

	user, err := h.service.Authenticate(r.Context(), req.InitData)
	if err != nil {
		matchAuthServiceError(w, err)
		return
	}

	if user == nil {
		log.Print("telegram authentication returned nil user")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	response := LoginResponse{User: users.ToPublicUser(user)}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("encode authentication response: %v", err)
	}
}
