package client

import "net"

type Client struct {
	name string
	conn net.Conn
}
