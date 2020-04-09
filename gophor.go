package main

import (
    "log"
    "fmt"
    "os"
    "syscall"
    "os/signal"
    "flag"
    "net"
)

/*
 * Global Constants
 */
const (
    ShowHidden = false

    NewConnsBeforeClean = 5
    SocketReadBufSize = 512
    FileReadBufSize = 512
    MaxSocketReadChunks = 8
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

    /* Setup logger */
    log.SetOutput(os.Stderr)
    log.SetFlags(0)

    /* Enter chroot */
    if enterChroot() != nil {
        os.Exit(1)
    }

    /* Set-up socket */
    listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *ServerHostname, *ServerPort))
    if err != nil {
        log.Printf("Error opening socket on port %d: %v\n", *ServerPort, err)
        os.Exit(1)
    }

    /* Setup client manager */
    manager := new(ClientManager)
    manager.Init()
    manager.Start()

    /* Handle signals so we can _actually_ shutdown */
    signals := make(chan os.Signal)
    signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

    /* Listener accept channel */
    accept := make(chan bool)

    /* Main server loop */
    count := 0
    for {
        var newConn net.Conn
        var err error

        /* By default, listener.Accept() IS blocking, but can't be used in a select case. This gets around that */
        go func() {
            newConn, err = listener.Accept()
            accept<-true // signifies listened, newConn is ready
        }()

        /* Block on waiting for new connection, or previously requested OS signal */
        select {
            case _ = <-accept:
                if err != nil {
                    log.Printf("Error accepting connection from %s: %v\n", newConn, err)
                    continue
                }

                /* Add new client based on newConn to be managed */
                client := new(Client)
                client.Init(&newConn)
                manager.Register<-client

                /* Every __ new connections, request manager perform clean */
                if count == NewConnsBeforeClean {
                    manager.Cmd<-Clean
                    count = 0
                } else {
                    count += 1
                }

            case sig := <-signals:
                log.Printf("Received signal %v, waiting to finish up... (hit CTRL-C to terminate without cleanup)\n", sig)
                for {
                    select {
                        case manager.Cmd<-Stop:
                            /* wait on clean */
                            log.Printf("Clean-up finished, exiting now...\n")
                            os.Exit(0)

                        case sig = <-signals:
                            if sig == syscall.SIGTERM {
                                log.Printf("Stopping NOW.\n")
                                os.Exit(1)
                            }
                            continue
                    }
                }
            }
        }
    }

func enterChroot() error {
    err := syscall.Chdir(*ServerRoot)
    if err != nil {
        log.Printf("Error changing dir to server root %s: %v\n", *ServerRoot, err)
        return err
    }

    err = syscall.Chroot(*ServerRoot)
    if err != nil {
        log.Printf("Error chroot'ing into server root %s: %v\n", *ServerRoot, err)
        return err
    }
    
    return nil
}
