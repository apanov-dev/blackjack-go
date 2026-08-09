package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	telegramInitDataHeader = "X-Telegram-Init-Data"
	defaultInitDataTTL     = 24 * time.Hour
)

var (
	errTelegramInitDataMissing = errors.New("telegram init data is missing")
	errTelegramBotTokenMissing = errors.New("telegram bot token is not configured")
	errTelegramInitDataInvalid = errors.New("telegram init data is invalid")
	errTelegramInitDataExpired = errors.New("telegram init data is expired")
	errTelegramUserMissing     = errors.New("telegram user is missing")
)

type TelegramAuth struct {
	InitData string       `json:"-"`
	User     TelegramUser `json:"user"`
}

type TelegramUser struct {
	ID              int64  `json:"id"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name,omitempty"`
	Username        string `json:"username,omitempty"`
	LanguageCode    string `json:"language_code,omitempty"`
	PhotoURL        string `json:"photo_url,omitempty"`
	IsPremium       bool   `json:"is_premium,omitempty"`
	AllowsWriteToPM bool   `json:"allows_write_to_pm,omitempty"`
}

func telegramAuthFromRequest(r *http.Request) (*TelegramAuth, error) {
	if telegramAuthDisabled() {
		return devTelegramAuth(r), nil
	}

	initData := telegramInitDataFromRequest(r)
	if initData == "" {
		return nil, errTelegramInitDataMissing
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		return nil, errTelegramBotTokenMissing
	}

	user, err := validateTelegramInitData(initData, botToken, telegramInitDataTTL())
	if err != nil {
		return nil, err
	}

	return &TelegramAuth{
		InitData: initData,
		User:     user,
	}, nil
}

func validateTelegramInitData(initData string, botToken string, ttl time.Duration) (TelegramUser, error) {
	values, err := url.ParseQuery(initData)
	if err != nil {
		return TelegramUser{}, errTelegramInitDataInvalid
	}

	receivedHash := values.Get("hash")
	if receivedHash == "" {
		return TelegramUser{}, errTelegramInitDataInvalid
	}

	authDate, err := strconv.ParseInt(values.Get("auth_date"), 10, 64)
	if err != nil {
		return TelegramUser{}, errTelegramInitDataInvalid
	}
	if ttl > 0 && time.Unix(authDate, 0).Add(ttl).Before(time.Now()) {
		return TelegramUser{}, errTelegramInitDataExpired
	}

	if !telegramInitDataHashValid(values, botToken, receivedHash) {
		return TelegramUser{}, errTelegramInitDataInvalid
	}

	userRaw := values.Get("user")
	if userRaw == "" {
		return TelegramUser{}, errTelegramUserMissing
	}

	var user TelegramUser
	if err := json.Unmarshal([]byte(userRaw), &user); err != nil {
		return TelegramUser{}, errTelegramInitDataInvalid
	}
	if user.ID == 0 {
		return TelegramUser{}, errTelegramUserMissing
	}

	return user, nil
}

func telegramInitDataHashValid(values url.Values, botToken string, receivedHash string) bool {
	dataCheckString := telegramDataCheckString(values)

	secretHash := hmac.New(sha256.New, []byte("WebAppData"))
	secretHash.Write([]byte(botToken))
	secretKey := secretHash.Sum(nil)

	dataHash := hmac.New(sha256.New, secretKey)
	dataHash.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(dataHash.Sum(nil))

	return hmac.Equal([]byte(expectedHash), []byte(receivedHash))
}

func telegramDataCheckString(values url.Values) string {
	pairs := make([]string, 0, len(values))

	for key, vals := range values {
		if key == "hash" {
			continue
		}
		if len(vals) == 0 {
			continue
		}
		pairs = append(pairs, key+"="+vals[0])
	}

	sort.Strings(pairs)
	return strings.Join(pairs, "\n")
}

func telegramInitDataFromRequest(r *http.Request) string {
	if initData := r.Header.Get(telegramInitDataHeader); initData != "" {
		return initData
	}

	authHeader := r.Header.Get("Authorization")
	for _, prefix := range []string{"tma ", "TMA ", "Telegram ", "telegram "} {
		if strings.HasPrefix(authHeader, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
		}
	}

	return ""
}

func telegramInitDataTTL() time.Duration {
	raw := os.Getenv("TELEGRAM_INIT_DATA_TTL")
	if raw == "" {
		return defaultInitDataTTL
	}

	seconds, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || seconds < 0 {
		return defaultInitDataTTL
	}

	return time.Duration(seconds) * time.Second
}

func telegramAuthDisabled() bool {
	return os.Getenv("TELEGRAM_AUTH_DISABLED") == "true"
}

func devTelegramAuth(r *http.Request) *TelegramAuth {
	userID, err := strconv.ParseInt(r.Header.Get("X-Dev-Telegram-User-ID"), 10, 64)
	if err != nil || userID == 0 {
		userID = 1
	}

	return &TelegramAuth{
		User: TelegramUser{
			ID:        userID,
			FirstName: "Dev",
			Username:  "dev_user",
		},
	}
}

func writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errTelegramBotTokenMissing):
		writeError(w, http.StatusServiceUnavailable, err.Error())
	case errors.Is(err, errTelegramInitDataMissing):
		writeError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, errTelegramInitDataExpired):
		writeError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, errTelegramUserMissing):
		writeError(w, http.StatusUnauthorized, err.Error())
	default:
		writeError(w, http.StatusUnauthorized, errTelegramInitDataInvalid.Error())
	}
}
