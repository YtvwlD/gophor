package main

import (
    "net"
    "strconv"
)

type ConnHost struct {
    /* Hold host specific details */

    Name    string
    Port    string
    RootDir string
}

type ConnClient struct {
    /* Hold client specific details */
    Ip   string
    Port string
}

type GophorListener struct {
    /* Simple net.Listener wrapper that holds onto virtual
     * host information + generates GophorConn instances
     */

    Listener net.Listener
    Host     *ConnHost
}

func BeginGophorListen(bindAddr, hostname, port, rootDir string) (*GophorListener, error) {
    gophorListener := new(GophorListener)
    gophorListener.Host = &ConnHost{ hostname, port, rootDir }

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

    /* Copy over listener host */
    gophorConn.Host = l.Host

    /* Should always be ok as listener is type TCP (see above) */
    addr, _ := conn.RemoteAddr().(*net.TCPAddr)
    gophorConn.Client = &ConnClient{
        addr.IP.String(),
        strconv.Itoa(addr.Port),
    }

    return gophorConn, nil
}

func (l *GophorListener) Addr() net.Addr {
    return l.Listener.Addr()
}

type GophorConn struct {
    /* Simple net.Conn wrapper with virtual host and client info */

    Conn   net.Conn
    Host   *ConnHost
    Client *ConnClient
}

func (c *GophorConn) Read(b []byte) (int, error) {
    return c.Conn.Read(b)
}

func (c *GophorConn) Write(b []byte) (int, error) {
    return c.Conn.Write(b)
}

func (c *GophorConn) Close() error {
    return c.Conn.Close()
}
