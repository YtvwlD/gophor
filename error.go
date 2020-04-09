package main

import (
    "fmt"
)

/*
 * Client error data structure
 */
type ErrorCode int
const (
    /* Filesystem */
    FileStat         ErrorCode = iota
    FileOpen         ErrorCode = iota
    FileRead         ErrorCode = iota
    FileType         ErrorCode = iota
    DirList          ErrorCode = iota
    
    /* Sockets */
    SocketWrite      ErrorCode = iota
    SocketWriteCount ErrorCode = iota
    
    /* Parsing */
    EmptyItemType    ErrorCode = iota
    EntityPortParse  ErrorCode = iota
)

type GophorError struct {
    Code ErrorCode
    Err  error
}

func (e *GophorError) Error() string {
    var str string
    switch e.Code {
        case FileStat:
            str = "file stat fail"
        case FileOpen:
            str = "file open fail"
        case FileRead:
            str = "file read fail"
        case FileType:
            str = "invalid file type"
        case DirList:
            str = "directory read fail"

        case SocketWrite:
            str = "socket write fail"
        case SocketWriteCount:
            str = "socket write count mismatch"

        case EmptyItemType:
            str = "line string provides no dir entity type"
        case EntityPortParse:
            str = "parsing dir entity port"

        default:
            str = "Unknown"
    }

    if e.Err != nil {
        return fmt.Sprintf("GophorError: %s (%s)", str, e.Err.Error())
    } else {
        return fmt.Sprintf("GophorError: %s", str)
    }
}
