package main

import (
    "net"
    "crypto/tls"
    "crypto/rand"
)

/* Simple wrapper to Listener that generates
 * GophorConn instances on each accept
 */
type GophorListener struct {
    Listener net.Listener
    Hostname string
    Port     string
}

func BeginGophorListen(bindAddr, hostname, port string) (*GophorListener, error) {
    gophorListener := new(GophorListener)
    gophorListener.Hostname = hostname
    gophorListener.Port = port

    var err error
    gophorListener.Listener, err = net.Listen("tcp", bindAddr+":"+port)
    if err != nil {
        return nil, err
    } else {
        return gophorListener, nil
    }
}

func BeginGophorTlsListen(bindAddr, hostname, port, certFile, keyFile string) (*GophorListener, error) {
    gophorListener := new(GophorListener)
    gophorListener.Hostname = hostname
    gophorListener.Port = port

    /* Try load the key pair */
    cert, err := tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil {
        return nil, err
    }

    /* Setup TLS configuration */
    config := &tls.Config{
        Certificates: []tls.Certificate{ cert },
    }

    /* Use a more cryptographically safe rand source */
    config.Rand = rand.Reader

    gophorListener.Listener, err = tls.Listen("tcp", bindAddr+":"+port, config)
    if err != nil {
        return nil, err
    } else {
        return gophorListener, nil
    }
}

func (l *GophorListener) Accept() (*GophorConn, error) {
    conn, err := l.Listener.Accept()
    if err != nil {
        return nil, err
    }

    gophorConn := new(GophorConn)
    gophorConn.Conn = conn
    gophorConn.Hostname = l.Hostname
    gophorConn.Port = l.Port
    return gophorConn, nil
}

func (l *GophorListener) Addr() net.Addr {
    return l.Listener.Addr()
}

/* Simple wrapper to Conn with easier acccess
 * to hostname / port information
 */
type GophorConn struct {
    Conn     net.Conn
    Hostname string
    Port     string
}

func (c *GophorConn) Read(b []byte) (int, error) {
    return c.Conn.Read(b)
}

func (c *GophorConn) Write(b []byte) (int, error) {
    return c.Conn.Write(b)
}

func (c *GophorConn) RemoteAddr() net.Addr {
    return c.Conn.RemoteAddr()
}

func (c *GophorConn) Close() error {
    return c.Conn.Close()
}
