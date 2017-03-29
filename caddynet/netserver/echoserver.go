package netserver

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddytls"
)

// EchoServer is an echo implementation of the
// caddy.Server interface type
type EchoServer struct {
	LocalTCPAddr string
	tcpListener  net.Listener
	udpListener  net.PacketConn
	udpSemaphore chan int
	config       *Config
}

// NewEchoServer returns a new echo server
func NewEchoServer(l string, c *Config) (*EchoServer, error) {
	return &EchoServer{
		LocalTCPAddr: l,
		udpSemaphore: make(chan int, 100),
		config:       c,
	}, nil
}

// Listen starts listening by creating a new listener
// and returning it. It does not start accepting
// connections.
func (s *EchoServer) Listen() (net.Listener, error) {
	var listener net.Listener

	tlsConfig, err := caddytls.MakeTLSConfig([]*caddytls.Config{s.config.TLS})
	if err != nil {
		return nil, err
	}

	inner, err := net.Listen("tcp", fmt.Sprintf("%s", s.LocalTCPAddr))
	if err != nil {
		return nil, err
	}

	if tlsConfig != nil {
		listener = tls.NewListener(inner, tlsConfig)
	} else {
		listener = inner
	}

	return listener, nil

}

// ListenPacket starts listening by creating a new Packet listener
// and returning it. It does not start accepting
// connections.
func (s *EchoServer) ListenPacket() (net.PacketConn, error) {
	return net.ListenPacket("udp", fmt.Sprintf("%s", s.LocalTCPAddr))

}

// Serve starts serving using the provided listener.
// Serve blocks indefinitely, or in other
// words, until the server is stopped.
func (s *EchoServer) Serve(ln net.Listener) error {

	s.tcpListener = ln

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}

		go func(c net.Conn) {
			// Echo all incoming data.
			_, err := io.Copy(c, c)
			if err != nil {
				fmt.Printf("io.Copy error: %v\n", err)
			}

			// Shut down the connection.
			c.Close()
		}(conn)
	}
}

// ServePacket starts serving using the provided listener.
// ServePacket blocks indefinitely, or in other
// words, until the server is stopped.
func (s *EchoServer) ServePacket(con net.PacketConn) error {

	s.udpListener = con

	for {
		s.udpSemaphore <- 1 //semaphore
		go s.echoUDP(con)
	}

}

func (s *EchoServer) echoUDP(con net.PacketConn) {
	defer func() { <-s.udpSemaphore }()

	buf := make([]byte, 4096)
	nr, addr, err := con.ReadFrom(buf)
	if err != nil {
		s.udpListener.Close()
	}
	_, err = con.WriteTo(buf[:nr], addr)
	if err != nil {
		s.udpListener.Close()
	}
}

// Stop stops s gracefully and closes its listener.
func (s *EchoServer) Stop() error {

	err := s.tcpListener.Close()
	if err != nil {
		return err
	}

	err = s.udpListener.Close()
	if err != nil {
		return err
	}

	return nil
}

// OnStartupComplete lists the sites served by this server
// and any relevant information
func (s *EchoServer) OnStartupComplete() {
	if !caddy.Quiet {
		fmt.Println("[INFO] Echoing on port ", s.LocalTCPAddr)
	}
}
