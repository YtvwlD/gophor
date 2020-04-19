package main

import (
    "flag"
)

var (
    /* Base server settings */
    ServerRoot        = flag.String("root", "/var/gopher", "Change server root directory.")
    ServerHostname    = flag.String("hostname", "127.0.0.1", "Change server hostname (FQDN).")
    ServerPort        = flag.Int("port", 70, "Change server port (0 to disable unencrypted traffic).")
    ServerBindAddr    = flag.String("bind-addr", "127.0.0.1", "Change server socket bind address")
//    ServerTlsPort  = flag.Int("tls-port", 0, "Change server TLS/SSL port (0 to disable).")
//    ServerTlsCert  = flag.String("cert", "", "Change server TLS/SSL cert file.")
    ExecAsUid         = flag.Int("uid", 1000, "Change UID to drop executable privileges to.")
    ExecAsGid         = flag.Int("gid", 100, "Change GID to drop executable privileges to.")

    /* User supplied caps.txt information */
    ServerDescription = flag.String("description", "Gophor: a Gopher server in GoLang", "Change server description in auto-generated caps.txt.")
    ServerAdmin       = flag.String("admin-email", "", "Change admin email in auto-generated caps.txt.")
    ServerGeoloc      = flag.String("geoloc", "", "Change server gelocation string in auto-generated caps.txt.")

    /* Content settings */
    PageWidth         = flag.Int("page-width", 80, "Change page width used when formatting output.")
    RestrictedFiles   = flag.String("restrict-files", "", "New-line separated list of regex statements restricting files from showing in directory listings.")

    /* Logging settings */
    SystemLog         = flag.String("system-log", "", "Change server system log file (blank outputs to stderr).")
    AccessLog         = flag.String("access-log", "", "Change server access log file (blank outputs to stderr).")
    LoggingType       = flag.Int("log-type", 0, "Change server log file handling -- 0:default 1:disable")

    /* Cache settings */
    CacheCheckFreq    = flag.String("cache-check", "30s", "Change file cache freshness check frequency.")
    CacheSize         = flag.Int("cache-size", 1000, "Change file cache size, measured in file count.")
    CacheFileSizeMax  = flag.Float64("cache-file-max", 5, "Change maximum file size to be cached (in megabytes).")
)
