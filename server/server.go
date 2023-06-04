package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"sync"
	"test-sse/client"
	"test-sse/message"
	r "test-sse/room"
	"time"
)

var ErrConnection = errors.New("connection error")
var ErrDisconnectClient = errors.New("disconnect client")

const (
	Welcome uint8 = iota
	Create
	Join
	Room
)

type TcpServer struct {
	listenAddr string
	ln         net.Listener
	quitch     chan struct{}
}

type Session struct {
	ctx    uint8
	conn   net.Conn
	room   *r.Room
	roomId string
}

var rooms []*r.Room
var clients = client.CreateClientsMap()

func NewServer(addr string) *TcpServer {
	return &TcpServer{listenAddr: addr, quitch: make(chan struct{})}
}

func (s *TcpServer) Start() error {
	lstn, err := net.Listen("tcp", "127.0.0.1"+s.listenAddr)
	if err != nil {
		return err
	}

	defer lstn.Close()
	s.ln = lstn

	go s.AcceptLoop()

	<-s.quitch
	return nil
}

func (s *TcpServer) AcceptLoop() {
	log.Printf("server starting on port [%s]\n", s.listenAddr)
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			log.Println("accept connection error: ", err)
			continue
		}
		log.Printf("client %s connected\n", conn.RemoteAddr().String())
		go s.HandleConn(conn)
	}
}

func (s *TcpServer) HandleConn(conn net.Conn) {
	clientSession := &Session{ctx: Welcome, conn: conn}
	defer conn.Close()
	var err error
loop:
	for {
		// fmt.Println(clientSession.ctx)
		switch clientSession.ctx {
		case Welcome:
			err = handleWelcome(clientSession)
			if err != nil {
				break loop
			}
		case Create:
			err = handleCreate(clientSession)
			if err != nil {
				break loop
			}
		case Join:
			err = handleJoin(clientSession)
			if err != nil {
				break loop
			}
		case Room:
			err = handleRoom(clientSession)
			log.Println(err)
			if err != nil {
				break loop
			}
		default:
			log.Fatal("loop invalid context")
		}
	}
}

func handleWelcome(session *Session) error {
	session.conn.Write([]byte("\033[2J\033[1;1H\nWelcome to the app\n1-join room\n2-create room\n"))

	for {
		session.conn.Write([]byte("choice : "))
		inp, err := readInput(session.conn)
		if err != nil {
			return ErrConnection
		}
		if string(inp) != "1" && string(inp) != "2" {
			session.conn.Write([]byte("invalid entry\n"))
			continue
		}

		if string(inp) == "1" {
			session.ctx = Join
		} else {
			session.ctx = Create
		}

		return nil
	}

}

func handleCreate(session *Session) error {
	room := r.CreateRoom()
	room.Conns = make(map[net.Conn]struct{})
	var client client.Client
	prompt := []string{"room name : ", "room size(max 255) : ", "clientname : "}
	session.conn.Write([]byte("\033[2J\033[1;1H"))

	ptr_idx := 0
	for ptr_idx != len(prompt) {
		session.conn.Write([]byte(prompt[ptr_idx]))
		inp, err := readInput(session.conn)
		if err != nil {
			return ErrConnection
		}
		if ptr_idx == 0 {
			room.Name = string(inp)
		} else if ptr_idx == 1 {
			maxconns, err := strconv.Atoi(string(inp))
			if err != nil {
				session.conn.Write([]byte("invalid entry\n"))
				continue
			}
			if maxconns <= 1 || maxconns > 255 {
				session.conn.Write([]byte("invalid entry\n"))
				continue
			}
			room.MaxConns = uint8(maxconns)
		} else if ptr_idx == 2 {
			client.Name = string(inp)
			client.Conn = session.conn
		}
		ptr_idx++
	}

	room.BroadcastChan = make(chan r.Event)
	room.Owner = client
	// room.Conns = append(room.Conns, client.Conn)

	session.ctx = Room
	session.room = &room
	rooms = append(rooms, &room)
	// clients[session.conn] = client
	clients.Set(session.conn, client)
	go func() {
		room.Broadcast()
	}()

	// HACK : to make sure that the owner will see x joined the room before initiating the broadcast channel
	go func() {
		time.Sleep(time.Millisecond * 10)
		room.Join(client.Conn, clients.Get(session.conn).Name)
	}()
	return nil
}

