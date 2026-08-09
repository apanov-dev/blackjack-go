package main

type Player struct {
	cards []Card `json:"p.cards"`
}

type Dealer struct {
	cards []Card `json:"d.cards"`
}

func FirstDistribution(p *Player, d *Dealer, deck []Card) []Card { // Deals two cards to the player and dealer by turns.
	var st Card
	for i := 1; i < 5; i++ {
		if i%2 != 0 {
			st, deck = randomPicker(deck)
			p.cards = append(p.cards, st)
		} else {
			st, deck = randomPicker(deck)
			d.cards = append(d.cards, st)
		}
	}
	return deck
}

func dealerLogic(dcards []Card, deck []Card) ([]Card, []Card) {
	score := cardTranslate(dcards)

	for score < 17 {
		deck, dcards = hit(dcards, deck)
		score = cardTranslate(dcards)
	}

	return dcards, deck
}
