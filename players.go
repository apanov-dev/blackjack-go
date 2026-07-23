package main

import "fmt"

type Player struct {
	cards []card
}

type Dealer struct {
	cards []card
}

func FirstDistribution(p *Player, d *Dealer, deck []card) []card { //Первая раздача карт, игрок и диллер получают по 2 карты по очареди
	var st card
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

func dealerLogic(dcards []card, deck []card) {
	score := cardTranslate(dcards)

	for {
		if score < 17 {
			dcards, deck = hit(dcards, deck)
			score = cardTranslate(dcards)
			fmt.Println("Dealer hits")
			fmt.Println("Cards: ", dcards, "Score: ", score)
			break
		} else if score > 21 {
			fmt.Println("Dealer busts")
			fmt.Println("Cards: ", dcards, "Score: ", score)
			return
		} else {
			fmt.Println("Dealer stands")
			fmt.Println("Cards: ", dcards, "Score: ", score)
			return
		}
	}
}
