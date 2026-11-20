package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidInitData = errors.New("invalid telegram init data")
	ErrExpiredInitData = errors.New("expired telegram init data")
)

type TelegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

type TelegramAuth struct {
	botToken string
	maxAge   time.Duration
}

func NewTelegramAuth(botToken string) *TelegramAuth {
	return &TelegramAuth{
		botToken: botToken,
		maxAge:   24 * time.Hour,
	}
}

func (a *TelegramAuth) Validate(initData string) (*TelegramUser, error) {
	if a.botToken == "" || initData == "" {
		return nil, ErrInvalidInitData
	}

	params, err := url.ParseQuery(initData)
	if err != nil {
		return nil, ErrInvalidInitData
	}

	// Reject duplicate parameters.
	for _, values := range params {
		if len(values) != 1 {
			return nil, ErrInvalidInitData
		}
	}

	receivedHash, err := hex.DecodeString(params.Get("hash"))
	if err != nil || len(receivedHash) != sha256.Size {
		return nil, ErrInvalidInitData
	}

	// Build the data-check-string.
	var pairs []string

	for key, values := range params {
		if key == "hash" || key == "signature" {
			continue
		}

		pairs = append(pairs, key+"="+values[0])
	}

	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	secret := hmac.New(sha256.New, []byte("WebAppData"))
	secret.Write([]byte(a.botToken))
	secretKey := secret.Sum(nil)

	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(dataCheckString))

	if !hmac.Equal(receivedHash, mac.Sum(nil)) {
		return nil, ErrInvalidInitData
	}

	// Check authorization time.
	authDate, err := strconv.ParseInt(
		params.Get("auth_date"), 10, 64,
	)
	if err != nil {
		return nil, ErrInvalidInitData
	}

	now := time.Now()
	createdAt := time.Unix(authDate, 0)

	if createdAt.After(now.Add(time.Minute)) {
		return nil, ErrInvalidInitData
	}

	if now.Sub(createdAt) > a.maxAge {
		return nil, ErrExpiredInitData
	}

	// Parse verified user.
	rawUser := params.Get("user")
	if rawUser == "" {
		return nil, ErrInvalidInitData
	}

	var user TelegramUser

	if err := json.Unmarshal([]byte(rawUser), &user); err != nil {
		return nil, ErrInvalidInitData
	}

	if user.ID <= 0 || strings.TrimSpace(user.FirstName) == "" {
		return nil, ErrInvalidInitData
	}

	return &user, nil
}
