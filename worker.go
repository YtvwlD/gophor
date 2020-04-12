package main

import (
    "fmt"
    "os"
    "io"
    "net"
    "bufio"
    "path"
    "strings"
    "bytes"
)

const (
    ShowHidden          = false
    SocketReadBufSize   = 512
    MaxSocketReadChunks = 4
    FileReadBufSize     = 512
    GopherMapFile       = "/gophermap"
    DefaultShell        = "/bin/sh"
)

type Worker struct {
    Socket  net.Conn
    Hidden  map[string]bool
}

func NewWorker(socket *net.Conn) *Worker {
    worker := new(Worker)
    worker.Socket = *socket
    worker.Hidden = map[string]bool{
        "gophermap": true,
    }
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

    /* Naughty directory traversal! Hackers get ROOT */
    if strings.HasPrefix(requestPath, "/") {
        worker.LogError("illegal path requested: %s\n", dataStr)
        requestPath = "."
    }

    return requestPath
}

func (worker *Worker) Respond(data []byte) *GophorError {
    /* Only read up to first tab or cr-lf */
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

    /* Sanitize supplied path */
    requestPath := worker.SanitizePath(dataStr)

    /* Open requestPath */
    file, err := os.Open(requestPath)
    if err != nil {
        worker.SendErrorType("read fail\n") /* Purposely vague errors */
        return &GophorError{ FileOpenErr, err }
    }
    defer file.Close()

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

    /* Handle file type */
    var response []byte
    var gophorErr *GophorError
    switch fileType {
        /* Directory */
        case Dir:
            /* First try to serve gopher map */
            requestPath = path.Join(requestPath, GopherMapFile)
            mapFile, err := os.Open(requestPath)
            defer mapFile.Close()

            if err == nil {
                /* Read GopherMapFile contents */
                worker.Log("serve gophermap: /%s\n", requestPath)

                response, gophorErr = worker.ReadGophermap(file, mapFile)
                if gophorErr != nil {
                    worker.SendErrorType("gophermap read fail\n")
                    return gophorErr
                }
            } else {
                /* Get directory listing */
                worker.Log("serve dir: /%s\n", requestPath)

                response, gophorErr = worker.ListDir(file)
                if gophorErr != nil {
                    worker.SendErrorType("dir list fail\n")
                    return gophorErr
                }

                /* Finish directory listing with LastLine */
                response = append(response, []byte(LastLine)...)
            }

        /* Regular file */
        case File:
            /* Read file contents */
            worker.Log("%s serve file: /%s\n", requestPath)

            response, gophorErr = worker.ReadFile(file)
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

func (worker *Worker) ReadGophermap(dir, mapFile *os.File) ([]byte, *GophorError) {
    fileContents := make([]byte, 0)

    /* Create reader and scanner from this */
    reader := bufio.NewReader(mapFile)
    scanner := bufio.NewScanner(reader)

    /* Setup scanner to split on CrLf */
    scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
        if atEOF && len(data) == 0  {
            /* At EOF, no more data */
            return 0, nil, nil
        }

        if i := bytes.Index(data, []byte{ '\r', '\n' }); i >= 0 {
            /* We have a full new-line terminate line */
            return i+2, data[0:i], nil
        }

        /* Request more data */
        return 0, nil, nil
    })

    /* Scan, format each token and add to fileContents */
    doEnd := false
    for scanner.Scan() {
        line := scanner.Text()

        /* Parse the line item type and handle */
        lineType := parseLineType(line)
        switch lineType {
            case TypeInfoNotStated:
                /* Append TypeInfo to the beginning of line */
                line = string(TypeInfo)+line+CrLf

            case TypeComment:
                /* We ignore this line */
                continue

            case TypeHiddenFile:
                /* Add to hidden files map */
                worker.Hidden[line[1:]] = true

            case TypeSubGophermap:
                /* Try to read subgophermap of file name */
                line = string(TypeInfo)+"Error: subgophermaps not supported"+CrLf

/*
                subMapFile, err := os.Open(line[1:])
                if err != nil {
                    worker.LogError("error opening subgophermap: /%s --> %s\n", mapFile.Name(), line[1:])
                    line = fmt.Sprintf(string(TypeInfo)+"Error reading subgophermap: %s"+CrLf, line[1:])
                } else {
                    subMapContent, gophorError := worker.ReadFile(subMapFile)
                    if gophorError != nil {
                        worker.LogError("error reading subgophermap: /%s --> %s\n", mapFile.Name(), line[1:])
                        line = fmt.Sprintf(string(TypeInfo)+"Error reading subgophermap: %s"+CrLf, line[1:])
                    } else {
                        line = strings.Replace(string(subMapContent), "\n", CrLf, -1)
                        if !strings.HasSuffix(line, CrLf) {
                            line += CrLf
                        }
                    }
                }
*/

            case TypeExec:
                /* Try executing supplied line */
                line = string(TypeInfo)+"Error: inline shell commands not support"+CrLf

/*
                err := exec.Command(line[1:]).Run()
                if err != nil {
                    line = fmt.Sprintf(string(TypeInfo)+"Error executing command: %s"+CrLf, line[1:])
                } else {
                    line = strings.Replace(string(""), "\n", CrLf, -1)
                    if !strings.HasSuffix(line, CrLf) {
                        line += CrLf
                    }
                }
*/

            case TypeEnd:
                /* Lastline, break out at end of loop */
                doEnd = true
                line = LastLine

            case TypeEndBeginList:
                /* Read current directory listing then break out at end of loop */
                doEnd = true
                dirListing, gophorErr := worker.ListDir(dir)
                if gophorErr != nil {
                    return nil, gophorErr
                }
                line = string(dirListing) + LastLine

            default:
                line += CrLf
        }

        /* Append generated line to total fileContents */
        fileContents = append(fileContents, []byte(line)...)

        /* Break out of read loop if requested */
        if doEnd {
            break
        }
    }

    /* If scanner didn't finish cleanly, return nil and error */
    if scanner.Err() != nil {
        return nil, &GophorError{ FileReadErr, scanner.Err() }
    }

    /* If we never hit doEnd, append a LastLine ourselves */
    if !doEnd {
        fileContents = append(fileContents, []byte(LastLine)...)
    }
    
    return fileContents, nil
}

