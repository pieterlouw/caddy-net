package netserver

import "net"

// proxyUDPConnection resembles a UDP proxy connection and pipe data between local and remote.
type proxyUDPConnection struct {
	lconn     net.PacketConn
	laddr     net.Addr     // Address of the client
	rconn     *net.UDPConn // UDP connection to remote server
	closeChan chan string
}

// Wait reads packets from remote server and forwards it on to the client connection
func (p *proxyUDPConnection) Wait() {
	buf := make([]byte, 32*1024) // THIS SHOULD BE CONFIGURABLE
	for {
		// Read from server
		n, err := p.rconn.Read(buf)
		if err != nil {
			p.closeChan <- p.laddr.String()
			return
		}
		// Relay data from remote back to client
		_, err = p.lconn.WriteTo(buf[0:n], p.laddr)
		if err != nil {
			p.closeChan <- p.laddr.String()
			return
		}
	}
}

func (p *proxyUDPConnection) Close() {
	p.rconn.Close()
}
