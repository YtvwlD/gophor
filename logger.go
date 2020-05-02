package main

import (
    "log"
    "os"
    "strings"
)

const (
    /* Prefixes */
    LogPrefixInfo  = ": I :: "
    LogPrefixError = ": E :: "
    LogPrefixFatal = ": F :: "

    /* Log output types */
    LogDisabled = "disable"
    LogToStderr = "stderr"
    LogToFile   = "file"

    /* Log options */
    LogTimestamps = "timestamp"
    LogIps        = "ip"
)

type LoggerInterface interface {
    Info(string, string, ...interface{})
    Error(string, string, ...interface{})
    Fatal(string, string, ...interface{})
}

type NullLogger struct {}
func (l *NullLogger) Info(prefix, format string, args ...interface{}) {}
func (l *NullLogger) Error(prefix, format string, args ...interface{}) {}
func (l *NullLogger) Fatal(prefix, format string, args ...interface{}) {}

type Logger struct {
    Logger *log.Logger
}

func (l *Logger) Info(prefix, format string, args ...interface{}) {
    l.Logger.Printf(LogPrefixInfo+prefix+format, args...)
}

func (l *Logger) Error(prefix, format string, args ...interface{}) {
    l.Logger.Printf(LogPrefixError+prefix+format, args...)
}

func (l *Logger) Fatal(prefix, format string, args ...interface{}) {
    l.Logger.Fatalf(LogPrefixFatal+prefix+format, args...)
}

type LoggerNoPrefix struct {
    Logger *log.Logger
}

func (l *LoggerNoPrefix) Info(prefix, format string, args ...interface{}) {
    /* Ignore the prefix */
    l.Logger.Printf(LogPrefixInfo+format, args...)
}

func (l *LoggerNoPrefix) Error(prefix, format string, args ...interface{}) {
    /* Ignore the prefix */
    l.Logger.Printf(LogPrefixError+format, args...)
}

func (l *LoggerNoPrefix) Fatal(prefix, format string, args ...interface{}) {
    /* Ignore the prefix */
    l.Logger.Fatalf(LogPrefixFatal+format, args...)
}

func setupLoggers(logOutput, logOpts, systemLogPath, accessLogPath string) (LoggerInterface, LoggerInterface) {
    logIps := false
    logFlags := 0
    for _, opt := range strings.Split(logOpts, ",") {
        switch opt {
            case "":
                continue

            case LogTimestamps:
                logFlags = log.LstdFlags

            case LogIps:
                logIps = true

            default:
                log.Fatalf("Unrecognized log opt: %s\n")
        }
    }

    switch logOutput {
        case "":
            /* Assume empty means stderr */
            fallthrough

        case LogToStderr:
            /* Return two separate stderr loggers */
            sysLogger := &LoggerNoPrefix{ NewLoggerToStderr(logFlags) }
            if logIps {
                return sysLogger, &Logger{ NewLoggerToStderr(logFlags) }
            } else {
                return sysLogger, &LoggerNoPrefix{ NewLoggerToStderr(logFlags) }
            }

        case LogDisabled:
            /* Return two pointers to same null logger */
            nullLogger := &NullLogger{}
            return nullLogger, nullLogger

        case LogToFile:
            /* Return two separate file loggers */
            sysLogger := &Logger{ NewLoggerToFile(systemLogPath, logFlags) }
            if logIps {
                return sysLogger, &Logger{ NewLoggerToFile(accessLogPath, logFlags) }
            } else {
                return sysLogger, &LoggerNoPrefix{ NewLoggerToFile(accessLogPath, logFlags) }
            }

        default:
            log.Fatalf("Unrecognised log output type: %s\n", logOutput)
            return nil, nil
    }

}

func NewLoggerToStderr(logFlags int) *log.Logger {
    return log.New(os.Stderr, "", logFlags)
}

func NewLoggerToFile(path string, logFlags int) *log.Logger {
    writer, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
    if err != nil {
        log.Fatalf("Failed to create logger to file %s: %s\n", path, err.Error())
    }
    return log.New(writer, "", logFlags)
}

func printVersionExit() {
    /* Reset the flags before printing version */
    log.SetFlags(0)
    log.Printf("%s\n", GophorVersion)
    os.Exit(0)
}
