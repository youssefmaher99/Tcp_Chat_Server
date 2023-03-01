package room

import (
	"fmt"
	"net"
	"sync"
	"test-sse/client"
	"test-sse/event"
	"test-sse/message"
)

type Event interface {
	Send()
}

type Room struct {
	Name          string
	Owner         client.Client
	MaxConns      uint8
	Conns         map[net.Conn]struct{}
	BroadcastChan chan Event
	mu            *sync.Mutex
}

func CreateRoom() Room {
	return Room{mu: &sync.Mutex{}}
}

func (r *Room) Join(conn net.Conn, name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Conns[conn] = struct{}{}
	r.BroadcastChan <- event.JoinEvent{Name: name}
}

func (r *Room) Leave(conn net.Conn, name string) {
	//HACK
	if r.Owner.Conn == conn {
		r.BroadcastChan <- event.CloseEvent{}
	} else {
		delete(r.Conns, conn)
		r.BroadcastChan <- event.LeaveEvent{Name: name}
	}
}

func (r *Room) Exist(conn net.Conn) bool {
	if _, ok := r.Conns[conn]; ok {
		return true
	}
	return false
}

func (r *Room) RoomLen() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.Conns)
}

func (r *Room) GetConns() []net.Conn {
	conns := []net.Conn{}
	r.mu.Lock()
	defer r.mu.Unlock()
	for conn := range r.Conns {
		conns = append(conns, conn)
	}
	return conns
}

func (r *Room) Broadcast() {
loop:
	for {
		switch ev := (<-r.BroadcastChan).(type) {
		case event.JoinEvent:
			conns := r.GetConns()
			for _, conn := range conns {
				conn.Write([]byte("<<< " + ev.Name + " " + "joined the room" + " >>>\n"))
			}
		case event.LeaveEvent:
			conns := r.GetConns()
			for _, conn := range conns {
				conn.Write([]byte(fmt.Sprintf("%s left the room\n", ev.Name)))
			}
		case event.CloseEvent:
			for conn := range r.Conns {
				conn.Write([]byte("ROOM IS CLOSED\n"))
			}
			r.Conns = nil
			r.BroadcastChan = nil
			break loop
		case message.Message:
			conns := r.GetConns()
			for _, conn := range conns {
				conn.Write([]byte(fmt.Sprintf("[%s] %s", ev.Owner.Name, string(ev.Text))))
			}
		}
	}
}
