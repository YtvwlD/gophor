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
    ServerGid      = flag.Int("gid", 100, "GID to run server under")
    NoChroot       = flag.Bool("no-chroot", false, "don't chroot into the server directory")
    SystemLogPath  = flag.String("system-log", "", "system log path")
    AccessLogPath  = flag.String("access-log", "", "access log path")
)

func main() {
    /* Setup global logger */
    log.SetOutput(os.Stderr)
    log.SetFlags(0)

    /* Parse run-time arguments */
    flag.Parse()
    if flag.NFlag() == 0 {
        log.Printf("Usage: gophor ...\n")
        flag.PrintDefaults()
        os.Exit(1)
    }

    /* Setup _OUR_ loggers */
    loggingSetup(*SystemLogPath, *AccessLogPath)

    /* Enter server dir */
    enterServerDir()
    logSystem("Entered server directory: %s\n", *ServerRoot)

    /* Try enter chroot if requested */
    if !*NoChroot {
        chrootServerDir()
        logSystem("Chroot success, new root: %s\n", *ServerRoot)
    }

    /* Set-up socket while we still have privileges (if held) */
    listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *ServerHostname, *ServerPort))
    if err != nil {
        logSystemFatal("Error opening socket on port %d: %s\n", *ServerPort, err.Error())
    }
    defer listener.Close()
    logSystem("Listening: gopher://%s\n", listener.Addr())

    /* Set privileges, see function definition for better explanation */
    setPrivileges()

    /* Handle signals so we can _actually_ shutdowm */
    signals := make(chan os.Signal)
    signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

    /* listener.Accept() server loop, in its own go-routine */
    go func() {
        for {
            newConn, err := listener.Accept()
            if err != nil {
                logSystemError("Error accepting connection: %s\n", err.Error())
                continue
            }

            go func() {
                /* Create Client from newConn and register with the manager */
                client := new(Client)
                client.Init(&newConn)
                client.Start()
            }()
        }
    }()

    /* When OS signal received, we close-up */
    sig := <-signals
    logSystem("Signal received: %v. Shutting down...\n", sig)
    os.Exit(0)
}

func enterServerDir() {
    err := syscall.Chdir(*ServerRoot)
    if err != nil {
        logSystemFatal("Error changing dir to server root %s: %s\n", *ServerRoot, err.Error())
    }
}

func chrootServerDir() {
    err := syscall.Chroot(*ServerRoot)
    if err != nil {
        logSystemFatal("Error chroot'ing into server root %s: %s\n", *ServerRoot, err.Error())
    }
    ServerDir = "/"
}

func setPrivileges() {
    /* Check root privileges aren't being requested */
    if *ServerUid == 0 || *ServerGid == 0 {
        logSystemFatal("Gophor does not support directly running as root\n")
    }

    /* Get currently running user info */
    uid, gid := syscall.Getuid(), syscall.Getgid()

    /* Set GID if necessary */
    if gid != *ServerGid || gid == 0 {
        /* C-bind setgid */
        result := C.setgid(C.uint(*ServerGid))
        if result != 0 {
            logSystemFatal("Failed setting GID %d: %d\n", *ServerGid, result)
        }
    }

    /* Set UID if necessary */
    if uid != *ServerUid || uid == 0 {
        /* C-bind setuid */
        result := C.setuid(C.uint(*ServerUid))
        if result != 0 {
            logSystemFatal("Failed setting UID %d: %d\n", *ServerUid, result)
        }
    }
}