func handleJoin(session *Session) error {
	session.conn.Write([]byte("\033[2J\033[1;1H"))
	session.conn.Write([]byte("Select room you want to join\n"))
	session.conn.Write([]byte("0-return to welcome page\n"))
	for idx, room := range rooms {
		session.conn.Write([]byte(fmt.Sprintf("%d-%s (owner: %s) (%d/%d)\n", idx+1, room.Name, room.Owner.Name, len(room.Conns), room.MaxConns)))
	}

	for {
		session.conn.Write([]byte("choice : "))
		inp, err := readInput(session.conn)
		if err != nil {
			return ErrConnection
		}

		intInp, err := strconv.Atoi(string(inp))
		if err != nil {
			session.conn.Write([]byte("invalid entry\n"))
			continue
		}

		if intInp < 0 || intInp > len(rooms) {
			session.conn.Write([]byte("invalid entry\n"))
			continue
		}

		if intInp == 0 {
			session.ctx = Welcome
			return nil
		}

		if len(rooms[intInp-1].Conns) == int(rooms[intInp-1].MaxConns) {
			session.conn.Write([]byte("room is full choose another room\n"))
			continue
		}

		session.ctx = Room
		session.room = rooms[intInp-1]
		mu := sync.Mutex{}
		mu.Lock()
		defer mu.Unlock()
		if !clients.Exist(session.conn) {
			err = registerClient(session)
			if err != nil {
				return err
			}
		}
		rooms[intInp-1].Join(session.conn, clients.Get(session.conn).Name)
		return nil
	}
}

func handleRoom(session *Session) error {
	session.conn.Write([]byte("\033[2J\033[1;1H"))
	roomId := session.room.Id
	err := readInputContinuously(session.conn, session.room, roomId)
	return err
}

func readInput(conn net.Conn) ([]byte, error) {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		if err == io.EOF {
			log.Printf("client %s disconnected\n", conn.RemoteAddr().String())
			clients.Remove(conn)
			return []byte{}, err
		}
		log.Println("read from connection error: ", err)
		return []byte{}, err
	}
	return buf[:n-1], nil
}

// BUG: data race because 2 goroutines have access to the same reference
func readInputContinuously(conn net.Conn, room *r.Room, roomId string) error {
	buf := bufio.NewReaderSize(conn, 1024)
	var message = message.Message{Owner: clients.Get(conn)}
	for {
		input, err := buf.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				log.Printf("client %s disconnected\n", conn.RemoteAddr().String())
			} else {
				log.Println("read from connection error: ", err)
			}
			if room != nil {
				room.Leave(conn, clients.Get(conn).Name)
				if room.Owner.Conn == conn {
					// removeRoom(room)
				}
			}
			clients.Remove(conn)
			return err
		}
		// brodcast message to all clients in the room
		message.Text = input
		select {
		case room.BroadcastChan <- message:
		default:
			return ErrDisconnectClient
		}
	}
}

func registerClient(session *Session) error {
	session.conn.Write([]byte("\033[2J\033[1;1H"))
	session.conn.Write([]byte("clientname : "))

	inp, err := readInput(session.conn)
	if err != nil {
		return err
	}
	var client client.Client
	client.Name = string(inp)
	client.Conn = session.conn
	clients.Set(session.conn, client)
	return nil
}

func removeRoom(room *r.Room) {
	for i := 0; i < len(rooms); i++ {
		if room == rooms[i] {
			rooms = append(rooms[:i], rooms[i+1:]...)
			break
		}
	}
}
