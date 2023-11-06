package main

import "p2p/p2p"

func main() {
	// deck := deck.New()
	// for i := 0; i < 52; i++ {
	// 	fmt.Println(deck[i])
	// }
	server := p2p.NewServer(p2p.ServerConfig{
		ListenAddr: ":8080",
	})
	server.Start()
}
