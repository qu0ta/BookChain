package users

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type UserService interface {
	GetById(ctx context.Context, id string) (*User, error)
}

type Handler struct {
	service UserService
}

func NewHandler(service UserService) *Handler {
	return &Handler{service: service}
}
func matchServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidUserID):
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
	case errors.Is(err, ErrUserNotFound):
		http.Error(w, "User not found", http.StatusNotFound)
	default:
		http.Error(w, "Unknown error", http.StatusInternalServerError)
	}
}
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/users/{id}", h.GetByID)
}
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	user, err := h.service.GetById(r.Context(), id)
	if err != nil {
		matchServiceError(w, err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Printf("error encoding user: %v", err)
	}
}
