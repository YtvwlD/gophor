package main

import (
    "log"
    "os"
    "io"
    "io/ioutil"
)

var (
    systemLogger *log.Logger
    accessLogger *log.Logger
)

func loggingSetup() {
    useSame     := (*SystemLog == *AccessLog)

    /* Check requested logging type */
    switch *LoggingType {
        case 0:
            /* Default */

            /* Setup system logger to output to file, or stderr if none supplied */
            var systemWriter io.Writer
            if *SystemLog != "" {
                fd, err := os.OpenFile(*SystemLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
                if err != nil {
                    log.Fatalf("Failed to create system logger: %s\n", err.Error())
                }
                systemWriter = fd        
            } else {
                systemWriter = os.Stderr
            }
            systemLogger = log.New(systemWriter, "", log.LstdFlags)

            /* If both output to same, may as well use same logger for both */
            if useSame {
                accessLogger = systemLogger
                return
            }

            /* Setup access logger to output to file, or stderr if none supplied */
            var accessWriter io.Writer
            if *AccessLog != "" {
                fd, err := os.OpenFile(*AccessLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
                if err != nil {
                    log.Fatalf("Failed to create access logger: %s\n", err.Error())
                }
                accessWriter = fd
            } else {
                accessWriter = os.Stderr
            }
            accessLogger = log.New(accessWriter, "", log.LstdFlags)

        case 1:
            /* Disable -- pipe logs to "discard". May as well use same for both */
            systemLogger = log.New(ioutil.Discard, "", 0)
            accessLogger = systemLogger
            return

        default:
            log.Fatalf("Unrecognized logging type: %d\n", *LoggingType)
    }
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
