# Telegram Mini App API

## Environment

```bash
TELEGRAM_BOT_TOKEN=123456:bot-token
PORT=8080
TELEGRAM_INIT_DATA_TTL=86400
```

For local frontend development without Telegram, you may enable explicit dev auth:

```bash
TELEGRAM_AUTH_DISABLED=true
```

When dev auth is enabled, pass `X-Dev-Telegram-User-ID` to simulate different Telegram users.

## Auth

The frontend must send `window.Telegram.WebApp.initData` with every protected request.

Preferred header:

```http
X-Telegram-Init-Data: query_id=...&user=...&auth_date=...&hash=...
```

Alternative header:

```http
Authorization: tma query_id=...&user=...&auth_date=...&hash=...
```

Protected endpoints:

- `GET /telegram/me`
- `POST /games`
- `GET /games/{game_id}`
- `POST /games/{game_id}/hit`
- `POST /games/{game_id}/stand`

Public endpoints:

- `GET /health`
- `GET /deck`

## Endpoints

### `GET /telegram/me`

Returns the validated Telegram user.

### `POST /games`

Creates a new blackjack game for the authenticated Telegram user.

### `GET /games/{game_id}`

Returns the current game state. A user can only access their own games.

### `POST /games/{game_id}/hit`

Adds one card to the player hand and returns the updated state.

### `POST /games/{game_id}/stand`

Runs the dealer turn and returns the final state.

## Game Response

```json
{
  "id": "game-id",
  "status": "playing",
  "player_cards": [{"rank": "A", "suit": "Spades"}],
  "dealer_cards": [{"rank": "10", "suit": "Hearts"}],
  "player_score": 21,
  "dealer_score": 17,
  "cards_left": 42
}
```

Possible statuses:

- `playing`
- `player_win`
- `dealer_win`
- `tie`
- `player_bust`
- `dealer_bust`
