package main

import (
	"fmt"
	"p2p/p2p"
	"time"
)

func main() {
	// deck := deck.New()
	// for i := 0; i < 52; i++ {
	// 	fmt.Println(deck[i])
	// }
	cfg := p2p.ServerConfig{
		ListenAddr:  ":8080",
		Version:     "POKER v0.1",
		GameVariant: p2p.TexasHoldem,
	}
	server := p2p.NewServer(cfg)
	go server.Start()
	time.Sleep(1 * time.Second)
	remoteCfg := p2p.ServerConfig{
		ListenAddr:  ":8081",
		Version:     "POKER v0.1",
		GameVariant: p2p.TexasHoldem,
	}
	remoteServer := p2p.NewServer(remoteCfg)
	go remoteServer.Start()
	time.Sleep(1 * time.Second)
	if err := remoteServer.Connect(":8080"); err != nil {
		fmt.Println(err)
	}
	select {}
}
