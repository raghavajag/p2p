package p2p

import (
	"net"

	"github.com/sirupsen/logrus"
)

type TCPTransport struct {
	listenAddr string
	listener   net.Listener
	AddPeer    chan *Peer
	RemovePeer chan *Peer
}

func NewTCPTransport(addr string) *TCPTransport {
	return &TCPTransport{
		listenAddr: addr,
		AddPeer:    make(chan *Peer),
		RemovePeer: make(chan *Peer),
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	ln, err := net.Listen("tcp", t.listenAddr)
	if err != nil {
		return err
	}
	t.listener = ln
	for {
		conn, err := ln.Accept()
		if err != nil {
			logrus.Error(err)
			continue
		}
		peer := &Peer{
			conn:       conn,
			listenAddr: conn.RemoteAddr().String(),
		}
		t.AddPeer <- peer
	}
}
