package main

import "strconv"

type Card struct {
	Rank string `json:"rank"`
	Suit string `json:"suit"`
}

func cardTranslate(deck []Card) int {
	count := 0
	aces := 0

	for _, c := range deck {
		var value int

		switch c.Rank {
		case "J", "Q", "K":
			value = 10
		case "A":
			value = 11
			aces++
		default:
			var err error
			value, err = strconv.Atoi(c.Rank)
			if err != nil {
				value = 0
			}
		}
		count += value
	}

	for count > 21 && aces > 0 {
		count -= 10
		aces--
	}

	return count
}
