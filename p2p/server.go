package p2p

import (
	"bufio"
	"fmt"
	"net"
)

type Peer struct {
	conn       net.Conn
	listenAddr string
}

func (p *Peer) Send(data []byte) error {
	_, err := p.conn.Write(data)
	return err
}

type ServerConfig struct {
	ListenAddr string
}
type Server struct {
	peers    map[string]*Peer
	listener net.Listener
	ServerConfig
	addPeer chan *Peer
}

func NewServer(cfg ServerConfig) *Server {
	return &Server{
		peers:        make(map[string]*Peer),
		ServerConfig: cfg,
		addPeer:      make(chan *Peer),
	}
}

func (s *Server) Start() {
	go s.loop()
	if err := s.listen(); err != nil {
		panic(err)
	}
	fmt.Printf("Listening on %s\n", s.ListenAddr)
	s.acceptLoop()
}
func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		receivedData := scanner.Text()
		fmt.Printf("Received: %s\n", receivedData)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading from %s: %s\n", conn.RemoteAddr(), err)
	}
}
func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			panic(err)
		}
		peer := &Peer{
			conn:       conn,
			listenAddr: conn.RemoteAddr().String(),
		}
		peer.Send([]byte("Hello from server\n"))
		s.addPeer <- peer
		go s.handleConn(conn)
	}
}
func (s *Server) listen() error {
	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	s.listener = ln
	return nil
}
func (s *Server) loop() {
	for {
		select {
		case peer := <-s.addPeer:
			s.peers[peer.listenAddr] = peer
			fmt.Printf("Peer %s connected\n", peer.listenAddr)
		}
	}
}
