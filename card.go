package main

import "strconv"

type card struct {
	Rank string
	Suit string
}

func cardTranslate(deck []card) int {
	count := 0

	for _, c := range deck {
		var value int

		switch c.Rank {
		case "J", "Q", "K":
			value = 10
		case "A":
			value = 11
		default:
			var err error
			value, err = strconv.Atoi(c.Rank)
			if err != nil {
				value = 0
			}
		}
		count += value
	}

	return count
}
