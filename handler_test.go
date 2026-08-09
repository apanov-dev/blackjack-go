package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestCardTranslateUsesAceAsOneWhenNeeded(t *testing.T) {
	cards := []Card{
		{Rank: "A", Suit: "Spades"},
		{Rank: "9", Suit: "Hearts"},
		{Rank: "5", Suit: "Clubs"},
	}

	if got := cardTranslate(cards); got != 15 {
		t.Fatalf("cardTranslate() = %d, want 15", got)
	}
}

func TestWinCheckReturnsTie(t *testing.T) {
	if got := winCheck(20, 20); got != StatusTie {
		t.Fatalf("winCheck() = %q, want %q", got, StatusTie)
	}
}

func TestCreateGameHandler(t *testing.T) {
	t.Setenv("TELEGRAM_AUTH_DISABLED", "true")
	resetGameStore()

	rr := serveTestRequest(http.MethodPost, "/games")
	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusCreated, rr.Body.String())
	}

	var response GameResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.ID == "" {
		t.Fatal("expected game id")
	}
	if len(response.PlayerCards) != 2 {
		t.Fatalf("player cards = %d, want 2", len(response.PlayerCards))
	}
	if len(response.DealerCards) != 2 {
		t.Fatalf("dealer cards = %d, want 2", len(response.DealerCards))
	}
	if response.CardsLeft != 48 {
		t.Fatalf("cards left = %d, want 48", response.CardsLeft)
	}
}

func TestStandHandlerUpdatesDealerCardsAndDeck(t *testing.T) {
	t.Setenv("TELEGRAM_AUTH_DISABLED", "true")
	resetGameStore()

	game := &Game{
		ID:             "test-game",
		TelegramUserID: 1,
		Deck: []Card{
			{Rank: "3", Suit: "Spades"},
			{Rank: "K", Suit: "Hearts"},
		},
		Player: Player{cards: []Card{
			{Rank: "2", Suit: "Hearts"},
			{Rank: "2", Suit: "Clubs"},
		}},
		Dealer: Dealer{cards: []Card{
			{Rank: "Q", Suit: "Hearts"},
			{Rank: "2", Suit: "Spades"},
		}},
		Status: StatusPlaying,
	}

	gameStore.Lock()
	gameStore.games[game.ID] = game
	gameStore.Unlock()

	rr := serveTestRequest(http.MethodPost, "/games/test-game/stand")
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var response GameResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Status != StatusDealerBust {
		t.Fatalf("status = %q, want %q", response.Status, StatusDealerBust)
	}
	if len(response.DealerCards) != 4 {
		t.Fatalf("dealer cards = %d, want 4", len(response.DealerCards))
	}
	if response.DealerScore != 25 {
		t.Fatalf("dealer score = %d, want 25", response.DealerScore)
	}
	if response.CardsLeft != 0 {
		t.Fatalf("cards left = %d, want 0", response.CardsLeft)
	}
}

func TestGameHandlerRejectsAnotherTelegramUser(t *testing.T) {
	t.Setenv("TELEGRAM_AUTH_DISABLED", "true")
	resetGameStore()

	game := &Game{
		ID:             "test-game",
		TelegramUserID: 100,
		Status:         StatusPlaying,
	}

	gameStore.Lock()
	gameStore.games[game.ID] = game
	gameStore.Unlock()

	rr := serveTestRequestWithUser(http.MethodGet, "/games/test-game", 200)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusForbidden, rr.Body.String())
	}
}

func TestGamesHandlerRequiresTelegramAuth(t *testing.T) {
	t.Setenv("TELEGRAM_AUTH_DISABLED", "")
	t.Setenv("TELEGRAM_BOT_TOKEN", "")
	resetGameStore()

	mux := http.NewServeMux()
	registerRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/games", nil)
	rr := httptest.NewRecorder()

	withCORS(mux).ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusUnauthorized, rr.Body.String())
	}
}

func TestTelegramMeHandler(t *testing.T) {
	t.Setenv("TELEGRAM_AUTH_DISABLED", "true")

	rr := serveTestRequestWithUser(http.MethodGet, "/telegram/me", 42)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var auth TelegramUser
	if err := json.NewDecoder(rr.Body).Decode(&auth); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if auth.ID != 42 {
		t.Fatalf("telegram user id = %d, want 42", auth.ID)
	}
}

func resetGameStore() {
	gameStore.Lock()
	defer gameStore.Unlock()

	gameStore.games = make(map[string]*Game)
}

func serveTestRequest(method string, path string) *httptest.ResponseRecorder {
	return serveTestRequestWithUser(method, path, 1)
}

func serveTestRequestWithUser(method string, path string, userID int64) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	registerRoutes(mux)

	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("X-Dev-Telegram-User-ID", strconv.FormatInt(userID, 10))
	rr := httptest.NewRecorder()

	withCORS(mux).ServeHTTP(rr, req)
	return rr
}
