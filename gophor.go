package main

import (
    "fmt"
    "os"
    "syscall"
    "flag"
    "net"
)

/*
 * Global Constants
 */
const (
    ShowHidden = false

    SocketReadBufSize = 1024
    FileReadBufSize = 1024
    MaxSocketReadChunks = 4
)

type Command int
const (
    Stop  Command = iota
    Clean Command = iota
)

/*
 * Gopher server
 */
var (
    /* Run-time arguments */
    ServerRoot     = flag.String("root", "/var/gopher", "server root directory")
    ServerPort     = flag.Int("port", 70, "server listening port")
    ServerHostname = flag.String("hostname", "127.0.0.1", "server hostname")
)

func main() {
    /* Parse run-time arguments */
    flag.Parse()

    /* Enter chroot */
    if enterChroot() != nil {
        os.Exit(1)
    }

    /* Set-up socket */
    listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *ServerHostname, *ServerPort))
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error opening socket on port %d: %v\n", *ServerPort, err)
        os.Exit(1)
    }

    /* Setup client manager */
    manager := new(ClientManager)
    manager.Init()
    manager.Start()

    /* Main server loop */
    count := 0
    for {
        newConn, err := listener.Accept()
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error accepting connection from %s: %v\n", newConn, err)
            continue
        }

        client := new(Client)
        client.Init(&newConn)
        manager.Register<-client

        if count == 5 {
            manager.Cmd<-Clean
            count = 0
        } else {
            count += 1
        }
    }

    /* Handle stopping */
}

func enterChroot() error {
    err := syscall.Chdir(*ServerRoot)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error changing dir to server root %s: %v\n", *ServerRoot, err)
        return err
    }

    err = syscall.Chroot(*ServerRoot)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error chroot'ing into server root %s: %v\n", *ServerRoot, err)
        return err
    }
    
    return nil
}
