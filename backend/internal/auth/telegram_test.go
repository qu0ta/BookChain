package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

func signedInitData(token string, user string, authDate time.Time) string {
	params := url.Values{}

	params.Set("auth_date", strconv.FormatInt(authDate.Unix(), 10))
	params.Set("user", user)

	dataCheckString := "auth_date=" + params.Get("auth_date") +
		"\nuser=" + params.Get("user")

	secret := hmac.New(sha256.New, []byte("WebAppData"))
	secret.Write([]byte(token))

	mac := hmac.New(sha256.New, secret.Sum(nil))
	mac.Write([]byte(dataCheckString))

	params.Set("hash", hex.EncodeToString(mac.Sum(nil)))

	return params.Encode()
}

func TestTelegramAuth_Validate(t *testing.T) {
	const token = "123456:TEST_BOT_TOKEN"
	const userJSON = `{"id":123456789,"first_name":"Alice"}`

	validator := NewTelegramAuth(token)

	t.Run("valid initData", func(t *testing.T) {
		data := signedInitData(token, userJSON, time.Now())

		user, err := validator.Validate(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if user.ID != 123456789 || user.FirstName != "Alice" {
			t.Fatalf("unexpected user: %+v", user)
		}
	})

	t.Run("tampered user", func(t *testing.T) {
		data := signedInitData(token, userJSON, time.Now())

		params, err := url.ParseQuery(data)
		if err != nil {
			t.Fatal(err)
		}

		params.Set("user", `{"id":987654321,"first_name":"Bob"}`)

		_, err = validator.Validate(params.Encode())

		if !errors.Is(err, ErrInvalidInitData) {
			t.Fatalf("expected ErrInvalidInitData, got %v", err)
		}
	})

	t.Run("expired initData", func(t *testing.T) {
		data := signedInitData(
			token,
			userJSON,
			time.Now().Add(-48*time.Hour),
		)

		_, err := validator.Validate(data)

		if !errors.Is(err, ErrExpiredInitData) {
			t.Fatalf("expected ErrExpiredInitData, got %v", err)
		}
	})

	t.Run("invalid hash", func(t *testing.T) {
		data := signedInitData(token, userJSON, time.Now())

		params, err := url.ParseQuery(data)
		if err != nil {
			t.Fatal(err)
		}

		params.Set("hash", strings.Repeat("0", 64))

		_, err = validator.Validate(params.Encode())

		if !errors.Is(err, ErrInvalidInitData) {
			t.Fatalf("expected ErrInvalidInitData, got %v", err)
		}
	})

	t.Run("missing user", func(t *testing.T) {
		data := signedInitData(token, "", time.Now())

		_, err := validator.Validate(data)

		if !errors.Is(err, ErrInvalidInitData) {
			t.Fatalf("expected ErrInvalidInitData, got %v", err)
		}
	})
}
