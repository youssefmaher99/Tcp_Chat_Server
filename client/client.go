package client

import (
	"net"
	"sync"
)

type Client struct {
	Name string
	Conn net.Conn
}

type Clients struct {
	Store map[net.Conn]Client
	Lock  *sync.RWMutex
}

func (cs Clients) Set(key net.Conn, value Client) {
	cs.Lock.Lock()
	cs.Store[key] = value
	cs.Lock.Unlock()
}

func (cs Clients) Get(key net.Conn) Client {
	cs.Lock.RLock()
	client := cs.Store[key]
	cs.Lock.RUnlock()
	return client
}

func (cs Clients) Remove(key net.Conn) {
	cs.Lock.Lock()
	delete(cs.Store, key)
	cs.Lock.Unlock()
}

func (cs Clients) Exist(key net.Conn) bool {
	cs.Lock.Lock()
	_, ok := cs.Store[key]
	cs.Lock.Unlock()
	return ok
}
