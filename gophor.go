package main

import (
    "os"
    "strconv"
    "syscall"
    "os/signal"
    "flag"
    "time"
)

const (
    GophorVersion   = "0.7-beta-PR1"
)

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
            Config.SysLog.Info("", "Listening on: gopher://%s\n", l.Addr())

            for {
                newConn, err := l.Accept()
                if err != nil {
                    Config.SysLog.Error("", "Error accepting connection: %s\n", err.Error())
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
    Config.SysLog.Info("", "Signal received: %v. Shutting down...\n", sig)
    os.Exit(0)
}

func setupServer() []*GophorListener {
    /* First we setup all the flags and parse them... */

    /* Base server settings */
    serverRoot        := flag.String("root-dir", "/var/gopher", "Change server root directory.")
    serverHostname    := flag.String("hostname", "127.0.0.1", "Change server hostname (FQDN).")
    serverPort        := flag.Int("port", 70, "Change server port (0 to disable unencrypted traffic).")
    serverBindAddr    := flag.String("bind-addr", "127.0.0.1", "Change server socket bind address")

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
    logOutput         := flag.String("log-output", "stderr", "Change server log file handling (disable|stderr|file)")
    logOpts           := flag.String("log-opts", "timestamp,ip", "Comma-separated list of log options (timestamp|ip)")

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
    Config.SysLog, Config.AccLog = setupLoggers(*logOutput, *logOpts, *systemLogPath, *accessLogPath)

    /* If running as root, get ready to drop privileges */
    if syscall.Getuid() == 0 || syscall.Getgid() == 0 {
        Config.SysLog.Fatal("", "Gophor does not support running as root!\n")
    }

    /* Enter server dir */
    enterServerDir(*serverRoot)
    Config.SysLog.Info("", "Entered server directory: %s\n", *serverRoot)

    /* Setup initial server shell environment with the info we have to hand */

    /* Setup regular and cgi shell environments */
    Config.Env    = setupExecEnviron()
    Config.CgiEnv = setupInitialCgiEnviron()

    /* Setup listeners */
    listeners := make([]*GophorListener, 0)

    /* If requested, setup unencrypted listener */
    if *serverPort != 0 {
        l, err := BeginGophorListen(*serverBindAddr, *serverHostname, strconv.Itoa(*serverPort), *serverRoot)
        if err != nil {
            Config.SysLog.Fatal("", "Error setting up (unencrypted) listener: %s\n", err.Error())
        }
        listeners = append(listeners, l)
    } else {
        Config.SysLog.Fatal("", "No valid port to listen on\n")
    }

    /* Compile CmdParse regular expression */
    Config.CmdParseLineRegex = compileCmdParseRegex()

    /* Compile user restricted files regex */
    Config.RestrictedFiles = compileUserRestrictedFilesRegex(*restrictedFiles)

    /* Setup file cache */
    Config.FileSystem = new(FileSystem)

    /* Check if cache requested disabled */
    if !*cacheDisabled {
        /* Parse suppled cache check frequency time */
        fileMonitorSleepTime, err := time.ParseDuration(*cacheCheckFreq)
        if err != nil {
            Config.SysLog.Fatal("", "Error parsing supplied cache check frequency %s: %s\n", *cacheCheckFreq, err)
        }

        /* Init file cache */
        Config.FileSystem.Init(*cacheSize, *cacheFileSizeMax)

        /* Before file monitor or any kind of new goroutines started,
         * check if we need to cache generated policy files
         */
        cachePolicyFiles(*serverDescription, *serverAdmin, *serverGeoloc)

        /* Start file cache freshness checker */
        go startFileMonitor(fileMonitorSleepTime)
        Config.SysLog.Info("", "File caching enabled with: maxcount=%d maxsize=%.3fMB checkfreq=%s\n", *cacheSize, *cacheFileSizeMax, fileMonitorSleepTime)
    } else {
        /* File caching disabled, init with zero max size so nothing gets cached */
        Config.FileSystem.Init(2, 0)
        Config.SysLog.Info("", "File caching disabled\n")

        /* Safe to cache policy files now */
        cachePolicyFiles(*serverDescription, *serverAdmin, *serverGeoloc)
    }

    /* Return the created listeners slice :) */
    return listeners
}

func enterServerDir(path string) {
    err := syscall.Chdir(path)
    if err != nil {
        Config.SysLog.Fatal("", "Error changing dir to server root %s: %s\n", path, err.Error())
    }
}
