package main

import (
	"fmt"
	"p2p/deck"
)

func main() {
	deck := deck.New()
	for i := 0; i < 52; i++ {
		fmt.Println(deck[i])
	}
}
