package netserver

import (
	"fmt"
	"io"
	"net"
)

// EchoServer is an echo implementation of the
// caddy.Server interface type
type EchoServer struct {
	LocalTCPAddr string
	listener     net.Listener
}

// NewEchoServer returns a new echo server
func NewEchoServer(l string) (*EchoServer, error) {
	return &EchoServer{
		LocalTCPAddr: l,
	}, nil
}

// Listen starts listening by creating a new listener
// and returning it. It does not start accepting
// connections.
func (s *EchoServer) Listen() (net.Listener, error) {
	return net.Listen("tcp", fmt.Sprintf("%s", s.LocalTCPAddr))
}

// Serve starts serving using the provided listener.
// Serve blocks indefinitely, or in other
// words, until the server is stopped.
func (s *EchoServer) Serve(ln net.Listener) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}

		fmt.Printf("EchoServer: accepted from %s\n", conn.RemoteAddr())

		go func(c net.Conn) {
			// Echo all incoming data.
			_, err := io.Copy(c, c)
			if err != nil {
				fmt.Printf("io.Copy error: %v\n", err)
			}

			fmt.Println("Closing down connection")
			// Shut down the connection.
			c.Close()
		}(conn)
	}
}

// Stop stops s gracefully (or forcefully after timeout) and
// closes its listener.
func (s *EchoServer) Stop() error {

	fmt.Println("EchoServer Stop")
	err := s.listener.Close()
	if err != nil {
		return err
	}

	return nil
}

// ListenPacket is a no-op to satisfy caddy.Server interface
func (s *EchoServer) ListenPacket() (net.PacketConn, error) { return nil, nil }

// ServePacket is a no-op to satisfy caddy.Server interface
func (s *EchoServer) ServePacket(net.PacketConn) error { return nil }

// OnStartupComplete lists the sites served by this server
// and any relevant information
func (s *EchoServer) OnStartupComplete() {
	fmt.Println("EchoServer OnStartupComplete:", s.LocalTCPAddr)
}
