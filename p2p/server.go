package p2p

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"

	"github.com/sirupsen/logrus"
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
	handler Handler
	peers   map[string]*Peer
	ServerConfig
	addPeer    chan *Peer
	removePeer chan *Peer
	msgCh      chan *Message
	transport  *TCPTransport
}

func NewServer(cfg ServerConfig) *Server {
	s := &Server{
		peers:        make(map[string]*Peer),
		ServerConfig: cfg,
		addPeer:      make(chan *Peer),
		msgCh:        make(chan *Message),
		handler:      &DefaultHandler{},
		removePeer:   make(chan *Peer),
	}
	tr := NewTCPTransport(s.ListenAddr)
	s.transport = tr
	tr.AddPeer = s.addPeer
	tr.RemovePeer = s.removePeer
	return s
}

func (s *Server) Start() {
	go s.loop()
	logrus.WithFields(logrus.Fields{
		"addr": s.ListenAddr,
	}).Info("server started")

	s.transport.ListenAndAccept()
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
			logrus.WithFields(logrus.Fields{
				"addr": peer.listenAddr,
			}).Info("player disconnected")
			addr := peer.listenAddr
			delete(s.peers, addr)
		case peer := <-s.addPeer:
			logrus.WithFields(logrus.Fields{
				"addr": peer.listenAddr,
			}).Info("player connected")
			s.peers[peer.listenAddr] = peer
		case msg := <-s.msgCh:
			if err := s.handler.HandleMessage(msg); err != nil {
				panic(err)
			}
		}
	}
}
