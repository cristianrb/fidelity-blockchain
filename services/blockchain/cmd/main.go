package main

import (
	"flag"
	"github.com/cristianrb/fidelityblockchain/api"
)

func main() {
	address := flag.String("address", "0.0.0.0", "IP Address for Blockchain Server")
	port := flag.Int("port", 5000, "Port for Blockchain Server")
	flag.Parse()

	server := api.NewServer(*address, uint16(*port))
	err := server.Start()
	if err != nil {
		println(err.Error())
		panic("cannot start server")
	}
}
