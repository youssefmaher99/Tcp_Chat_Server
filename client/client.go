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
	lock  *sync.RWMutex
}

func (cs Clients) Set(key net.Conn, value Client) {
	cs.lock.Lock()
	cs.Store[key] = value
	cs.lock.Unlock()
}

func (cs Clients) Get(key net.Conn) Client {
	cs.lock.RLock()
	client := cs.Store[key]
	cs.lock.RUnlock()
	return client
}

func (cs Clients) Remove(key net.Conn) {
	cs.lock.Lock()
	delete(cs.Store, key)
	cs.lock.Unlock()
}

func (cs Clients) Exist(key net.Conn) bool {
	cs.lock.Lock()
	_, ok := cs.Store[key]
	cs.lock.Unlock()
	return ok
}

func CreateClientsMap() Clients {
	return Clients{lock: &sync.RWMutex{}, Store: make(map[net.Conn]Client)}
}
