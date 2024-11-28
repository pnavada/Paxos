package main

import (
	"log"
	"time"

	"paxos/paxos/config"
	"paxos/paxos/handlers"
	"paxos/paxos/network"
)

func main() {

	cfg := config.ParseFlags()

	if cfg.ProposalDelay > 0 {
		time.Sleep(time.Duration(cfg.ProposalDelay) * time.Second)
	}

	peer, err := network.NewPeer(cfg.HostsFile, cfg.ProposalValue)
	if err != nil {
		log.Fatalf("Failed to initialize peer: %v", err)
	}

	go peer.Start()

	mh := handlers.NewMessageHandler(peer)
	mh.HandleMessages()

}
