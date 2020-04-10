package main

import (
    "log"
    "fmt"
    "os"
    "io"
    "net"
    "bufio"
    "path"
    "strings"
)

const (
    GopherMapFile = "/gophermap"
)

type Client struct {
    Cmd    chan Command
    Socket net.Conn
}

func (client *Client) Init(conn *net.Conn) {
    client.Cmd    = make(chan Command)
    client.Socket = *conn
}

func (client *Client) Start() {
    go func() {
        defer func() {
            /* Close-up shop */
            client.Socket.Close()
            close(client.Cmd)
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
            count, err = client.Socket.Read(buf)
            if err != nil {
                client.Log("Error reading from socket %s: %v\n", client.Socket, err.Error())
                return
            }

            /* Only copy non-null bytes */
            received = append(received, buf[:count])

            /* If count is less than expected read size, we've hit EOF */
            if count < SocketReadBufSize {
                /* EOF */
                break
            }

            /* Hit max read chunk size, send error + close connection */
            if iter == MaxSocketReadChunks {
                client.Error("max socket read size reached\n")
                client.Log("Reached max socket read size: %d. Closing connection...\n", MaxSocketReadChunks*SocketReadBufSize)
                return
            }

            /* Keep count :) */
            iter += 1
        }

        /* Respond */
        gophorErr := serverRespond(client, received)
        if gophorErr != nil {
            log.Printf(gophorErr.Error() + "\n")
        }
    }()
}

func (client *Client) Log(format string, args ...interface{}) {
    log.Printf(client.Socket.RemoteAddr().String()+" "+format, args...)
}

func (client *Client) SendError(format string, args ...interface{}) {
    response := make([]byte, 0)
    response = append(response, byte(TypeError))

    /* Format error message and append to response */
    message := fmt.Sprintf(format, args...)
    response = append(response, []byte(message)...)
    response = append(response, []byte(LastLine)...)

    /* We're sending an error, if this fails then fuck it lol */
    client.Socket.Write(response)
}

func serverRespond(client *Client, data []byte) *GophorError {
    /* Clean initial data:
     * Should usually start with a '/' since the selector response we send
     * starts with a '/' client formatting reasons.
     */
    dataStr := strings.TrimPrefix(strings.TrimSuffix(string(data), CrLf), "/")

    /* Clean path and get shortest possible from current directory */
    requestPath := path.Clean(dataStr)

    /* Ensure alway a relative paths + WITHIN ServerDir, serve them root otherwise */
    if strings.HasPrefix(requestPath, "/") || strings.HasPrefix(requestPath, "..") {
        client.Log("Illegal path requested: %s\n", dataStr)
        requestPath = "."
    }

    var response []byte
    var gophorErr *GophorError
    var err error

    /* Open requestPath */
    fd, err := os.Open(requestPath)
    if err != nil {
        client.SendError("%s read fail\n", requestPath) /* Purposely vague errors */ 
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
            client.SendError("%s read fail\n", requestPath) /* Purposely vague errors */
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
                client.Log("%s SERVER GOPHERMAP: %s\n", fd2.Name())

                response, gophorErr = readFile(fd2)
                if gophorErr != nil {
                    client.SendError("%s read fail\n", fd2.Name())
                    return gophorErr
                }
            } else {
                /* Get directory listing */
                client.Log("SERVE DIR: %s\n", fd.Name())

                response, gophorErr = listDir(fd)
                if gophorErr != nil {
                    client.SendError("%s dir list fail\n", fd.Name())
                    return gophorErr
                }
            }

        /* Regular file */
        case File:
            /* Read file contents */
            client.Log("SERVE FILE: %s\n", fd.Name())

            response, gophorErr = readFile(fd)
            if gophorErr != nil {
                client.SendError("%s read fail\n", fd.Name())
                return gophorErr
            }

        /* Unsupport file type */
        default:
            return &GophorError{ FileTypeErr, nil }
    }

    /* Always finish response with LastLine bytes */
    response = append(response, []byte(LastLine)...)

    /* Serve response */
    count, err := client.Socket.Write(response)
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
