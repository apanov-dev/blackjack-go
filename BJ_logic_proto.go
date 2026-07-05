package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

func deckCreator(size int) []string { //Создание колоды
	deck := make([]string, size)
	counter := 0
	card := 2

	for i := 0; i < size; i++ {
		if card == 10 {
			if i+4 >= size {

				if counter == 4 {
					counter = 0
				}
				counter++
				deck[i] = "11" + "," + strconv.Itoa(counter)
				continue
			}

			if counter == 4 {
				counter = 0
			}
			counter++
			deck[i] = strconv.Itoa(card) + "," + strconv.Itoa(int(counter))
			continue
		}

		if counter == 4 {
			card += 1
			counter = 0

		}
		counter++
		deck[i] = strconv.Itoa(int(card)) + "," + strconv.Itoa(int(counter))
	}
	return deck
}

func randomPicker(deck []string) (string, []string) { //Выбирает случайную карту из колоды и удаляет ее из колоды
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

type Player struct {
	cards []string
}

type Diller struct {
	cards []string
}

func FirstDistribution(p *Player, d *Diller, deck []string) []string { //Первая раздача карт, игрок и диллер получают по 2 карты по очареди
	var st string
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

func cardParser(deck []string) ([]int, int) {
	newDeck := make([]int, len(deck))
	count := 0
	for i, cardStr := range deck {
		newDeck[i], _ = strconv.Atoi(strings.Split(cardStr, ",")[0])
		count += newDeck[i]
	}
	return newDeck, count
}

func instantBJcheck(count int) bool {
	if count == 21 {
		return true
	}
	return false
}

func hit(pcards []string, deck []string) ([]string, []string) {
	choice, deck := randomPicker(deck)
	pcardsplus := append(pcards, choice)
	return deck, pcardsplus
}

func main() {
	player := Player{}
	diller := Diller{}
	var wantToHit string = "h"

	fmt.Println("let's play Black Jack!")
	deck := deckCreator(52)
	fmt.Println("Deck is", deck)
	fmt.Println("Length", len(deck))

	deck = FirstDistribution(&player, &diller, deck)

	_, dillerCount := cardParser(diller.cards)
	_, playerCount := cardParser(player.cards)

	fmt.Println("Diller cards is:", diller.cards, dillerCount)
	fmt.Println("Your cards is:", player.cards, playerCount)

	fmt.Println("Check for diller instant BJ:", instantBJcheck(dillerCount))
	fmt.Println("Check for player instant BJ:", instantBJcheck(playerCount))

	for wantToHit == "h" {
		fmt.Scanln(&wantToHit)
		fmt.Println("Press h to hit")
		if wantToHit == "h" {
			deck, player.cards = hit(player.cards, deck)
			_, playerCount = cardParser(player.cards)
			fmt.Println("Your cards is:", player.cards, playerCount)
		}
	}
}
