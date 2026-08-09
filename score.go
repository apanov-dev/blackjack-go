package main

type GameStatus string

const (
	StatusPlaying    GameStatus = "playing"
	StatusPlayerWin  GameStatus = "player_win"
	StatusDealerWin  GameStatus = "dealer_win"
	StatusTie        GameStatus = "tie"
	StatusPlayerBust GameStatus = "player_bust"
	StatusDealerBust GameStatus = "dealer_bust"
)

func instantBJcheck(count int) bool {
	return count == 21
}

func winCheck(pscore int, dscore int) GameStatus {
	if pscore > 21 {
		return StatusPlayerBust
	}
	if dscore > 21 {
		return StatusDealerBust
	}
	if pscore > dscore {
		return StatusPlayerWin
	}
	if pscore < dscore {
		return StatusDealerWin
	}
	return StatusTie
}
