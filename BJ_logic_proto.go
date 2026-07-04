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

func main() {
	player := Player{}
	diller := Diller{}
	fmt.Println("let's play Black Jack!")
	deck := deckCreator(52)
	fmt.Println("Deck is", deck)
	fmt.Println("Length", len(deck))

	deck = FirstDistribution(&player, &diller, deck)

	fmt.Println("Your cards is:", player.cards)
	fmt.Println("Diller cards is:", diller.cards)

	dillerDeck, count := cardParser(diller.cards)

	fmt.Println("Check for diller instant BJ:", dillerDeck, count)
}
