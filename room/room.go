package room

import (
	"fmt"
	"net"
	"sync"
	"test-sse/client"
	"test-sse/event"
	"test-sse/message"

	"github.com/google/uuid"
)

type Event interface {
	Send()
}

type Room struct {
	Id            string
	Name          string
	Owner         client.Client
	MaxConns      uint8
	Conns         map[net.Conn]struct{}
	BroadcastChan chan Event
	lock          *sync.RWMutex
}

func CreateRoom() Room {
	return Room{lock: &sync.RWMutex{}, Id: uuid.New().String()}
}

func (r *Room) Join(conn net.Conn, name string) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.Conns[conn] = struct{}{}
	r.BroadcastChan <- event.JoinEvent{Name: name}
}

func (r *Room) Leave(conn net.Conn, name string) {
	r.lock.RLock()
	if r.Owner.Conn == conn {
		r.BroadcastChan <- event.CloseEvent{}
	} else {
		// fmt.Println(r.Conns, r.BroadcastChan)
		if r.Conns != nil && r.BroadcastChan != nil {
			delete(r.Conns, conn)
			r.BroadcastChan <- event.LeaveEvent{Name: name}
		}
	}
	r.lock.RUnlock()
}

func (r *Room) GetConns() []net.Conn {
	conns := []net.Conn{}
	r.lock.RLock()
	defer r.lock.RUnlock()
	for conn := range r.Conns {
		conns = append(conns, conn)
	}
	return conns
}

func (r *Room) Broadcast() {
loop:
	for r.BroadcastChan != nil {
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

			break loop
		case message.Message:
			conns := r.GetConns()
			for _, conn := range conns {
				conn.Write([]byte(fmt.Sprintf("[%s] %s", ev.Owner.Name, string(ev.Text))))
			}
		}
	}
	r.BroadcastChan = nil
	r.Conns = nil
	r = nil
	//HACK : without having a println the function is still seen in the switch case which is causing data race when running -race and owner close the room then a client try to leave
}
