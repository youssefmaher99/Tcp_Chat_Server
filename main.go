package main

import (
	"log"
	"test-sse/server"
)

func main() {
	serv := server.NewServer("127.0.0.1:8080")
	log.Fatal(serv.Start())
}
