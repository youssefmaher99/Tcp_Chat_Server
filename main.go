package main

import (
	"fmt"
	"log"
	"net"
	"test-sse/server"
)

func main() {
	serv := server.NewServer("127.0.0.1:8080")
	log.Fatal(serv.Start())
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	b := make([]byte, 1024)
	for {
		n, err := conn.Read(b)
		if err != nil {
			fmt.Println("conn read error: ", err)
		}
		msg := b[:n-1]
		fmt.Println(string(msg))
	}
}
