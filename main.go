package main

import (
	"log"
	"test-sse/server"
)

func main() {
	serv := server.NewServer(":5000")
	log.Fatal(serv.Start())
}
