package proxyserver

import (
	"fmt"
	"net"
)

// Server is an echo implementation of the
// caddy.Server interface type
type Server struct {
	LocalTCPAddr string
	listener     net.Listener
	DestTCPAddr  string
}

// NewServer returns a new proxy server
func NewServer(l string, d string) (*Server, error) {
	return &Server{
		LocalTCPAddr: l,
		DestTCPAddr:  d,
	}, nil
}

// Listen starts listening by creating a new listener
// and returning it. It does not start accepting
// connections.
func (s *Server) Listen() (net.Listener, error) {
	return net.Listen("tcp", fmt.Sprintf("%s", s.LocalTCPAddr))
}

// Serve starts serving using the provided listener.
// Serve blocks indefinitely, or in other
// words, until the server is stopped.
func (s *Server) Serve(ln net.Listener) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}

		p := &proxyConnection{
			lconn:       conn,
			laddr:       s.LocalTCPAddr,
			raddr:       s.DestTCPAddr,
			erred:       false,
			closeSignal: make(chan bool),
		}
		fmt.Printf("server: accepted from %s\n", conn.RemoteAddr())

		go p.proxy()
	}
}

// ListenPacket is a no-op to satisfy caddy.Server interface
func (s *Server) ListenPacket() (net.PacketConn, error) { return nil, nil }

// ServePacket is a no-op to satisfy caddy.Server interface
func (s *Server) ServePacket(net.PacketConn) error { return nil }

// OnStartupComplete lists the sites served by this server
// and any relevant information
func (s *Server) OnStartupComplete() {
	fmt.Println("OnStartupComplete: Proxying from ", s.LocalTCPAddr, " -> ", s.DestTCPAddr)
}
