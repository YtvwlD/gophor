package main

import (
    "os"
    "os/user"
    "strconv"
    "syscall"
    "os/signal"
    "flag"
    "time"
)

/*
 * GoLang's built-in syscall.{Setuid,Setgid}() methods don't work as expected (all I ever
 * run into is 'operation not supported'). Which from reading seems to be a result of Linux
 * not always performing setuid/setgid constistent with the Unix expected result. Then mix
 * that with GoLang's goroutines acting like threads but not quite the same... I can see
 * why they're not fully supported.
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

var (
    Config *ServerConfig
)

func main() {
    /* Setup the entire server, getting slice of listeners in return */
    listeners := setupServer()

    /* Handle signals so we can _actually_ shutdowm */
    signals := make(chan os.Signal)
    signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

    /* Start accepting connections on any supplied listeners */
    for _, l := range listeners {
        go func() {
            for {
                newConn, err := l.Accept()
                if err != nil {
                    Config.LogSystemError("Error accepting connection: %s\n", err.Error())
                    continue
                }

                /* Run this in it's own goroutine so we can go straight back to accepting */
                go func() {
                    NewWorker(newConn).Serve()
                }()
            }
        }()
    }

    /* When OS signal received, we close-up */
    sig := <-signals
    Config.LogSystem("Signal received: %v. Shutting down...\n", sig)
    os.Exit(0)
}

func setupServer() []*GophorListener {
    /* First we setup all the flags and parse them... */

    /* Base server settings */
    serverRoot        := flag.String("root", "/var/gopher", "Change server root directory.")
    serverHostname    := flag.String("hostname", "127.0.0.1", "Change server hostname (FQDN).")
    serverPort        := flag.Int("port", 70, "Change server port (0 to disable unencrypted traffic).")
    serverBindAddr    := flag.String("bind-addr", "127.0.0.1", "Change server socket bind address")
    execAs            := flag.String("user", "", "Drop to supplied user's UID and GID permissions before execution.")

    /* User supplied caps.txt information */
    serverDescription := flag.String("description", "Gophor: a Gopher server in GoLang", "Change server description in auto-generated caps.txt.")
    serverAdmin       := flag.String("admin-email", "", "Change admin email in auto-generated caps.txt.")
    serverGeoloc      := flag.String("geoloc", "", "Change server gelocation string in auto-generated caps.txt.")

    /* Content settings */
    pageWidth         := flag.Int("page-width", 80, "Change page width used when formatting output.")
    restrictedFiles   := flag.String("restrict-files", "", "New-line separated list of regex statements restricting files from showing in directory listings.")

    /* Logging settings */
    systemLogPath     := flag.String("system-log", "", "Change server system log file (blank outputs to stderr).")
    accessLogPath     := flag.String("access-log", "", "Change server access log file (blank outputs to stderr).")
    logType           := flag.Int("log-type", 0, "Change server log file handling -- 0:default 1:disable")

    /* Cache settings */
    cacheCheckFreq    := flag.String("cache-check", "60s", "Change file cache freshness check frequency.")
    cacheSize         := flag.Int("cache-size", 50, "Change file cache size, measured in file count.")
    cacheFileSizeMax  := flag.Float64("cache-file-max", 0.5, "Change maximum file size to be cached (in megabytes).")

    /* Parse parse parse!! */
    flag.Parse()

    /* Setup the server configuration instance and enter as much as we can right now */
    Config = new(ServerConfig)
    Config.RootDir     = *serverRoot
    Config.Description = *serverDescription
    Config.AdminEmail  = *serverAdmin
    Config.Geolocation = *serverGeoloc
    Config.PageWidth   = *pageWidth

    /* Setup Gophor logging system */
    Config.SystemLogger, Config.AccessLogger = setupLogging(*logType, *systemLogPath, *accessLogPath)

    /* Get UID + GID for requested user. Has to be done BEFORE chroot or it fails */
    var uid, gid int
    if *execAs == "" {
        /* No 'execAs' user specified, try run as default user account permissions */
        uid = 1000
        gid = 1000
    } else if *execAs == "root" {
        /* Naughty, naughty! */
        Config.LogSystemFatal("Gophor does not support directly running as root\n")
    } else {
        /* Try lookup specified username */
        user, err := user.Lookup(*execAs)
        if err != nil {
            Config.LogSystemFatal("Error getting information for requested user %s: %s\n", *execAs, err)
        }

        /* These values should be coming straight out of /etc/passwd, so assume safe */
        uid, _ = strconv.Atoi(user.Uid)
        gid, _ = strconv.Atoi(user.Gid)
    }

    /* Enter server dir */
    enterServerDir(*serverRoot)
    Config.LogSystem("Entered server directory: %s\n", *serverRoot)

    /* Try enter chroot if requested */
    chrootServerDir(*serverRoot)
    Config.LogSystem("Chroot success, new root: %s\n", *serverRoot)

    /* Setup listeners */
    listeners := make([]*GophorListener, 0)

    /* If requested, setup unencrypted listener */
    if *serverPort != 0 {
        l, err := BeginGophorListen(*serverBindAddr, *serverHostname, strconv.Itoa(*serverPort))
        if err != nil {
            Config.LogSystemFatal("Error setting up (unencrypted) listener: %s\n", err.Error())
        }
        Config.LogSystem("Listening (unencrypted): gopher://%s\n", l.Addr())
        listeners = append(listeners, l)
    } else {
        Config.LogSystemFatal("No valid port to listen on :(\n")
    }

    /* Drop privileges to retrieved UID + GID */
    setPrivileges(uid, gid)
    Config.LogSystem("Successfully dropped privileges to UID:%d GID:%d\n", uid, gid)

    /* Compile user restricted files regex if supplied */
    if *restrictedFiles != "" {
        Config.RestrictedFiles = compileUserRestrictedFilesRegex(*restrictedFiles)

        /* Setup the listDir function to use regex matching */
        listDir = _listDirRegexMatch
    } else {
        /* Setup the listDir function to skip regex matching */
        listDir = _listDir
    }

    /* Parse suppled cache check frequency time */
    fileMonitorSleepTime, err := time.ParseDuration(*cacheCheckFreq)
    if err != nil {
        Config.LogSystemFatal("Error parsing supplied cache check frequency %s: %s\n", *cacheCheckFreq, err)
    }

    /* Setup file cache */
    Config.FileCache = new(FileCache)
    Config.FileCache.Init(*cacheSize, *cacheFileSizeMax)

    /* Start file cache freshness checker */
    go startFileMonitor(fileMonitorSleepTime)

    /* Return the created listeners slice :) */
    return listeners
}

