package room

import (
	"net"
	"test-sse/client"
)

type Room struct {
	Name     string
	Owner    client.Client
	MaxConns uint8
	Conns    []net.Conn
}
