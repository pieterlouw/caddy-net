package netserver

import (
	"fmt"
	"io"
	"net"
)

// proxyConnection resembles a proxy connection and pipe data between local and remote.
type proxyConnection struct {
	sentBytes     uint64
	receivedBytes uint64
	laddr, raddr  string
	lconn, rconn  net.Conn
	erred         bool
	closeSignal   chan bool
}

// proxy establishes the connection to the remote server and
// starts data exchange. It will block until a close signal is received
// so it's advisable to call as a goroutine
func (p *proxyConnection) proxy() {
	defer p.lconn.Close()
	var err error

	p.rconn, err = net.Dial("tcp", p.raddr)
	if err != nil {
		p.errorFunc("Cannot connect to remote connection: %s", err)
		return
	}
	defer p.rconn.Close()

	go p.exchangeData(p.rconn, p.lconn)
	go p.exchangeData(p.lconn, p.rconn)

	//wait for close signal
	<-p.closeSignal
	fmt.Printf("Done proxying: %s %s\n", p.lconn.LocalAddr(), p.rconn.LocalAddr())
}

// exchangeData reads from source connection and forwards
// data to destination connection
func (p *proxyConnection) exchangeData(dst, src net.Conn) {
	buf := make([]byte, 32*1024) // THIS SHOULD BE CONFIGURABLE
	for {
		bytesRead, err := src.Read(buf)
		if err != nil {
			p.errorFunc(fmt.Sprintf("Error reading from client connection. src=%s %s dst=%s %s", src.LocalAddr(), src.RemoteAddr(), dst.LocalAddr(), dst.RemoteAddr()), err)
			return
		}

		if bytesRead > 0 {
			b := buf[:bytesRead]
			_, err = dst.Write(b)
			if err != nil {
				p.errorFunc("Cannot write to remote connection", err)
				return
			}
		}
	}
}

// errorFunc handles errors and send a close signal
func (p *proxyConnection) errorFunc(s string, err error) {
	if p.erred {
		return
	}
	if err != io.EOF {
		fmt.Printf("[ERROR] %s Err:%s\n", s, err)
	}
	p.closeSignal <- true
	p.erred = true
}

// close sends close signal
func (p *proxyConnection) close() {
	p.closeSignal <- true
	p.erred = true
}
