package main

import (
    "log"
    "os"
    "io"
)

var (
    systemLogger *log.Logger
    accessLogger *log.Logger
)

func loggingSetup(systemLogPath, accessLogPath string) {
    /* Setup the system logger:
     * - no prefix
     * - standard flags (time+date prefix) set
     * - output to file path, or os.Stderr default
     */
    var systemWriter io.Writer
    if systemLogPath != "" {
        fd, err := os.OpenFile(systemLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
        if err != nil {
            log.Fatalf("Failed to create system logger: %s\n", err.Error())
        }
        systemWriter = fd        
    } else {
        systemWriter = os.Stderr
    }
    systemLogger = log.New(systemWriter, "", log.LstdFlags)

    /* Setup the access logger:
     * - no prefix
     * - standard flags (time+date prefix) set
     * - output to file path, or os.Stderr default
     */
    var accessWriter io.Writer
    if accessLogPath != "" {
        fd, err := os.OpenFile(accessLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
        if err != nil {
            log.Fatalf("Failed to create access logger: %s\n", err.Error())
        }
        accessWriter = fd
    } else {
        accessWriter = os.Stderr
    }
    accessLogger = log.New(accessWriter, "", log.LstdFlags)
}

func logSystem(fmt string, args ...interface{}) {
    systemLogger.Printf(":: I :: "+fmt, args...)
}

func logSystemError(fmt string, args ...interface{}) {
    systemLogger.Printf(":: E :: "+fmt, args...)
}

func logSystemFatal(fmt string, args ...interface{}) {
    systemLogger.Fatalf(":: F :: "+fmt, args...)
}

func logAccess(fmt string, args ...interface{}) {
    accessLogger.Printf(":: I :: "+fmt, args...)
}

func logAccessError(fmt string, args ...interface{}) {
    accessLogger.Printf(":: E :: "+fmt, args...)
}
