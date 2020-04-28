package main

import (
    "log"
    "os"
    "io"
    "io/ioutil"
)

func setupLogging(loggingType int, systemLogPath, accessLogPath string) (*log.Logger, *log.Logger) {
    /* Setup global logger */
    log.SetOutput(os.Stderr)
    log.SetFlags(0)

    /* Calculate now, because, *shrug* */
    useSame := (systemLogPath == accessLogPath)

    /* Check requested logging type */
    var systemLogger, accessLogger *log.Logger
    switch loggingType {
        case 0:
            /* Default */

            /* Setup system logger to output to file, or stderr if none supplied */
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

            /* If both output to same, may as well use same logger for both */
            if useSame {
                accessLogger = systemLogger
            }

            /* Setup access logger to output to file, or stderr if none supplied */
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

        case 1:
            /* Disable -- pipe logs to "discard". May as well use same for both */
            systemLogger = log.New(ioutil.Discard, "", 0)
            accessLogger = systemLogger

        default:
            log.Fatalf("Unrecognized logging type: %d\n", loggingType)
    }

    return systemLogger, accessLogger
}

func printVersionExit() {
    /* Reset the flags before printing version */
    log.SetFlags(0)
    log.Printf("%s\n", GophorVersion)
    os.Exit(0)
}
