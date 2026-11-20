package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrInvalidToken    = errors.New("invalid session token")
)

const (
	sessionPrefix = "session:"
	sessionTTL    = 24 * time.Hour
	tokenSize     = 32
)

type SessionStore struct {
	redis *redis.Client
}

func NewSessionStore(client *redis.Client) *SessionStore {
	return &SessionStore{redis: client}
}
func sessionKey(token string) string {
	hash := sha256.Sum256([]byte(token))
	return sessionPrefix + hex.EncodeToString(hash[:])
}
func validateToken(token string) error {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(raw) != tokenSize {
		return ErrInvalidToken
	}

	return nil
}

func (s *SessionStore) Create(ctx context.Context, userID uuid.UUID) (string, error) {
	if userID == uuid.Nil {
		return "", errors.New("user ID cannot be empty")
	}

	randomBytes := make([]byte, tokenSize)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}

	token := base64.RawURLEncoding.EncodeToString(randomBytes)
	key := sessionKey(token)
	err := s.redis.Set(ctx, key, userID.String(), sessionTTL).Err()
	if err != nil {
		return "", fmt.Errorf("save session: %w", err)
	}

	return token, nil
}

func (s *SessionStore) Get(ctx context.Context, token string) (uuid.UUID, error) {
	if err := validateToken(token); err != nil {
		return uuid.Nil, err
	}

	value, err := s.redis.Get(ctx, sessionKey(token)).Result()
	if errors.Is(err, redis.Nil) {
		return uuid.Nil, ErrSessionNotFound
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("get session: %w", err)
	}

	userID, err := uuid.Parse(value)
	if err != nil || userID == uuid.Nil {
		return uuid.Nil, errors.New("invalid session user ID")
	}

	return userID, nil
}

func (s *SessionStore) Revoke(
	ctx context.Context,
	token string,
) error {
	if err := validateToken(token); err != nil {
		return err
	}

	err := s.redis.Del(ctx, sessionKey(token)).Err()
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}

	return nil
}
