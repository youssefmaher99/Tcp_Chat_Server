package server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"test-sse/client"
	"test-sse/room"
)

var ErrConnection = errors.New("connection error")

type TcpServer struct {
	listenAddr string
	ln         net.Listener
	quitch     chan struct{}
}

type Session struct {
	ctx  string
	conn net.Conn
}

var rooms []room.Room
var clients []client.Client

func NewServer(addr string) *TcpServer {
	return &TcpServer{listenAddr: addr, quitch: make(chan struct{})}
}

func (s *TcpServer) Start() error {
	lstn, err := net.Listen("tcp", s.listenAddr)
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
	clientSession := &Session{ctx: "welcome", conn: conn}
	defer conn.Close()
	var err error
loop:
	for {
		fmt.Println(clientSession.ctx)
		switch clientSession.ctx {
		case "welcome":
			err = handleWelcome(clientSession)
			if err != nil {
				break loop
			}
		case "create":
			err = handleCreate(clientSession)
			if err != nil {
				break loop
			}
		case "room":
			err = handleCreate(clientSession)
			if err != nil {
				break loop
			}
		default:
			log.Fatal("loop invalid context")
		}
	}
}

func handleWelcome(session *Session) error {
	session.conn.Write([]byte("\033[2J\033[1;1H"))
	session.conn.Write([]byte("\nWelcome to the app\n"))
	session.conn.Write([]byte("1-join room\n"))
	session.conn.Write([]byte("2-create room\n"))

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
			session.ctx = "join"
		} else {
			session.ctx = "create"
		}

		return nil
	}

}

func handleCreate(session *Session) error {
	var room room.Room
	var client client.Client
	prompt := []string{"room name : ", "room size(max 255) : ", "username : "}
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
				session.conn.Write([]byte(err.Error()))
				continue
			}
			if maxconns < 0 || maxconns > 255 {
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
	session.ctx = "room"
	rooms = append(rooms, room)
	clients = append(clients, client)
	return nil
}

func handleRoom(session *Session) {
}

func readInput(conn net.Conn) ([]byte, error) {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		if err == io.EOF {
			log.Printf("client %s disconnected\n", conn.RemoteAddr().String())
			return []byte{}, err
		}
		log.Println("read from connection error: ", err)
		return []byte{}, err
	}
	return buf[:n-1], nil
}
func readInput2(conn net.Conn) ([]byte, error) {
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Printf("client %s disconnected\n", conn.RemoteAddr().String())
				return []byte{}, err
			}
			log.Println("read from connection error: ", err)
			return []byte{}, err
		}
		return buf[:n-1], nil
	}
}

func (s *TcpServer) DisplayPrompt(session *Session) {
	switch session.ctx {
	case "welcome":
		s.DisplayWelcomePrompt(session.conn)
	case "join":
		s.DisplayJoinRoomPrompt(session.conn)
	case "create":
		s.DisplayCreateRoomPrompt(session.conn)
	default:
		panic("Can't display context")
	}
}

func (s *TcpServer) DisplayWelcomePrompt(conn net.Conn) {
	// TODO: could be optimised as one conn.Write
	conn.Write([]byte("\033[2J\033[1;1H"))
	conn.Write([]byte("\nWelcome to the app\n"))
	conn.Write([]byte("1-join room\n"))
	conn.Write([]byte("2-create room\n"))
	conn.Write([]byte("choice : "))
}

func (s *TcpServer) DisplayJoinRoomPrompt(conn net.Conn) {
	conn.Write([]byte("\033[2J\033[1;1H"))
	conn.Write([]byte("\njoin room tab : "))
}

func (s *TcpServer) DisplayCreateRoomPrompt(conn net.Conn) func() {
	conn.Write([]byte("\033[2J\033[1;1H"))
	conn.Write([]byte("\ncreate room name : "))
	promptQueue := [][]byte{[]byte("\ncreate maximum connections : "), []byte("\nowner name : ")}
	idx := 0
	return func() {
		conn.Write(promptQueue[idx])
	}
}

func parseInput(input []byte, session *Session) error {
	var err error
	switch session.ctx {
	case "welcome":
		err = parseWelcomeInput(input, session)
	case "create":
		// err = parseCreateRoomInput(input)
	case "join":
		// err = parseJoinRoomInput(input)
	case "room":
		// err = parseInRoomInput(input)
	default:
		panic("Can't parse message context")
	}

	if err != nil {
		return err
	}
	return nil
}

func parseWelcomeInput(input []byte, session *Session) error {
	inp := string(input)
	if inp != "1" && inp != "2" {
		return errors.New("invalid entry")
	}

	if inp == "1" {
		session.ctx = "join"
	} else {
		session.ctx = "create"
	}

	return nil
}

// func parseCreateRoomInput(room *room.Room, input []byte) error {

// 	return nil
// }

// func parseInRoomInput(input []byte) error {
// 	return nil
// }

// func parseJoinRoomInput(input []byte) error {
// 	return nil
// }

// func acceptWelcomeInput(session *Session) error {
// 	buf := make([]byte, 1024)
// 	prevCtx := session.ctx
// 	for {
// 		inp, _ := readInput(buf, session)
// 		parseInput(inp, session)
// 		nextCtx := session.ctx
// 		if prevCtx != nextCtx {
// 			break
// 		}
// 	}
// 	return nil
// }

// func acceptCreateRoomInput(session *Session) error {
// 	buf := make([]byte, 1024)
// 	prevCtx := session.ctx
// 	// room := room.Room{}
// 	for {
// 		inp, _ := readInput(buf, session)
// 		parseInput(inp, session)
// 		nextCtx := session.ctx
// 		if prevCtx != nextCtx {
// 			break
// 		}
// 	}
// 	return nil
// }

// func readInput(buf []byte, session *Session) ([]byte, error) {
// 	n, err := session.conn.Read(buf)
// 	if err != nil {
// 		if err == io.EOF {
// 			log.Printf("client %s disconnected\n", session.conn.RemoteAddr().String())
// 			return []byte{}, err
// 		}
// 		log.Println("read from connection error: ", err)
// 		return []byte{}, err
// 	}
// 	return buf[:n-1], nil
// }

// func AcceptInput(session *Session) error {
// 	var err error
// 	switch session.ctx {
// 	case "welcome":
// 		err = acceptWelcomeInput(session)
// 	case "create":
// 		err = acceptCreateRoomInput(session)
// 	// case "join":
// 	// 	err = createRoomInput(session)
// 	default:
// 		panic("invalid accept input context")
// 	}
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
