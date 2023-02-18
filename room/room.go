package room

import (
	"test-sse/client"
)

type Room struct {
	Name     string
	Owner    client.Client
	MaxConns uint8
	Conns    []client.Client
}
