package server

import (
	"errors"
	"io"
	"log"
	"net"
)

type TcpServer struct {
	listenAddr string
	ln         net.Listener
	quitch     chan struct{}
}

type Session struct {
	ctx  string
	conn net.Conn
}

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
	clientSession := Session{ctx: "welcome", conn: conn}
	defer conn.Close()

	// currentPrompt:= "welcome"

	/*		Welcome		*/
	for {
		// s.DisplayWelcomePrompt(conn)
		s.DisplayPrompt(clientSession)
		if err := AcceptInput(clientSession); err != nil {
			break
		}
	}

}

func (s *TcpServer) DisplayPrompt(session Session) {
	switch session.ctx {
	case "welcome":
		s.DisplayWelcomePrompt(session.conn)
	default:
		panic("Can't display context")
	}
}

func (s *TcpServer) DisplayWelcomePrompt(conn net.Conn) {
	// TODO: could be optimised as one conn.Write
	conn.Write([]byte("\nWelcome to the app\n"))
	conn.Write([]byte("1-join room\n"))
	conn.Write([]byte("2-create room\n"))
	conn.Write([]byte("choice : "))
}

func (s *TcpServer) DisplayCreateRoomPrompt(conn net.Conn) {
	conn.Write([]byte("\nroom name : "))
}

func parseInput(input []byte, ctx string) error {
	var err error
	switch ctx {
	case "welcome":
		err = parseWelcomeInput(input)
	case "createRoom":
		err = parseCreateRoomInput(input)
	case "joinRoom":
		err = parseJoinRoomInput(input)
	case "inRoom":
		err = parseInRoomInput(input)
	default:
		panic("Can't parse message context")
	}

	if err != nil {
		return err
	}
	return nil
}

func parseWelcomeInput(input []byte) error {
	inp := string(input)
	if inp != "1" && inp != "2" {
		return errors.New("invalid entry")
	}
	return nil
}

func parseCreateRoomInput(input []byte) error {
	return nil
}

func parseInRoomInput(input []byte) error {
	return nil
}

func parseJoinRoomInput(input []byte) error {
	return nil
}

func AcceptInput(session Session) error {

	buf := make([]byte, 1024)
	for {
		n, err := session.conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Printf("client %s disconnected\n", session.conn.RemoteAddr().String())
				return err
			}
			log.Println("read from connection error: ", err)
			continue
		}
		err = parseInput(buf[:n-1], session.ctx)
		if err != nil {
			session.conn.Write([]byte(err.Error() + "\n" + "choice : "))
		}
	}
}
