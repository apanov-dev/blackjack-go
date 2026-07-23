package main

import (
	"fmt"
	"os"
)

func startMenu() {
	var startTheGame int

	for {
		fmt.Println("Press 1 to start game")
		fmt.Println("Press 2 to exit")

		fmt.Scanln(&startTheGame)

		if startTheGame == 1 {
			break
		} else if startTheGame == 2 {
			fmt.Println("Exiting...")
			os.Exit(0)
		}
		fmt.Println("Invalid input. Try again.")
	}
}

func hit(pcards []card, deck []card) ([]card, []card) {
	choice, deck := randomPicker(deck)
	pcardsplus := append(pcards, choice)
	return deck, pcardsplus
}

func stand(pcards []card, dcards []card, deck []card) (int, int) {
	pscore := cardTranslate(pcards)
	fmt.Println("Player stands")
	fmt.Println("Player cards and score are", pcards, pscore)
	dealerLogic(dcards, deck)
	dscore := cardTranslate(dcards)
	return pscore, dscore
}

func StartGame() {
	var playersMove string

	startMenu()

	player := Player{}
	dealer := Dealer{}

	deck := deckCreator()
	fmt.Println("Deck is", deck)
	fmt.Println("Length", len(deck))

	deck = FirstDistribution(&player, &dealer, deck)

	dealerCardCount := cardTranslate(dealer.cards)
	playerCardCount := cardTranslate(player.cards)

	fmt.Println("Dealer cards is:", dealer.cards, dealerCardCount)
	fmt.Println("Your cards is:", player.cards, playerCardCount)

	fmt.Println("Check for dealer instant BJ:", instantBJcheck(dealerCardCount))
	fmt.Println("Check for player instant BJ:", instantBJcheck(playerCardCount))

GAME_LOOP:
	for {
		fmt.Println("Choose h to hit, s to stand")
		fmt.Scanln(&playersMove)

		switch playersMove {
		case "h":
			deck, player.cards = hit(player.cards, deck)
			playerCardCount = cardTranslate(player.cards)
			fmt.Println("Your cards is:", player.cards, playerCardCount)
			if playerCardCount > 21 {
				fmt.Println("You lose")
			}
		case "s":
			pscore, dscore := stand(player.cards, dealer.cards, deck)
			winCheck(pscore, dscore)
			break GAME_LOOP
		}
	}
}
