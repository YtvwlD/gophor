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
 * Gopher server
 */
var (
    ServerRoot     = flag.String("root", "/var/gopher", "Change server root directory.")
    ServerHostname = flag.String("hostname", "127.0.0.1", "Change server hostname (FQDN).")
    ServerPort     = flag.Int("port", 70, "Change server port (0 to disable unencrypted traffic).")
//    ServerTlsPort  = flag.Int("tls-port", 0, "Change server TLS/SSL port (0 to disable).")
//    ServerTlsCert  = flag.String("cert", "", "Change server TLS/SSL cert file.")
    ExecAsUid      = flag.Int("uid", 1000, "Change UID to drop executable privileges to.")
    ExecAsGid      = flag.Int("gid", 100, "Change GID to drop executable privileges to.")
    SystemLog      = flag.String("system-log", "", "Change server system log file (blank outputs to stderr).")
    AccessLog      = flag.String("access-log", "", "Change server access log file (blank outputs to stderr).")
    LoggingType    = flag.Int("log-type", 0, "Change server log file handling -- 0:default 1:disable")

    /* FileCaches */
    GophermapCache *GophermapFileCache
    RegularCache   *RegularFileCache
)

func main() {
    /* Setup global logger */
    log.SetOutput(os.Stderr)
    log.SetFlags(0)

    /* Parse run-time arguments */
    flag.Parse()

    /* Setup _OUR_ loggers */
    loggingSetup()

    /* Enter server dir */
    enterServerDir()
    logSystem("Entered server directory: %s\n", *ServerRoot)

    /* Try enter chroot if requested */
    chrootServerDir()
    logSystem("Chroot success, new root: %s\n", *ServerRoot)

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

    /* Create file caches */
    GophermapCache = new(GophermapFileCache)
    GophermapCache.Init(5)
    RegularCache = new(RegularFileCache)
    RegularCache.Init(5)

    /* Start file cache monitor */
    StartFileMonitor(RegularCache, GophermapCache)

    /* Serve unencrypted traffic */
    go func() {
        for {
            newConn, err := listener.Accept()
            if err != nil {
                logSystemError("Error accepting connection: %s\n", err.Error())
                continue
            }

            /* Run this in it's own goroutine so we can go straight back to accepting */
            go func() {
                w := NewWorker(&newConn)
                w.Serve()
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
}

func setPrivileges() {
    /* Check root privileges aren't being requested */
    if *ExecAsUid == 0 || *ExecAsGid == 0 {
        logSystemFatal("Gophor does not support directly running as root\n")
    }

    /* Get currently running user info */
    uid, gid := syscall.Getuid(), syscall.Getgid()

    /* Set GID if necessary */
    if gid != *ExecAsGid || gid == 0 {
        /* C-bind setgid */
        result := C.setgid(C.uint(*ExecAsGid))
        if result != 0 {
            logSystemFatal("Failed setting GID %d: %d\n", *ExecAsGid, result)
        }
        logSystem("Dropping to GID: %d\n", *ExecAsGid)
    }

    /* Set UID if necessary */
    if uid != *ExecAsUid || uid == 0 {
        /* C-bind setuid */
        result := C.setuid(C.uint(*ExecAsUid))
        if result != 0 {
            logSystemFatal("Failed setting UID %d: %d\n", *ExecAsUid, result)
        }
        logSystem("Dropping to UID: %d\n", *ExecAsUid)
    }
}