func (worker *Worker) ReadFile(file *os.File) ([]byte, *GophorError) {
    var count int
    fileContents := make([]byte, 0)
    buf := make([]byte, FileReadBufSize)

    var err error
    reader := bufio.NewReader(file)

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

func (worker *Worker) ListDir(dir *os.File) ([]byte, *GophorError) {
    files, err := dir.Readdir(-1)
    if err != nil {
        return nil, &GophorError{ DirListErr, err }
    }

    var entity *DirEntity
    dirContents := make([]byte, 0)

    for _, file := range files {
        /* Skip dotfiles + gophermap file + requested hidden */
        if file.Name()[0] == '.' || file.Name() == "gophermap" {
            continue
        } else if _, ok := worker.Hidden[file.Name()]; ok {
            continue
        }

        /* Handle file, directory or ignore others */
        switch {
            case file.Mode() & os.ModeDir != 0:
                /* Directory -- create directory listing */
                itemPath := path.Join(dir.Name(), file.Name())
                entity = newDirEntity(TypeDirectory, file.Name(), "/"+itemPath, *ServerHostname, *ServerPort)
                dirContents = append(dirContents, entity.Bytes()...)

            case file.Mode() & os.ModeType == 0:
                /* Regular file -- find item type and creating listing */
                itemPath := path.Join(dir.Name(), file.Name())
                itemType := getItemType(itemPath)
                entity = newDirEntity(itemType, file.Name(), "/"+itemPath, *ServerHostname, *ServerPort)
                dirContents = append(dirContents, entity.Bytes()...)

            default:
                /* Ignore */
        }
    }

    return dirContents, nil
}
