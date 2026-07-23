package main

import "math/rand"

func deckCreator() []card {
	deck := make([]card, 0, 52)
	ranks := []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}
	suit := []string{"Spades", "Hearts", "Diamonds", "Clubs"}

	for _, rankstr := range ranks {
		for _, suitstr := range suit {
			newCard := card{
				Rank: rankstr,
				Suit: suitstr,
			}
			deck = append(deck, newCard)
		}
	}
	return deck
}

func randomPicker(deck []card) (card, []card) { //Выбирает случайную карту из колоды и удаляет ее из колоды
	index := rand.Intn(len(deck))
	choice := deck[index]

	for i := 0; i < len(deck); i++ {
		if deck[i] == choice {
			deck = append(deck[:i], deck[i+1:]...)
			break
		}
	}
	return choice, deck
}
