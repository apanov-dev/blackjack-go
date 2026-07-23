package main

import "fmt"

func instantBJcheck(count int) bool {
	if count == 21 {
		return true
	}
	return false
}

func winCheck(pscore int, dscore int) {
	if pscore > dscore {
		fmt.Println("You win!")
		startMenu()
	} else if pscore == dscore {
		fmt.Println("It's a tie!")
		startMenu()
	}
	fmt.Println("You lose!")
	startMenu()
}
