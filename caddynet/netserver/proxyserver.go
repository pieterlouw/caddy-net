package netserver

import (
	"crypto/tls"
	"fmt"
	"net"

	"github.com/caddyserver/caddy"
	"github.com/caddyserver/caddy/caddytls"
)

// ProxyServer is an implementation of the
// caddy.Server interface type
type ProxyServer struct {
	LocalTCPAddr    string
	DestTCPAddr     string
	tcpListener     net.Listener
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
func (s *ProxyServer) ListenPacket() (net.PacketConn, error) {
	return net.ListenPacket("udp", fmt.Sprintf("%s", s.LocalTCPAddr))

}

// Serve starts serving using the provided listener.
// Serve blocks indefinitely, or in other
// words, until the server is stopped.
func (s *ProxyServer) Serve(ln net.Listener) error {

	s.tcpListener = ln

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

			// wait for data from remote server
			go conn.Wait()
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

		conn, found := s.udpClients[clientAddr]
		if found {
			conn.Close()
			delete(s.udpClients, clientAddr)
		}
	}
}

// Stop stops s gracefully and closes its listener.
func (s *ProxyServer) Stop() error {

	err := s.tcpListener.Close()
	if err != nil {
		return err
	}

	s.udpPacketConn.Close()
	if err != nil {
		return err
	}

	return nil
}

// OnStartupComplete lists the sites served by this server
// and any relevant information
func (s *ProxyServer) OnStartupComplete() {
	if !caddy.Quiet {
		fmt.Println("[INFO] Proxying from ", s.LocalTCPAddr, " -> ", s.DestTCPAddr)
	}
}
