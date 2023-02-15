package room

import (
	"net"
	"test-sse/client"
)

type Room struct {
	Name     string
	Owner    client.Client
	MaxConns int
	Conns    []net.Conn
}
