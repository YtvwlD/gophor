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
    PathEnumerationErr  ErrorCode = iota
    IllegalPathErr      ErrorCode = iota
    FileStatErr         ErrorCode = iota
    FileOpenErr         ErrorCode = iota
    FileReadErr         ErrorCode = iota
    FileTypeErr         ErrorCode = iota
    DirListErr          ErrorCode = iota
    
    /* Sockets */
    SocketWriteErr      ErrorCode = iota
    SocketWriteCountErr ErrorCode = iota
    
    /* Parsing */
    InvalidRequestErr   ErrorCode = iota
    EmptyItemTypeErr    ErrorCode = iota
    EntityPortParseErr  ErrorCode = iota
    InvalidGophermapErr ErrorCode = iota
)

type GophorError struct {
    Code ErrorCode
    Err  error
}

func (e *GophorError) Error() string {
    var str string
    switch e.Code {
        case PathEnumerationErr:
            str = "path enumeration fail"
        case IllegalPathErr:
            str = "illegal path requested"
        case FileStatErr:
            str = "file stat fail"
        case FileOpenErr:
            str = "file open fail"
        case FileReadErr:
            str = "file read fail"
        case FileTypeErr:
            str = "invalid file type"
        case DirListErr:
            str = "directory read fail"

        case SocketWriteErr:
            str = "socket write fail"
        case SocketWriteCountErr:
            str = "socket write count mismatch"

        case InvalidRequestErr:
            str = "invalid request data"
        case EmptyItemTypeErr:
            str = "line string provides no dir entity type"
        case EntityPortParseErr:
            str = "parsing dir entity port"
        case InvalidGophermapErr:
            str = "invalid gophermap"

        default:
            str = "Unknown"
    }

    if e.Err != nil {
        return fmt.Sprintf("%s (%s)", str, e.Err.Error())
    } else {
        return fmt.Sprintf("%s", str)
    }
}
