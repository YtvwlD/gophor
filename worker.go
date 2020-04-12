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

func serverRespond(worker *Worker, data []byte) *GophorError {
    /* Only read up to first tab / cr-lf */
    dataStr := ""
    dataLen := len(data)
    for i := 0; i < dataLen; i += 1 {
        if data[i] == Tab {
            break
        } else if data[i] == CrLf[0] {
            if i == dataLen-1 {
                /* Chances are we'll NEVER reach here, still need to check */
                return &GophorError{ InvalidRequestErr, nil }
            } else if data[i+1] == CrLf[1] {
                break
            }
        }
        dataStr += string(data[i])
    }

    /* Clean path and get shortest possible from current directory */
    requestPath := path.Clean(dataStr)

    /* Even if asking for root, we trim the initial '/' now its been cleaned up */
    requestPath = strings.TrimPrefix(requestPath, "/")

    /* Ensure alway a relative paths + WITHIN ServerDir, serve them root otherwise */
    if strings.HasPrefix(requestPath, "..") {
        logAccessError("%s illegal path requested: %s\n", worker.Socket.RemoteAddr(), dataStr)
        requestPath = "."
    }

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

    /* Handle file type */
    var response []byte
    var gophorErr *GophorError
    switch fileType {
        /* Directory */
        case Dir:
            /* First try to serve gopher map */
            requestPath = path.Join(requestPath, GopherMapFile)
            fd2, err := os.Open(requestPath)
            defer fd2.Close()

            if err == nil {
                /* Read GopherMapFile contents */
                logAccess("%s serve gophermap: /%s\n", worker.Socket.RemoteAddr(), requestPath)

                response, gophorErr = readFile(fd2)
                if gophorErr != nil {
                    worker.SendErrorType("gophermap read fail\n")
                    return gophorErr
                }
            } else {
                /* Get directory listing */
                logAccess("%s serve dir: /%s\n", worker.Socket.RemoteAddr(), requestPath)

                response, gophorErr = listDir(fd)
                if gophorErr != nil {
                    worker.SendErrorType("dir list fail\n")
                    return gophorErr
                }
            }

            /* Have to finish directory listings with LastLine */
            response = append(response, []byte(LastLine)...)

        /* Regular file */
        case File:
            /* Read file contents */
            logAccess("%s serve file: /%s\n", worker.Socket.RemoteAddr(), requestPath)

            response, gophorErr = readFile(fd)
            if gophorErr != nil {
                worker.SendErrorText("file read fail\n")
                return gophorErr
            }

        /* Unsupport file type */
        default:
            return &GophorError{ FileTypeErr, nil }
    }

    /* Serve response */
    return worker.SendRaw(response)
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
        /* Unless specificially compiled not to, we skip hidden files */
        if !ShowHidden && file.Name()[0] == '.' {
            continue
        }

        /* Handle file, directory or ignore others */
        switch {
            case file.Mode() & os.ModeDir != 0:
                /* Directory -- create directory listing */
                itemPath := path.Join(fd.Name(), file.Name())
                entity = newDirEntity(TypeDirectory, file.Name(), "/"+itemPath, *ServerHostname, *ServerPort)
                dirContents = append(dirContents, entity.Bytes()...)

            case file.Mode() & os.ModeType == 0:
                /* Regular file -- find item type and creating listing */
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
