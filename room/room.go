package room

import (
	"fmt"
	"net"
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
}

func (r *Room) Join(conn net.Conn, name string) {
	r.Conns[conn] = struct{}{}
	r.BroadcastChan <- event.JoinEvent{Name: name}
}

func (r *Room) Leave(conn net.Conn, name string) {

	//HACK

	r.BroadcastChan <- event.LeaveEvent{Name: name}
	if r.Owner.Conn == conn {
		r.BroadcastChan <- event.CloseEvent{}
		r.Conns = nil
		r.BroadcastChan = nil
	} else {
		delete(r.Conns, conn)
	}
}

func (r *Room) Exist(conn net.Conn) bool {
	if _, ok := r.Conns[conn]; ok {
		return true
	}
	return false
}

func (r *Room) Broadcast() {
	for {
		// ev := <-r.BroadcastChan
		switch ev := (<-r.BroadcastChan).(type) {
		case event.JoinEvent:
			for conn := range r.Conns {
				conn.Write([]byte("<<< " + ev.Name + " " + "joined the room" + " >>>\n"))
			}
		case event.LeaveEvent:
			for conn := range r.Conns {
				conn.Write([]byte(fmt.Sprintf("%s left the room\n", ev.Name)))
			}
		case event.CloseEvent:
			for conn := range r.Conns {
				conn.Write([]byte("ROOM IS CLOSED\n"))
			}
		case message.Message:
			for conn := range r.Conns {
				conn.Write([]byte(fmt.Sprintf("[%s] %s\n", ev.Owner.Name, string(ev.Text))))
			}
		}
	}
}
