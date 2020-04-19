package main

import (
    "fmt"
    "os"
    "net"
    "path"
    "strings"
)

const (
    SocketReadBufSize   = 256 /* Supplied selector shouldn't be longer than this anyways */
    MaxSocketReadChunks = 4
    FileReadBufSize     = 1024
)

type Worker struct {
    Socket  net.Conn
}

func NewWorker(socket *net.Conn) *Worker {
    worker := new(Worker)
    worker.Socket = *socket
    return worker
}

func (worker *Worker) Serve() {
    go func() {
        defer func() {
            /* Close-up shop */
            worker.Socket.Close()
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
            count, err = worker.Socket.Read(buf)
            if err != nil {
                logSystemError("Error reading from socket %s: %s\n", worker.Socket, err.Error())
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
                worker.SendErrorType("max socket read size reached\n")
                logSystemError("Reached max socket read size %d. Closing connection...\n", MaxSocketReadChunks*SocketReadBufSize)
                return
            }

            /* Keep count :) */
            iter += 1
        }

        /* Respond */
        gophorErr := worker.Respond(received)
        if gophorErr != nil {
            logSystemError("%s\n", gophorErr.Error())
        }
    }()
}

func (worker *Worker) SendErrorType(format string, args ...interface{}) {
    worker.SendRaw([]byte(fmt.Sprintf(string(TypeError)+"Error: "+format+LastLine, args...)))
}

func (worker *Worker) SendErrorText(format string, args ...interface{}) {
    worker.SendRaw([]byte(fmt.Sprintf("Error: "+format, args...)))
}

func (worker *Worker) SendRaw(b []byte) *GophorError {
    count, err := worker.Socket.Write(b)
    if err != nil {
        return &GophorError{ SocketWriteErr, err }
    } else if count != len(b) {
        return &GophorError{ SocketWriteCountErr, nil }
    }
    return nil
}

func (worker *Worker) Log(format string, args ...interface{}) {
    logAccess(worker.Socket.RemoteAddr().String()+" "+format, args...)
}

func (worker *Worker) LogError(format string, args ...interface{}) {
    logAccessError(worker.Socket.RemoteAddr().String()+" "+format, args...)
}

func (worker *Worker) SanitizePath(dataStr string) string {
    /* Clean path and trim '/' prefix if still exists */
    requestPath := strings.TrimPrefix(path.Clean(dataStr), "/")

    if !strings.HasPrefix(requestPath, "/") {
        requestPath = "/" + requestPath
    }

    return requestPath
}

func (worker *Worker) Respond(data []byte) *GophorError {
    /* Only read up to first tab or cr-lf */
    dataStr := ""
    dataLen := len(data)
    for i := 0; i < dataLen; i += 1 {
        if data[i] == '\t' {
            break
        } else if data[i] == DOSLineEnd[0] {
            if i == dataLen-1 {
                /* Chances are we'll NEVER reach here, still need to check */
                return &GophorError{ InvalidRequestErr, nil }
            } else if data[i+1] == DOSLineEnd[1] {
                break
            }
        }
        dataStr += string(data[i])
    }

    /* Sanitize supplied path */
    requestPath := worker.SanitizePath(dataStr)

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
        worker.SendErrorType("read fail\n") /* Purposely vague errors */
        return &GophorError{ FileOpenErr, err }
    }

    /* Leads to some more concise code below */
    type FileType int
    const(
        File FileType = iota
        Dir  FileType = iota
        Bad  FileType = iota
    )

    /* If not empty requestPath, check file type */
    fileType := Dir
    if requestPath != "." {
        stat, err := file.Stat()
        if err != nil {
            worker.SendErrorType("read fail\n") /* Purposely vague errors */
            return &GophorError{ FileStatErr, err }
        }

        switch {
            case stat.Mode() & os.ModeDir != 0:
                // do nothing :)
            case stat.Mode() & os.ModeType == 0:
                fileType = File
            default:
                fileType = Bad
        }
    }

    /* Don't need the file handle anymore */
    file.Close()

    /* TODO: work on efficiency */

    /* Handle file type */
    response := make([]byte, 0)
    switch fileType {
        /* Directory */
        case Dir:
            /* First try to serve gopher map */
            gophermapPath := path.Join(requestPath, "/"+GophermapFileStr)
            fileContents, gophorErr := GlobalFileCache.FetchGophermap(gophermapPath)
            if gophorErr != nil {
                /* Get directory listing instead */
                fileContents, gophorErr = listDir(requestPath, map[string]bool{})
                if gophorErr != nil {
                    worker.SendErrorType("dir list failed\n")
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
        case File:
            /* Read file contents */
            fileContents, gophorErr := GlobalFileCache.FetchRegular(requestPath)
            if gophorErr != nil {
                worker.SendErrorText("file read fail\n")
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
