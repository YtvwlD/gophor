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
 * Here begins hacky fun-time. GoLang's built-in syscall.{Setuid,Setgid}() methods don't
 * work as expected (all I ever run into is 'operation not supported'). Which from reading
 * seems to be a result of Linux not always performing setuid/setgid constistent with the
 * Unix expected result. Then mix that with GoLang's goroutines acting like threads but
 * not quite the same... I can see why they're not fully supported.
 *
 * Instead we're going to take C-bindings and call them directly ourselves, BEFORE spawning
 * any goroutines to prevent fuckery.
 *
 * Oh god here we go...
 */
 
/*
#include <unistd.h>
*/
import "C"

/*
 * Global Constants
 */
const (
    ShowHidden = false

    NewConnsBeforeClean = 5
    SocketReadBufSize = 512
    FileReadBufSize = 512
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
    ServerDir = ""

    /* Run-time arguments */
    ServerRoot     = flag.String("root", "/var/gopher", "server root directory")
    ServerPort     = flag.Int("port", 70, "server listening port")
    ServerHostname = flag.String("hostname", "127.0.0.1", "server hostname")
    ServerUid      = flag.Int("uid", 1000, "UID to run server under")
    ServerGid      = flag.Int("gid", 1000, "GID to run server under")
    UseChroot      = flag.Bool("use-chroot", true, "chroot into the server directory")
)

func main() {
    /* Parse run-time arguments */
    flag.Parse()

    /* Setup logger */
    log.SetOutput(os.Stderr)
    log.SetFlags(0)

    /* Enter server dir */
    enterServerDir()

    /* Try enter chroot if requested */
    if *UseChroot {
        chrootServerDir()
    }

    /* Set-up socket while we still have privileges (if held) */
    listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *ServerHostname, *ServerPort))
    if err != nil {
        log.Fatalf("Error opening socket on port %d: %s\n", *ServerPort, err.Error())
    }
    defer listener.Close()

    /* Set privileges, see function definition for better explanation */
    setPrivileges()

    /* Setup client manager */
    manager := new(ClientManager)
    manager.Init()
    manager.Start()

    /* Handle signals so we can _actually_ shutdown */
    signals := make(chan os.Signal)
    signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

    /* listener.Accept() server loop, in its own go-routine */
    go func() {
        count := 0
        for {
            newConn, err := listener.Accept()
            if err != nil {
                log.Fatalf("Error accepting connection: %s\n", err.Error())
            }

            /* Create Client from newConn and register with the manager */
            client := new(Client)
            client.Init(&newConn)
            manager.Register<-client // this starts the Client go-routine

            /* Every 'NewConnsBeforeClean' connections, request that manager perform clean */
            if count == NewConnsBeforeClean {
                manager.Cmd<-Clean
                count = 0
            } else {
                count += 1
            }
        }
    }()

    /* Main thread sits and listens for OS signals */
    for {
        sig := <-signals
        log.Printf("Received signal %v, waiting to finish up... (hit CTRL-C to terminate without cleanup)\n", sig)

        for {
            select {
                case manager.Cmd<-Stop:
                    /* wait on clean */
                    log.Printf("Clean-up finished, exiting now...\n")
                    os.Exit(0)

                case sig = <-signals:
                    if sig == syscall.SIGTERM {
                        log.Fatalf("Stopping NOW.\n")
                    }
                    continue
            }
        }
    }
}

func enterServerDir() {
    err := syscall.Chdir(*ServerRoot)
    if err != nil {
        log.Fatalf("Error changing dir to server root %s: %s\n", *ServerRoot, err.Error())
    }
}

func chrootServerDir() {
    err := syscall.Chroot(*ServerRoot)
    if err != nil {
        log.Fatalf("Error chroot'ing into server root %s: %s\n", *ServerRoot, err.Error())
    }
}

func setPrivileges() {
    /* Check root privileges aren't being requested */
    if *ServerUid == 0 || *ServerGid == 0 {
        log.Fatalf("Gophor does not support directly running as root\n")
    }

    /* Get currently running user info */
    uid, gid := syscall.Getuid(), syscall.Getgid()

    /* Set GID if necessary */
    if gid != *ServerGid || gid == 0 {
        /* C-bind setgid */
        result := C.setgid(C.uint(*ServerGid))
        if result != 0 {
            log.Fatalf("Failed setting GID %d: %d\n", *ServerGid, result)
        }
    }

    /* Set UID if necessary */
    if uid != *ServerUid || uid == 0 {
        /* C-bind setuid */
        result := C.setuid(C.uint(*ServerUid))
        if result != 0 {
            log.Fatalf("Failed setting UID %d: %d\n", *ServerUid, result)
        }
    }
}
