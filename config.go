package main

import (
    "regexp"
    "log"
)

type ServerConfig struct {
    /* Base settings */
    RootDir         string
    Hostname        string
    Port            string

    /* Caps.txt information */
    Description     string
    AdminEmail      string
    Geolocation     string

    /* Content settings */
    PageWidth       int
    RestrictedFiles []*regexp.Regexp

    /* Logging */
    SystemLogger    *log.Logger
    AccessLogger    *log.Logger

    /* Cache */
    FileCache       *FileCache
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
    config.AccessLogger.Printf(":: I :: ("+sourceAddr+") "+fmt, args...)
}

func (config *ServerConfig) LogAccessError(sourceAddr, fmt string, args ...interface{}) {
    config.AccessLogger.Printf(":: E :: ("+sourceAddr+") "+fmt, args...)
}
