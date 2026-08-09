package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func TestValidateTelegramInitData(t *testing.T) {
	token := "123456:test-token"
	initData := signedTelegramInitData(token, time.Now().Unix(), `{"id":42,"first_name":"Test","username":"test_user"}`)

	user, err := validateTelegramInitData(initData, token, time.Hour)
	if err != nil {
		t.Fatalf("validateTelegramInitData() error = %v", err)
	}
	if user.ID != 42 {
		t.Fatalf("user id = %d, want 42", user.ID)
	}
	if user.Username != "test_user" {
		t.Fatalf("username = %q, want test_user", user.Username)
	}
}

func TestValidateTelegramInitDataRejectsBadHash(t *testing.T) {
	token := "123456:test-token"
	initData := signedTelegramInitData(token, time.Now().Unix(), `{"id":42,"first_name":"Test"}`) + "bad"

	if _, err := validateTelegramInitData(initData, token, time.Hour); err == nil {
		t.Fatal("expected invalid init data error")
	}
}

func TestValidateTelegramInitDataRejectsExpiredData(t *testing.T) {
	token := "123456:test-token"
	authDate := time.Now().Add(-2 * time.Hour).Unix()
	initData := signedTelegramInitData(token, authDate, `{"id":42,"first_name":"Test"}`)

	if _, err := validateTelegramInitData(initData, token, time.Hour); err != errTelegramInitDataExpired {
		t.Fatalf("error = %v, want %v", err, errTelegramInitDataExpired)
	}
}

func signedTelegramInitData(token string, authDate int64, userJSON string) string {
	values := url.Values{}
	values.Set("auth_date", strconv.FormatInt(authDate, 10))
	values.Set("query_id", "test-query-id")
	values.Set("user", userJSON)

	dataCheckString := telegramDataCheckString(values)

	secretHash := hmac.New(sha256.New, []byte("WebAppData"))
	secretHash.Write([]byte(token))
	secretKey := secretHash.Sum(nil)

	dataHash := hmac.New(sha256.New, secretKey)
	dataHash.Write([]byte(dataCheckString))
	values.Set("hash", hex.EncodeToString(dataHash.Sum(nil)))

	return values.Encode()
}
