package p2p

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"net"

	"github.com/sirupsen/logrus"
)

const (
	TexasHoldem GameVariant = iota
	Other
)

type GameVariant uint8

func (gv GameVariant) String() string {
	switch gv {
	case TexasHoldem:
		return "TEXAS HOLDEM"
	case Other:
		return "other"
	default:
		return "unknown"
	}
}

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
	ListenAddr  string
	Version     string
	GameVariant GameVariant
}
type Server struct {
	peers map[string]*Peer
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

type Handshake struct {
	Version     string
	GameVariant GameVariant
}

func (s *Server) handshake(p *Peer) error {
	hs := Handshake{}
	if err := gob.NewDecoder(p.conn).Decode(&hs); err != nil {
		return err
	}
	fmt.Sprintf("Handshake: %+v\n", hs)
	logrus.WithFields(logrus.Fields{
		"peer":    p.conn.RemoteAddr(),
		"version": hs.Version,
		"variant": hs.GameVariant,
	}).Info("received handshake")
	return nil
}
func (s *Server) SendHandshake(p *Peer) error {
	hs := &Handshake{
		GameVariant: s.GameVariant,
		Version:     s.Version,
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(hs); err != nil {
		return err
	}

	return p.Send(buf.Bytes())
}
func (s *Server) handleMessage(msg *Message) error {
	fmt.Printf("%+v\n", msg)
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
			s.SendHandshake(peer)
			if err := s.handshake(peer); err != nil {
				logrus.Error("handshake with incoming peer failed: ", err)
				continue
			}

			logrus.WithFields(logrus.Fields{
				"addr": peer.listenAddr,
			}).Info("player connected")
			s.peers[peer.listenAddr] = peer
		case msg := <-s.msgCh:
			if err := s.handleMessage(msg); err != nil {
				panic(err)
			}
		}
	}
}
