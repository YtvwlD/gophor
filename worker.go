package main

import (
    "os"
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
    worker := new(Worker)
    worker.Conn = conn
    return worker
}

func (worker *Worker) Serve() {
    go func() {
        defer func() {
            /* Close-up shop */
            worker.Conn.Close()
        }()

        var count int
        var err error

        /* Read buffer + final result */
        buf := make([]byte, SocketReadBufSize)
        received := make([]byte, 0)

        /* Buffered read from listener */
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

            /* Try generate response bytes from error code */
            response := generateGopherErrorResponseFromCode(gophorErr.Code)

            /* If we got response bytes to send? SEND 'EM! */
            if response != nil {
                /* No gods. No masters. We don't care about error checking here */
                worker.SendRaw(response)
            }
        }
    }()
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

    /* Handle URL request if so */
    lenBefore := len(dataStr)
    dataStr = strings.TrimPrefix(dataStr, "URL:")
    switch len(dataStr) {
        case lenBefore-4:
            /* Handle URL prefix */
            worker.Log("Redirecting to URL: %s\n", data)
            return worker.SendRaw(generateHtmlRedirect(dataStr))
        default:
            /* Do nothing */
    }

    /* Sanitize supplied path */
    requestPath := sanitizePath(dataStr)

    /* Handle policy files */
    switch requestPath {
        case "/"+CapsTxtStr:
            return worker.SendRaw(generateCapsTxt())

        case "/"+RobotsTxtStr:
            return worker.SendRaw(generateRobotsTxt())
    }

    /* Open requestPath */
    file, err := os.Open(requestPath)
    if err != nil {
        return &GophorError{ FileOpenErr, err }
    }


    /* If not empty requestPath, check file type */
    fileType := FileTypeDir
    if requestPath != "." {
        stat, err := file.Stat()
        if err != nil {
            return &GophorError{ FileStatErr, err }
        }

        switch {
            case stat.Mode() & os.ModeDir != 0:
                // do nothing :)
            case stat.Mode() & os.ModeType == 0:
                fileType = FileTypeRegular
            default:
                fileType = FileTypeBad
        }
    }

    /* Don't need the file handle anymore */
    file.Close()

    /* TODO: work on efficiency */

    /* Handle file type */
    response := make([]byte, 0)
    switch fileType {
        /* Directory */
        case FileTypeDir:
            /* First try to serve gopher map */
            gophermapPath := path.Join(requestPath, "/"+GophermapFileStr)
            fileContents, gophorErr := Config.FileCache.FetchGophermap(&FileSystemRequest{ gophermapPath, worker.Conn.Host })
            if gophorErr != nil {
                /* Get directory listing instead */
                fileContents, gophorErr = listDir(&FileSystemRequest{ requestPath, worker.Conn.Host }, map[string]bool{})
                if gophorErr != nil {
                    return gophorErr
                }

                /* Add fileContents to response */
                response = append(response, fileContents...)
                worker.Log("serve dir: %s\n", requestPath)

                /* Finish directory listing with LastLine */
                response = append(response, []byte(LastLine)...)
            } else {
                /* Successfully loaded gophermap, add fileContents to response */
                response = append(response, fileContents...)
                worker.Log("serve gophermap: %s\n", gophermapPath)
            }

        /* Regular file */
        case FileTypeRegular:
            /* Read file contents */
            fileContents, gophorErr := Config.FileCache.FetchRegular(&FileSystemRequest{ requestPath, worker.Conn.Host })
            if gophorErr != nil {
                return gophorErr
            }

            /* Append fileContents to response */
            response = append(response, fileContents...)
            worker.Log("serve file: %s\n", requestPath)

        /* Unsupport file type */
        default:
            return &GophorError{ FileTypeErr, nil }
    }

    /* Append lastline */
    response = append(response, []byte(LastLine)...)

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
