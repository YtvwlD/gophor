package main

import (
    "regexp"
    "log"
)

/* ServerConfig:
 * Holds onto global server configuration details
 * and any data objects we want to keep in memory
 * (e.g. loggers, restricted files regular expressions
 * and file cache)
 */
type ServerConfig struct {
    /* Base settings */
    RootDir         string

    /* Content settings */
    FooterText      []byte
    PageWidth       int
    RestrictedFiles []*regexp.Regexp

    /* Logging */
    SystemLogger    *log.Logger
    AccessLogger    *log.Logger

    /* Filesystem access */
    FileSystem      *FileSystem
}

func (config *ServerConfig) LogSystem(fmt string, args ...interface{}) {
    config.SystemLogger.Printf(":: I :: "+fmt, args...)
}

func (config *ServerConfig) LogSystemError(fmt string, args ...interface{}) {
    config.SystemLogger.Printf(":: E :: "+fmt, args...)
}

func (config *ServerConfig) LogSystemFatal(fmt string, args ...interface{}) {
    config.SystemLogger.Fatalf(":: F :: "+fmt, args...)
}

func (config *ServerConfig) LogAccess(sourceAddr, fmt string, args ...interface{}) {
    config.AccessLogger.Printf(":: I :: ["+sourceAddr+"] "+fmt, args...)
}

func (config *ServerConfig) LogAccessError(sourceAddr, fmt string, args ...interface{}) {
    config.AccessLogger.Printf(":: E :: ["+sourceAddr+"] "+fmt, args...)
}
