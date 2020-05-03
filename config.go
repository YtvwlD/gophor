package main

import (
    "regexp"
)

/* ServerConfig:
 * Holds onto global server configuration details
 * and any data objects we want to keep in memory
 * (e.g. loggers, restricted files regular expressions
 * and file cache)
 */
type ServerConfig struct {
    /* Executable Settings */
    Env                  []string

    /* Content settings */
    FooterText           []byte
    PageWidth            int

    /* Regex */
    CmdParseLineRegex    *regexp.Regexp
    RestrictedFiles      []*regexp.Regexp

    /* Logging */
    SysLog               LoggerInterface
    AccLog               LoggerInterface

    /* Filesystem access */
    FileSystem           *FileSystem
}
