package main

func hit(pcards []Card, deck []Card) ([]Card, []Card) {
	choice, deck := randomPicker(deck)
	pcardsplus := append(pcards, choice)
	return deck, pcardsplus
}
