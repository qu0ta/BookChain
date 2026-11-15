package users

import (
	"time"

	"github.com/google/uuid"
)

type PublicUser struct {
	ID        uuid.UUID `json:"id"`
	Username  *string   `json:"username"`
	FirstName string    `json:"first_name"`
	LastName  *string   `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
}

func ToPublicUser(user *User) PublicUser {
	return PublicUser{
		ID:        user.ID,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
	}
}
