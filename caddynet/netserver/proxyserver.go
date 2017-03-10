package netserver

import (
	"fmt"
	"net"
)

// ProxyServer is an implementation of the
// caddy.Server interface type
type ProxyServer struct {
	LocalTCPAddr string
	listener     net.Listener
	DestTCPAddr  string
}

// NewProxyServer returns a new proxy server
func NewProxyServer(l string, d string) (*ProxyServer, error) {
	return &ProxyServer{
		LocalTCPAddr: l,
		DestTCPAddr:  d,
	}, nil
}

// Listen starts listening by creating a new listener
// and returning it. It does not start accepting
// connections.
func (s *ProxyServer) Listen() (net.Listener, error) {
	return net.Listen("tcp", fmt.Sprintf("%s", s.LocalTCPAddr))
}

// Serve starts serving using the provided listener.
// Serve blocks indefinitely, or in other
// words, until the server is stopped.
func (s *ProxyServer) Serve(ln net.Listener) error {
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
		fmt.Printf("ProxyServer: accepted from %s\n", conn.RemoteAddr())

		go p.proxy()
	}
}

// Stop stops s gracefully (or forcefully after timeout) and
// closes its listener.
func (s *ProxyServer) Stop() error {

	fmt.Println("ProxyServer Stop")
	err := s.listener.Close()
	if err != nil {
		return err
	}

	return nil
}

// ListenPacket is a no-op to satisfy caddy.Server interface
func (s *ProxyServer) ListenPacket() (net.PacketConn, error) { return nil, nil }

// ServePacket is a no-op to satisfy caddy.Server interface
func (s *ProxyServer) ServePacket(net.PacketConn) error { return nil }

// OnStartupComplete lists the sites served by this server
// and any relevant information
func (s *ProxyServer) OnStartupComplete() {
	fmt.Println("ProxyServer OnStartupComplete: Proxying from ", s.LocalTCPAddr, " -> ", s.DestTCPAddr)
}
