package p2p

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
)

type Message struct {
	From    net.Addr
	Payload io.Reader
}

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
	handler  Handler
	peers    map[string]*Peer
	listener net.Listener
	ServerConfig
	addPeer    chan *Peer
	removePeer chan *Peer
	msgCh      chan *Message
}

func NewServer(cfg ServerConfig) *Server {
	return &Server{
		peers:        make(map[string]*Peer),
		ServerConfig: cfg,
		addPeer:      make(chan *Peer),
		msgCh:        make(chan *Message),
		handler:      &DefaultHandler{},
		removePeer:   make(chan *Peer),
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
func (s *Server) handleConn(p *Peer) {
	scanner := bufio.NewScanner(p.conn)

	for scanner.Scan() {
		receivedData := scanner.Text()

		s.msgCh <- &Message{
			From:    p.conn.RemoteAddr(),
			Payload: bytes.NewReader([]byte(receivedData)),
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading from %s: %s\n", p.conn.RemoteAddr(), err)
	}
	s.removePeer <- p
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
		s.addPeer <- peer
		peer.Send([]byte("Hello from server\n"))
		go s.handleConn(peer)
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
func (s *Server) Connect(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	peer := &Peer{
		conn:       conn,
		listenAddr: conn.RemoteAddr().String(),
	}
	s.addPeer <- peer
	return nil
}

func (s *Server) loop() {
	for {
		select {
		case peer := <-s.removePeer:
			addr := peer.listenAddr
			delete(s.peers, addr)
			fmt.Printf("Peer %s disconnected\n", addr)
		case peer := <-s.addPeer:
			s.peers[peer.listenAddr] = peer
			fmt.Printf("Peer %s connected\n", peer.listenAddr)
		case msg := <-s.msgCh:
			if err := s.handler.HandleMessage(msg); err != nil {
				panic(err)
			}
		}
	}
}
