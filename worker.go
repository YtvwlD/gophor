package main

import (
    "strings"
)

const (
    /* Socket settings */
    SocketReadBufSize = 256 /* Supplied selector should be <= this len */
    MaxSocketReadChunks = 1
)

type Worker struct {
    Conn *GophorConn
}

func NewWorker(conn *GophorConn) *Worker {
    return &Worker{ conn }
}

func (worker *Worker) Serve() {
    defer func() {
        /* Close-up shop */
        worker.Conn.Close()
    }()

    var count int
    var err error

    /* Read buffer + final result */
    buf := make([]byte, SocketReadBufSize)
    received := make([]byte, 0)

    iter := 0
    for {
        /* Buffered read from listener */
        count, err = worker.Conn.Read(buf)
        if err != nil {
            Config.SysLog.Error("", "Error reading from socket on port %s: %s\n", worker.Conn.Host.Port, err.Error())
            return
        }

        /* Only copy non-null bytes */
        received = append(received, buf[:count]...)

        /* If count is less than expected read size, we've hit EOF */
        if count < SocketReadBufSize {
            /* EOF */
            break
        }

        /* Hit max read chunk size, send error + close connection */
        if iter == MaxSocketReadChunks {
            Config.SysLog.Error("", "Reached max socket read size %d. Closing connection...\n", MaxSocketReadChunks*SocketReadBufSize)
            return
        }

        /* Keep count :) */
        iter += 1
    }

    /* Handle request */
    gophorErr := worker.RespondGopher(received)

    /* Handle any error */
    if gophorErr != nil {
        /* Generate response bytes from error code */
        response := generateGopherErrorResponseFromCode(gophorErr.Code)

        /* If we got response bytes to send? SEND 'EM! */
        if response != nil {
            /* No gods. No masters. We don't care about error checking here */
            worker.Send(response)
        }
    }
}

func (worker *Worker) Log(format string, args ...interface{}) {
    Config.AccLog.Info("("+worker.Conn.RemoteAddr().String()+") ", format, args...)
}

func (worker *Worker) LogError(format string, args ...interface{}) {
    Config.AccLog.Error("("+worker.Conn.RemoteAddr().String()+") ", format, args...)
}

func (worker *Worker) Send(b []byte) *GophorError {
    count, err := worker.Conn.Write(b)
    if err != nil {
        return &GophorError{ SocketWriteErr, err }
    } else if count != len(b) {
        return &GophorError{ SocketWriteCountErr, nil }
    }
    return nil
}

func (worker *Worker) RespondGopher(data []byte) *GophorError {
    /* According to Gopher spec, only read up to first Tab or Crlf */
    dataStr := readUpToFirstTabOrCrlf(data)

    /* Handle URL request if presented */
    lenBefore := len(dataStr)
    dataStr = strings.TrimPrefix(dataStr, "URL:")
    switch len(dataStr) {
        case lenBefore-4:
            /* Send an HTML redirect to supplied URL */
            worker.LogError("Redirecting to %s\n", dataStr)
            return worker.Send(generateHtmlRedirect(dataStr))
        default:
            /* Do nothing */
    }

    /* Parse filesystem request and arg string slice */
    requestPath, args, gophorErr := parseFileSystemRequest(worker.Conn.Host.RootDir, dataStr)
    if gophorErr != nil {
        return gophorErr
    }

    /* Handle filesystem request */
    response, gophorErr := Config.FileSystem.HandleRequest(worker.Conn.Host, requestPath, args)
    if gophorErr != nil {
        /* Log to system and access logs, then return error */
        Config.SysLog.Error("", "Error serving %s: %s\n", dataStr, gophorErr.Error())
        worker.LogError("Failed to serve: %s\n", requestPath.AbsolutePath())
        return gophorErr
    }
    worker.Log("Served: %s\n", requestPath.AbsolutePath())

    /* Serve response */
    return worker.Send(response)
}

func readUpToFirstTabOrCrlf(data []byte) string {
    /* Only read up to first tab or cr-lf */
    dataStr := ""
    dataLen := len(data)
    for i := 0; i < dataLen; i += 1 {
        switch data[i] {
            case '\t':
                return dataStr
            case DOSLineEnd[0]:
                if i == dataLen-1 || data[i+1] == DOSLineEnd[1] {
                    return dataStr
                }
            default:
                dataStr += string(data[i])
        }
    }

    return dataStr
}
