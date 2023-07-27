package main

import (
	"flag"
	"github.com/cristianrb/fidelityblockchain/api"
)

func main() {
	address := flag.String("address", "0.0.0.0:5000", "IP Address for Blockchain Server")
	flag.Parse()

	server := api.NewServer(*address)
	err := server.Start()
	if err != nil {
		panic("cannot start server")
	}
}
