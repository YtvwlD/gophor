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
            Config.LogSystem("Listening on: gopher://%s\n", l.Addr())

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
    serverRoot        := flag.String("root-dir", "/var/gopher", "Change server root directory.")
    serverHostname    := flag.String("hostname", "127.0.0.1", "Change server hostname (FQDN).")
    serverPort        := flag.Int("port", 70, "Change server port (0 to disable unencrypted traffic).")
    serverBindAddr    := flag.String("bind-addr", "127.0.0.1", "Change server socket bind address")
    execAs            := flag.String("user", "", "Drop to supplied user's UID and GID permissions before execution.")
    rootless          := flag.Bool("rootless", false, "Run without root privileges (no chroot, no privilege drop, no restricted port nums).")

    /* User supplied caps.txt information */
    serverDescription := flag.String("description", "Gophor: a Gopher server in GoLang", "Change server description in generated caps.txt.")
    serverAdmin       := flag.String("admin-email", "", "Change admin email in generated caps.txt.")
    serverGeoloc      := flag.String("geoloc", "", "Change server gelocation string in generated caps.txt.")

    /* Content settings */
    footerText        := flag.String("footer", "", "Change gophermap footer text (Unix new-line separated lines).")
    footerSeparator   := flag.Bool("no-footer-separator", false, "Disable footer line separator.")

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
    cacheDisabled     := flag.Bool("disable-cache", false, "Disable file caching.")

    /* Version string */
    version           := flag.Bool("version", false, "Print version information.")

    /* Parse parse parse!! */
    flag.Parse()
    if *version {
        printVersionExit()
    }

    /* Setup the server configuration instance and enter as much as we can right now */
    Config = new(ServerConfig)
    Config.PageWidth   = *pageWidth

    /* Have to be set AFTER page width variable set */
    Config.FooterText  = formatGophermapFooter(*footerText, !*footerSeparator)

    /* Setup Gophor logging system */
    Config.SystemLogger, Config.AccessLogger = setupLogging(*logType, *systemLogPath, *accessLogPath)

    /* If running as root, get ready to drop privileges */
    var uid, gid int
    if !*rootless {
        /* Getting UID+GID for supplied user, has to be done BEFORE chroot */
        if *execAs == "" || *execAs == "root" {
            /* Naughty, naughty! */
            Config.LogSystemFatal("Gophor does not support directly running as root, please supply a non-root user account\n")
        } else {
            /* Try lookup specified username */
            user, err := user.Lookup(*execAs)
            if err != nil {
                Config.LogSystemFatal("Error getting information for requested user %s: %s\n", *execAs, err)
            }

            /* These values should be coming straight out of /etc/passwd, so assume safe */
            uid, _ = strconv.Atoi(user.Uid)
            gid, _ = strconv.Atoi(user.Gid)

            /* Double check this isn't a privileged account */
            if uid == 0 || gid == 0 {
                Config.LogSystemFatal("Gophor does not support running with any kind of privileges, please supply a non-root user account\n")
            }
        }
    } else {
        if syscall.Getuid() == 0 || syscall.Getgid() == 0 {
            Config.LogSystemFatal("Gophor (obviously) does not support running rootless, as root -.-\n")
        } else if *execAs != "" {
            Config.LogSystemFatal("Gophor does not support dropping privileges when running rootless\n")
        }
    }

    /* Enter server dir */
    enterServerDir(*serverRoot)
    Config.LogSystem("Entered server directory: %s\n", *serverRoot)

    /* Try enter chroot if requested */
    if *rootless {
        Config.RootDir = *serverRoot
        Config.LogSystem("Running rootless, server root set: %s\n", *serverRoot)
    } else {
        chrootServerDir(*serverRoot)
        Config.RootDir = "/"
        Config.LogSystem("Chroot success, new root: %s\n", *serverRoot)
    }

    /* Setup listeners */
    listeners := make([]*GophorListener, 0)

    /* If requested, setup unencrypted listener */
    if *serverPort != 0 {
        l, err := BeginGophorListen(*serverBindAddr, *serverHostname, strconv.Itoa(*serverPort))
        if err != nil {
            Config.LogSystemFatal("Error setting up (unencrypted) listener: %s\n", err.Error())
        }
        listeners = append(listeners, l)
    } else {
        Config.LogSystemFatal("No valid port to listen on\n")
    }

    /* Drop not rootless, privileges to retrieved UID+GID */
    if !*rootless {
        setPrivileges(uid, gid)
        Config.LogSystem("Successfully dropped privileges to UID:%d GID:%d\n", uid, gid)
    } else {
        Config.LogSystem("Running as current user\n")
    }

    /* Compile user restricted files regex if supplied */
    if *restrictedFiles != "" {
        Config.RestrictedFiles = compileUserRestrictedFilesRegex(*restrictedFiles)
        Config.LogSystem("Restricted files regular expressions compiled\n")

        /* Setup the listDir function to use regex matching */
        listDir = _listDirRegexMatch
    } else {
        /* Setup the listDir function to skip regex matching */
        listDir = _listDir
    }

    /* Setup file cache */
    Config.FileSystem = new(FileSystem)

    /* Check if cache requested disabled */
    if !*cacheDisabled {
        /* Parse suppled cache check frequency time */
        fileMonitorSleepTime, err := time.ParseDuration(*cacheCheckFreq)
        if err != nil {
            Config.LogSystemFatal("Error parsing supplied cache check frequency %s: %s\n", *cacheCheckFreq, err)
        }

        /* Init file cache */
        Config.FileSystem.Init(*cacheSize, *cacheFileSizeMax)

        /* Before file monitor or any kind of new goroutines started,
         * check if we need to cache generated policy files
         */
        cachePolicyFiles(*serverDescription, *serverAdmin, *serverGeoloc)

        /* Start file cache freshness checker */
        go startFileMonitor(fileMonitorSleepTime)
        Config.LogSystem("File caching enabled with: maxcount=%d maxsize=%.3fMB checkfreq=%s\n", *cacheSize, *cacheFileSizeMax, fileMonitorSleepTime)
    } else {
        /* File caching disabled, init with zero max size so nothing gets cached */
        Config.FileSystem.Init(2, 0)
        Config.LogSystem("File caching disabled\n")

        /* Safe to cache policy files now */
        cachePolicyFiles(*serverDescription, *serverAdmin, *serverGeoloc)
    }

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
    /* C-bind setgid */
    result := C.setgid(C.uint(execGid))
    if result != 0 {
        Config.LogSystemFatal("Failed setting GID %d: %d\n", execGid, result)
    }

    /* C-bind setuid */
    result = C.setuid(C.uint(execUid))
    if result != 0 {
        Config.LogSystemFatal("Failed setting UID %d: %d\n", execUid, result)
    }
}
