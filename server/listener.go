package server

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"time"
)

// https://groups.google.com/forum/#!topic/golang-nuts/4oZp1csAm2o

type Conn struct {
	net.Conn
	buf []byte
	err error
}

func (c *Conn) Read(b []byte) (int, error) {
	if len(c.buf) == 0 {
		return c.Conn.Read(b)
	}
	b[0] = c.buf[0]
	c.buf = nil
	if len(b) > 1 && c.err == nil {
		n, e := c.Conn.Read(b[1:])
		if e != nil {
			c.Close()
		}
		return n + 1, e
	}
	return 1, c.err
}

type splitListener struct {
	net.Listener
	config *tls.Config
}

func (l splitListener) Accept() (net.Conn, error) {
	con, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	b := make([]byte, 1)
	_, err = con.Read(b)
	if err != nil {
		con.Close()
		if err != io.EOF {
			return nil, err
		}
	}

	con = &Conn{con, b, err}
	if b[0] == 22 {
		log.Println("https")
		con = tls.Server(con, l.config)
	}
	return con, nil
}

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}
func NewListener(network, addr string, config *tls.Config) (net.Listener, error) {
	ln, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}
	ln = tcpKeepAliveListener{ln.(*net.TCPListener)}
	if config == nil {
		return ln, nil
	}
	return splitListener{ln, config}, nil
}
