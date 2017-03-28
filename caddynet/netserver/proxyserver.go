package netserver

import (
	"crypto/tls"
	"fmt"
	"net"

	"github.com/mholt/caddy/caddytls"
)

// ProxyServer is an implementation of the
// caddy.Server interface type
type ProxyServer struct {
	LocalTCPAddr    string
	listener        net.Listener
	DestTCPAddr     string
	config          *Config
	udpPacketConn   net.PacketConn
	udpClients      map[string]*proxyUDPConnection
	udpClientClosed chan string
}

// NewProxyServer returns a new proxy server
func NewProxyServer(l string, d string, c *Config) (*ProxyServer, error) {
	return &ProxyServer{
		LocalTCPAddr: l,
		DestTCPAddr:  d,
		config:       c,
		udpClients:   make(map[string]*proxyUDPConnection),
	}, nil
}

// Listen starts listening by creating a new listener
// and returning it. It does not start accepting
// connections.
func (s *ProxyServer) Listen() (net.Listener, error) {
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
func (s *ProxyServer) ListenPacket() (net.PacketConn, error) {
	return net.ListenPacket("udp", fmt.Sprintf("%s", s.LocalTCPAddr))

}

// Serve starts serving using the provided listener.
// Serve blocks indefinitely, or in other
// words, until the server is stopped.
func (s *ProxyServer) Serve(ln net.Listener) error {

	s.listener = ln

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

// ServePacket starts serving using the provided listener.
// ServePacket blocks indefinitely, or in other
// words, until the server is stopped.
func (s *ProxyServer) ServePacket(con net.PacketConn) error {

	s.udpPacketConn = con
	s.udpClientClosed = make(chan string)

	go s.handleClosedUDPConnections()

	buf := make([]byte, 4096)
	for {
		nr, addr, err := s.udpPacketConn.ReadFrom(buf)
		if err != nil {
			s.udpPacketConn.Close()
		}

		fmt.Printf("received: [%d] from [%s:%s]\n:[%s]\n", nr, addr.Network(), addr.String(), string(buf[:nr]))

		conn, found := s.udpClients[addr.String()]
		if !found {
			// Get remote server address
			raddr, err := net.ResolveUDPAddr("udp", s.DestTCPAddr)
			if err != nil {
				return err
			}

			remoteUDPConn, err := net.DialUDP("udp", nil, raddr)
			if err != nil {
				return err
			}

			conn = &proxyUDPConnection{
				lconn:     s.udpPacketConn,
				laddr:     addr,
				rconn:     remoteUDPConn,
				closeChan: s.udpClientClosed,
			}

			s.udpClients[addr.String()] = conn
			fmt.Printf("Created new connection for client %s\n", addr.String())

			// wait for data from remote server
			go conn.Wait()
		} else {
			fmt.Printf("Found connection for client %s\n", addr.String())
		}

		// proxy data received to remote server
		_, err = conn.rconn.Write(buf[0:nr])
		if err != nil {
			return err
		}
	}

}

// handleClosedUDPConnections blocks and waits for udp closed connections and do cleanup
func (s *ProxyServer) handleClosedUDPConnections() {
	for {
		clientAddr := <-s.udpClientClosed

		fmt.Printf("Closing UDP connection: [%s]\n", clientAddr)

		conn, found := s.udpClients[clientAddr]
		if found {
			conn.Close()
			delete(s.udpClients, clientAddr)
		}
	}
}

// Stop stops s gracefully and closes its listener.
func (s *ProxyServer) Stop() error {

	fmt.Println("TCP ProxyServer Stop")
	err := s.listener.Close()
	if err != nil {
		return err
	}

	fmt.Println("UDP ProxyServer Stop")
	s.udpPacketConn.Close()
	if err != nil {
		return err
	}

	return nil
}

// OnStartupComplete lists the sites served by this server
// and any relevant information
func (s *ProxyServer) OnStartupComplete() {
	fmt.Println("ProxyServer OnStartupComplete: Proxying from ", s.LocalTCPAddr, " -> ", s.DestTCPAddr)
}
