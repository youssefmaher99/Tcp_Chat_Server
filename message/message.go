package message

import "test-sse/client"

type Message struct {
	Text  []byte
	Owner client.Client
}