func enterServerDir(path string) {
    err := syscall.Chdir(path)
    if err != nil {
        Config.LogSystemFatal("Error changing dir to server root %s: %s\n", path, err.Error())
    }
}

func chrootServerDir(path string) {
    err := syscall.Chroot(path)
    if err != nil {
        Config.LogSystemFatal("Error chroot'ing into server root %s: %s\n", path, err.Error())
    }

    /* Change to server root just to ensure we're sitting at root of chroot */
    err = syscall.Chdir("/")
    if err != nil {
        Config.LogSystemFatal("Error changing to root of chroot dir: %s\n", err.Error())
    }
}

func setPrivileges(execUid, execGid int) {
    /* Check root privileges aren't being requested */
    if execUid == 0 || execGid == 0 {
        Config.LogSystemFatal("Gophor does not support directly running as root\n")
    }

    /* Get currently running user info */
    uid, gid := syscall.Getuid(), syscall.Getgid()

    /* Set GID if necessary */
    if gid != execUid {
        /* C-bind setgid */
        result := C.setgid(C.uint(execGid))
        if result != 0 {
            Config.LogSystemFatal("Failed setting GID %d: %d\n", execGid, result)
        }
    }

    /* Set UID if necessary */
    if uid != execGid {
        /* C-bind setuid */
        result := C.setuid(C.uint(execUid))
        if result != 0 {
            Config.LogSystemFatal("Failed setting UID %d: %d\n", execUid, result)
        }
    }
}
