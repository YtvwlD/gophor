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
            count, err = client.Socket.Read(buf)
            if err != nil {
                fmt.Fprintf(os.Stderr, "Error reading from socket %s: %v\n", client.Socket, err)
                return
            }

            for i := 0; i < count; i += 1 {
                if buf[i] == 0 {
                    break
                }

                received = append(received, buf[i])
            }

            if count < SocketReadBufSize {
                /* EOF */
                break
            }

            if iter == MaxSocketReadChunks {
                fmt.Fprintf(os.Stderr, "Reached max socket read size: %d. Closing connection...\n", MaxSocketReadChunks*SocketReadBufSize)
                return
            }

            iter += 1
        }

        /* Respond */
        gophorErr := serverRespond(client, received)
        if gophorErr != nil {
            fmt.Fprintf(os.Stderr, gophorErr.Error() + "\n")
        }
    }()
}

func serverRespond(client *Client, data []byte) *GophorError {
    path := string(data)
    pathLen := len(path)
    
    var response []byte
    var gophorErr *GophorError
    var err error
    if pathLen == 2 && path == "\r\n" {
        /* Empty line received, treat as dir listing for root */
        fd, err := os.Open(GopherMapFile)
        defer fd.Close()

        if err == nil {
            /* Read GopherMapFile contents */
            fileContents, gophorErr := readFile(fd)
            if gophorErr != nil {
                return gophorErr
            }

            /* Serve GopherMapFile */
            count, err := client.Socket.Write(fileContents)
            if err != nil {
                return &GophorError{ SocketWrite, err }
            } else if count != len(fileContents) {
                return &GophorError{ SocketWriteCount, nil }
            }
        } else {
            /* Close fd, re-open directory instead */
            fd.Close()
            fd, err = os.Open("/")

            /* Get directory listing */
            response, gophorErr = listDir(fd)
            if gophorErr != nil {
                return gophorErr
            }
        }
    } else {
        path = strings.Trim(path, "\r\n")

        fd, err := os.Open(path)
        if err != nil {
            return &GophorError{ FileOpen, err }
        }
        defer fd.Close()

        stat, err := fd.Stat()
        if err != nil {
            return &GophorError{ FileStat, err }
        }

        /* Determine if path or directory */
        switch {
            /* Directory */
            case stat.Mode() & os.ModeDir != 0:
                /* First try to serve gopher map */
                fd2, err := os.Open(path + GopherMapFile)
                defer fd2.Close()

                if err == nil {
                    /* Read GopherMapFile contents */
                    response, gophorErr = readFile(fd2)
                    if gophorErr != nil {
                        return gophorErr
                    }
                } else {
                    /* Get directory listing */
                    response, gophorErr = listDir(fd)
                    if gophorErr != nil {
                        return gophorErr
                    }
                }

            /* Regular file */
            case stat.Mode() & os.ModeType == 0:
                /* Read file contents */
                response, gophorErr = readFile(fd)
                if gophorErr != nil {
                    return gophorErr
                }

            /* Unsupport file type */
            default:
                return &GophorError{ FileType, nil }
        }
    }
    
    /* Always finish response with LastLine bytes */
    response = append(response, []byte(LastLine)...)

    /* Serve response */
    count, err := client.Socket.Write(response)
    if err != nil {
        return &GophorError{ SocketWrite, err }
    } else if count != len(response) {
        return &GophorError{ SocketWriteCount, nil }
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
            return nil, &GophorError{ FileRead, err }
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
        return nil, &GophorError{ DirList, err }
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
                fullPath := path.Join(fd.Name(), file.Name())
                entity = newDirEntity(TypeDirectory, file.Name(), fullPath, *ServerHostname, *ServerPort)
                dirContents = append(dirContents, entity.Bytes()...)

            case file.Mode() & os.ModeType == 0:
                /* Regular file */
                fullPath := path.Join(fd.Name(), file.Name())
                itemType := getItemType(fullPath)
                entity = newDirEntity(itemType, file.Name(), fullPath, *ServerHostname, *ServerPort)
                dirContents = append(dirContents, entity.Bytes()...)

            default:
                /* Ignore */
                fmt.Fprintf(os.Stderr, "List dir: skipping file %s of invalid type\n", file.Name())
        }
    }

    return dirContents, nil
}
