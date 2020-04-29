package main

import (
    "path"
    "strings"
)

type FileType int
const (
    SocketReadBufSize   = 256 /* Supplied selector shouldn't be longer than this anyways */
    MaxSocketReadChunks = 4
    FileReadBufSize     = 1024

    /* Leads to some more concise code below */
    FileTypeRegular FileType = iota
    FileTypeDir     FileType = iota
    FileTypeBad     FileType = iota
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
            Config.LogSystemError("Error reading from socket on port %s: %s\n", worker.Conn.Host.Port, err.Error())
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
            Config.LogSystemError("Reached max socket read size %d. Closing connection...\n", MaxSocketReadChunks*SocketReadBufSize)
            return
        }

        /* Keep count :) */
        iter += 1
    }

    /* Handle request */
    gophorErr := worker.RespondGopher(received)

    /* Handle any error */
    if gophorErr != nil {
        Config.LogSystemError("%s\n", gophorErr.Error())

        /* Generate response bytes from error code */
        response := generateGopherErrorResponseFromCode(gophorErr.Code)

        /* If we got response bytes to send? SEND 'EM! */
        if response != nil {
            /* No gods. No masters. We don't care about error checking here */
            worker.SendRaw(response)
        }
    }
}

func (worker *Worker) SendRaw(b []byte) *GophorError {
    count, err := worker.Conn.Write(b)
    if err != nil {
        return &GophorError{ SocketWriteErr, err }
    } else if count != len(b) {
        return &GophorError{ SocketWriteCountErr, nil }
    }
    return nil
}

func (worker *Worker) Log(format string, args ...interface{}) {
    Config.LogAccess(worker.Conn.RemoteAddr().String(), format, args...)
}

func (worker *Worker) LogError(format string, args ...interface{}) {
    Config.LogAccessError(worker.Conn.RemoteAddr().String(), format, args...)
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
            worker.Log("Redirecting to URL: %s\n", data)
            return worker.SendRaw(generateHtmlRedirect(dataStr))
        default:
            /* Do nothing */
    }

    /* Sanitize supplied path */
    requestPath := sanitizePath(dataStr)

    /* Append lastline */
    response, gophorErr := Config.FileSystem.HandleRequest(requestPath, worker.Conn.Host)
    if gophorErr != nil {
        return gophorErr
    }
    Config.LogSystem("Served: %s\n", requestPath)

    /* Serve response */
    return worker.SendRaw(response)
}

func readUpToFirstTabOrCrlf(data []byte) string {
    /* Only read up to first tab or cr-lf */
    dataStr := ""
    dataLen := len(data)
    for i := 0; i < dataLen; i += 1 {
        if data[i] == '\t' {
            break
        } else if data[i] == DOSLineEnd[0] {
            if i == dataLen-1 || data[i+1] == DOSLineEnd[1] {
                /* Finished on Unix line end, NOT DOS */
                break
            }
        }

        dataStr += string(data[i])
    }

    return dataStr
}

func sanitizePath(dataStr string) string {
    /* Clean path and trim '/' prefix if still exists */
    requestPath := strings.TrimPrefix(path.Clean(dataStr), "/")

    if requestPath == "." {
        requestPath = "/"
    } else if !strings.HasPrefix(requestPath, "/") {
        requestPath = "/" + requestPath
    }

    return requestPath
}
