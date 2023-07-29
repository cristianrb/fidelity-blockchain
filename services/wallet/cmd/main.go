package main

import (
	"flag"
	"github.com/cristianrb/fidelityblockchainwallet/api"
)

func main() {
	address := flag.String("address", "0.0.0.0:8080", "IP Address for Wallet Server")
	flag.Parse()

	server := api.NewServer(*address)
	err := server.Start()
	if err != nil {
		panic("cannot start server")
	}
}
