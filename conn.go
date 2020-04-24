package main

import (
    "net"
)

/* Data structure to hold specific host details */
type ConnHost struct {
    Name string
    Port string
}

/* Simple wrapper to Listener that holds onto virtual
 * host information and generates GophorConn
 * instances on each accept
 */
type GophorListener struct {
    Listener net.Listener
    Host     *ConnHost
}

func BeginGophorListen(bindAddr, hostname, port string) (*GophorListener, error) {
    gophorListener := new(GophorListener)
    gophorListener.Host = &ConnHost{ hostname, port }

    var err error
    gophorListener.Listener, err = net.Listen("tcp", bindAddr+":"+port)
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
    gophorConn.Host = &ConnHost{ l.Host.Name, l.Host.Port }
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
    Host     *ConnHost
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
