package netserver

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"

	"github.com/mholt/caddy/caddytls"
)

// EchoServer is an echo implementation of the
// caddy.Server interface type
type EchoServer struct {
	LocalTCPAddr string
	listener     net.Listener
	udpListener  net.PacketConn
	sem          chan int
	config       *Config
}

// NewEchoServer returns a new echo server
func NewEchoServer(l string, c *Config) (*EchoServer, error) {
	return &EchoServer{
		LocalTCPAddr: l,
		sem:          make(chan int, 100),
		config:       c,
	}, nil
}

// Listen starts listening by creating a new listener
// and returning it. It does not start accepting
// connections.
func (s *EchoServer) Listen() (net.Listener, error) {
	tlsConfigs := []*caddytls.Config{s.config.TLS}
	tlsConfig, err := caddytls.MakeTLSConfig(tlsConfigs)
	if err != nil {
		return nil, err
	}

	var (
		inner    net.Listener
		listener net.Listener
	)

	inner, err = net.Listen("tcp", fmt.Sprintf("%s", s.LocalTCPAddr))
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

	s.listener = ln

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
		s.sem <- 1 //semaphore
		go s.echoUDP(con)
	}

}

func (s *EchoServer) echoUDP(con net.PacketConn) {
	defer func() { <-s.sem }()

	buf := make([]byte, 4096)
	nr, addr, err := con.ReadFrom(buf)
	if err != nil {
		s.udpListener.Close()
	}
	nw, err := con.WriteTo(buf[:nr], addr)
	if err != nil {
		s.udpListener.Close()
	}

	fmt.Printf("received: [%d] sent [%d]\n", nr, nw)
}

// Stop stops s gracefully and closes its listener.
func (s *EchoServer) Stop() error {

	fmt.Println("EchoServer Stop")
	err := s.listener.Close()
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
	fmt.Println("EchoServer OnStartupComplete:", s.LocalTCPAddr)
}
