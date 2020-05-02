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
    /* Base settings */

    /* Content settings */
    FooterText      []byte
    PageWidth       int
    RestrictedFiles []*regexp.Regexp

    /* Logging */
    SysLog          LoggerInterface
    AccLog          LoggerInterface

    /* Filesystem access */
    FileSystem      *FileSystem
}
