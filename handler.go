package main

import (
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Game struct {
	ID             string     `json:"id"`
	TelegramUserID int64      `json:"-"`
	Deck           []Card     `json:"-"`
	Player         Player     `json:"-"`
	Dealer         Dealer     `json:"-"`
	Status         GameStatus `json:"status"`
}

type GameResponse struct {
	ID          string     `json:"id"`
	Status      GameStatus `json:"status"`
	PlayerCards []Card     `json:"player_cards"`
	DealerCards []Card     `json:"dealer_cards"`
	PlayerScore int        `json:"player_score"`
	DealerScore int        `json:"dealer_score"`
	CardsLeft   int        `json:"cards_left"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

var gameStore = struct {
	sync.RWMutex
	games map[string]*Game
}{
	games: make(map[string]*Game),
}

func registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/deck", deckHandler)
	mux.HandleFunc("/telegram/me", telegramMeHandler)
	mux.HandleFunc("/games", gamesHandler)
	mux.HandleFunc("/games/", gameActionHandler)

	// Serve web files
	mux.Handle("/", http.FileServer(http.Dir("web")))
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Telegram-Init-Data")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func deckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, deckCreator())
}

func telegramMeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	auth, err := telegramAuthFromRequest(r)
	if err != nil {
		writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, auth.User)
}

func gamesHandler(w http.ResponseWriter, r *http.Request) {
	auth, err := telegramAuthFromRequest(r)
	if err != nil {
		writeAuthError(w, err)
		return
	}

	switch r.Method {
	case http.MethodPost:
		createGameHandler(w, auth)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func gameActionHandler(w http.ResponseWriter, r *http.Request) {
	auth, err := telegramAuthFromRequest(r)
	if err != nil {
		writeAuthError(w, err)
		return
	}

	gameID, action, err := parseGamePath(r.URL.Path)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	switch {
	case action == "" && r.Method == http.MethodGet:
		getGameHandler(w, gameID, auth)
	case action == "hit" && r.Method == http.MethodPost:
		hitGameHandler(w, gameID, auth)
	case action == "stand" && r.Method == http.MethodPost:
		standGameHandler(w, gameID, auth)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func createGameHandler(w http.ResponseWriter, auth *TelegramAuth) {
	player := Player{}
	dealer := Dealer{}
	deck := deckCreator()

	deck = FirstDistribution(&player, &dealer, deck)

	game := &Game{
		ID:             newGameID(),
		TelegramUserID: auth.User.ID,
		Deck:           deck,
		Player:         player,
		Dealer:         dealer,
		Status:         StatusPlaying,
	}

	playerScore := cardTranslate(game.Player.cards)
	dealerScore := cardTranslate(game.Dealer.cards)

	if playerScore == 21 && dealerScore == 21 {
		game.Status = StatusTie
	} else if playerScore == 21 {
		game.Status = StatusPlayerWin
	} else if dealerScore == 21 {
		game.Status = StatusDealerWin
	}

	gameStore.Lock()
	gameStore.games[game.ID] = game
	gameStore.Unlock()

	writeJSON(w, http.StatusCreated, buildGameResponse(game))
}

func getGameHandler(w http.ResponseWriter, gameID string, auth *TelegramAuth) {
	game, ok := findGame(gameID)
	if !ok {
		writeError(w, http.StatusNotFound, "game not found")
		return
	}
	if !canAccessGame(game, auth) {
		writeError(w, http.StatusForbidden, "game belongs to another telegram user")
		return
	}

	writeJSON(w, http.StatusOK, buildGameResponse(game))
}

func hitGameHandler(w http.ResponseWriter, gameID string, auth *TelegramAuth) {
	gameStore.Lock()
	defer gameStore.Unlock()

	game, ok := gameStore.games[gameID]
	if !ok {
		writeError(w, http.StatusNotFound, "game not found")
		return
	}
	if !canAccessGame(game, auth) {
		writeError(w, http.StatusForbidden, "game belongs to another telegram user")
		return
	}
	if game.Status != StatusPlaying {
		writeJSON(w, http.StatusOK, buildGameResponse(game))
		return
	}

	game.Deck, game.Player.cards = hit(game.Player.cards, game.Deck)

	playerScore := cardTranslate(game.Player.cards)
	if playerScore > 21 {
		game.Status = StatusPlayerBust
	}

	writeJSON(w, http.StatusOK, buildGameResponse(game))
}

func standGameHandler(w http.ResponseWriter, gameID string, auth *TelegramAuth) {
	gameStore.Lock()
	defer gameStore.Unlock()

	game, ok := gameStore.games[gameID]
	if !ok {
		writeError(w, http.StatusNotFound, "game not found")
		return
	}
	if !canAccessGame(game, auth) {
		writeError(w, http.StatusForbidden, "game belongs to another telegram user")
		return
	}
	if game.Status != StatusPlaying {
		writeJSON(w, http.StatusOK, buildGameResponse(game))
		return
	}

	game.Dealer.cards, game.Deck = dealerLogic(game.Dealer.cards, game.Deck)

	game.Status = winCheck(
		cardTranslate(game.Player.cards),
		cardTranslate(game.Dealer.cards),
	)

	writeJSON(w, http.StatusOK, buildGameResponse(game))
}

func canAccessGame(game *Game, auth *TelegramAuth) bool {
	return game.TelegramUserID == auth.User.ID
}

func buildGameResponse(game *Game) GameResponse {
	dealerCards := game.Dealer.cards
	dealerScore := cardTranslate(game.Dealer.cards)

	// If game is in progress, hide the second dealer card
	if game.Status == StatusPlaying && len(dealerCards) >= 2 {
		maskedCards := make([]Card, len(dealerCards))
		copy(maskedCards, dealerCards)
		// Mask the second card (the one dealt last during distribution)
		maskedCards[1] = Card{Rank: "?", Suit: "?"}
		dealerCards = maskedCards
		// Show score only for the first card
		dealerScore = cardTranslate(game.Dealer.cards[:1])
	}

	return GameResponse{
		ID:          game.ID,
		Status:      game.Status,
		PlayerCards: game.Player.cards,
		DealerCards: dealerCards,
		PlayerScore: cardTranslate(game.Player.cards),
		DealerScore: dealerScore,
		CardsLeft:   len(game.Deck),
	}
}

func findGame(gameID string) (*Game, bool) {
	gameStore.RLock()
	defer gameStore.RUnlock()

	game, ok := gameStore.games[gameID]
	return game, ok
}

func parseGamePath(path string) (string, string, error) {
	trimmed := strings.Trim(path, "/")
	parts := strings.Split(trimmed, "/")

	if len(parts) < 2 || parts[0] != "games" || parts[1] == "" {
		return "", "", errors.New("game not found")
	}
	if len(parts) == 2 {
		return parts[1], "", nil
	}
	if len(parts) == 3 {
		return parts[1], parts[2], nil
	}

	return "", "", errors.New("game not found")
}

func newGameID() string {
	now := strconv.FormatInt(time.Now().UnixNano(), 36)
	randomPart := strconv.FormatInt(rand.Int63(), 36)
	return now + "-" + randomPart
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}
