package room

import (
	"net"
	"test-sse/client"
	"test-sse/message"
)

type Room struct {
	Name          string
	Owner         client.Client
	MaxConns      uint8
	Conns         []net.Conn
	BroadcastChan chan message.Message
}
