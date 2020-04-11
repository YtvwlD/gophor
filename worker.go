package main

import (
    "fmt"
    "os"
    "io"
    "net"
    "bufio"
    "path"
    "strings"
)

const (
    ShowHidden          = false
    SocketReadBufSize   = 512
    MaxSocketReadChunks = 4
    FileReadBufSize     = 512
    GopherMapFile       = "/gophermap"
)

type Worker struct {
    Socket net.Conn
}

func (worker *Worker) Init(socket *net.Conn) {
    worker.Socket = *socket
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
        gophorErr := serverRespond(worker, received)
        if gophorErr != nil {
            logSystemError("%s\n", gophorErr.Error())
        }
    }()
}

func (worker *Worker) SendErrorType(format string, args ...interface{}) {
    worker.SendRaw(string(TypeError)+"Error: "+format+LastLine, args...)
}

func (worker *Worker) SendErrorRaw(format string, args ...interface{}) {
    worker.SendRaw("Error: "+format, args...)
}

func (worker *Worker) SendRaw(format string, args ...interface{}) {
    worker.Socket.Write([]byte(fmt.Sprintf(format, args...)))
}

func serverRespond(worker *Worker, data []byte) *GophorError {
    /* Clean initial data:
     * Should usually start with a '/' since the selector response we send
     * starts with a '/' for worker formatting reasons.
     */
    dataStr := strings.TrimPrefix(strings.TrimSuffix(string(data), CrLf), "/")

    /* Clean path and get shortest possible from current directory */
    requestPath := path.Clean(dataStr)

    /* Ensure alway a relative paths + WITHIN ServerDir, serve them root otherwise */
    if strings.HasPrefix(requestPath, "/") || strings.HasPrefix(requestPath, "..") {
        logAccessError("%s illegal path requested: %s\n", worker.Socket.RemoteAddr(), dataStr)
        requestPath = "."
    }

    var response []byte
    var gophorErr *GophorError
    var err error

    /* Open requestPath */
    fd, err := os.Open(requestPath)
    if err != nil {
        worker.SendErrorType("read fail\n") /* Purposely vague errors */ 
        return &GophorError{ FileOpenErr, err }
    }
    defer fd.Close()

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
        stat, err := fd.Stat()
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

    /* Handle Dir / File / error otherwise */
    switch fileType {
        /* Directory */
        case Dir:
            /* First try to serve gopher map */
            requestPath = path.Join(requestPath, GopherMapFile)
            fd2, err := os.Open(requestPath)
            defer fd2.Close()

            if err == nil {
                /* Read GopherMapFile contents */
                logAccess("%s serve gophermap: %s\n", worker.Socket.RemoteAddr(), fd2.Name())

                response, gophorErr = readFile(fd2)
                if gophorErr != nil {
                    worker.SendErrorType("gophermap read fail\n")
                    return gophorErr
                }
            } else {
                /* Get directory listing */
                logAccess("%s serve dir: %s\n", worker.Socket.RemoteAddr(), fd.Name())

                response, gophorErr = listDir(fd)
                if gophorErr != nil {
                    worker.SendErrorType("dir list fail\n")
                    return gophorErr
                }
            }

        /* Regular file */
        case File:
            /* Read file contents */
            logAccess("%s serve file: %s\n", worker.Socket.RemoteAddr(), fd.Name())

            response, gophorErr = readFile(fd)
            if gophorErr != nil {
                worker.SendErrorRaw("file read fail\n")
                return gophorErr
            }

        /* Unsupport file type */
        default:
            return &GophorError{ FileTypeErr, nil }
    }

    /* Always finish response with LastLine bytes */
    response = append(response, []byte(LastLine)...)

    /* Serve response */
    count, err := worker.Socket.Write(response)
    if err != nil {
        return &GophorError{ SocketWriteErr, err }
    } else if count != len(response) {
        return &GophorError{ SocketWriteCountErr, nil }
    }

    return nil
}

func readFile(fd *os.File) ([]byte, *GophorError) {
    var count int
    fileContents := make([]byte, 0)
    buf := make([]byte, FileReadBufSize)

    var err error
    reader := bufio.NewReader(fd)

    for {
        count, err = reader.Read(buf)
        if err != nil && err != io.EOF {
            return nil, &GophorError{ FileReadErr, err }
        }        

        for i := 0; i < count; i += 1 {
            if buf[i] == 0 {
                break
            }

            fileContents = append(fileContents, buf[i])
        }

        if count < FileReadBufSize {
            break
        }
    }

    return fileContents, nil
}

func listDir(fd *os.File) ([]byte, *GophorError) {
    files, err := fd.Readdir(-1)
    if err != nil {
        return nil, &GophorError{ DirListErr, err }
    }

    var entity *DirEntity
    dirContents := make([]byte, 0)

    for _, file := range files {
        if !ShowHidden && file.Name()[0] == '.' {
            continue
        }

        switch {
            case file.Mode() & os.ModeDir != 0:
                /* Directory! */
                itemPath := path.Join(fd.Name(), file.Name())
                entity = newDirEntity(TypeDirectory, file.Name(), "/"+itemPath, *ServerHostname, *ServerPort)
                dirContents = append(dirContents, entity.Bytes()...)

            case file.Mode() & os.ModeType == 0:
                /* Regular file */
                itemPath := path.Join(fd.Name(), file.Name())
                itemType := getItemType(itemPath)
                entity = newDirEntity(itemType, file.Name(), "/"+itemPath, *ServerHostname, *ServerPort)
                dirContents = append(dirContents, entity.Bytes()...)

            default:
                /* Ignore */
        }
    }

    return dirContents, nil
}
