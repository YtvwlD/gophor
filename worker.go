package main

import (
    "strings"
    "io"
)

const (
    /* Socket settings */
    SocketReadBufSize = 1024
    MaxSocketReadChunks = 4
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
            if err == io.EOF {
                break
            }

            Config.SysLog.Error("", "Error reading from socket on port %s: %s\n", worker.Conn.Host.Port, err.Error())
            return
        }

        /* Only copy non-null bytes */
        received = append(received, buf[:count]...)
        if count < SocketReadBufSize {
            /* Reached EOF */
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
    Config.AccLog.Info("("+worker.Conn.Client.Ip+") ", format, args...)
}

func (worker *Worker) LogError(format string, args ...interface{}) {
    Config.AccLog.Error("("+worker.Conn.Client.Ip+") ", format, args...)
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
    /* Only read up to first tab or cr-lf */
    dataStr := ""
    dataLen := len(data)
    for i := 0; i < dataLen; i += 1 {
        if data[i] == Tab[0] {
            break
        } else if data[i] == DOSLineEnd[0] {
            if i == dataLen-1 || data[i+1] == DOSLineEnd[1] {
                break
            } else {
                dataStr += string(data[i])
            }
        } else {
            dataStr += string(data[i])
        }
    }

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

    /* Create new filesystem request */
    request := NewSanitizedFileSystemRequest(worker.Conn.Host, worker.Conn.Client, dataStr)

    /* Handle filesystem request */
    response, gophorErr := Config.FileSystem.HandleRequest(request)
    if gophorErr != nil {
        /* Log to system and access logs, then return error */
        return gophorErr
    }
    worker.Log("Served: %s\n", request.AbsPath())

    /* Serve response */
    return worker.Send(response)
}
