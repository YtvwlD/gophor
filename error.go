package main

import (
    "fmt"
)

/* Simple error code type defs */
type ErrorCode int
type ErrorResponseCode int
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
    InvalidGophermapErr ErrorCode = iota

    /* Executing */
    BufferReadErr       ErrorCode = iota
    CommandStartErr     ErrorCode = iota
    CommandExitCodeErr  ErrorCode = iota

    /* Error Response Codes */
    ErrorResponse200 ErrorResponseCode = iota
    ErrorResponse400 ErrorResponseCode = iota
    ErrorResponse401 ErrorResponseCode = iota
    ErrorResponse403 ErrorResponseCode = iota
    ErrorResponse404 ErrorResponseCode = iota
    ErrorResponse408 ErrorResponseCode = iota
    ErrorResponse410 ErrorResponseCode = iota
    ErrorResponse500 ErrorResponseCode = iota
    ErrorResponse501 ErrorResponseCode = iota
    ErrorResponse503 ErrorResponseCode = iota
    NoResponse       ErrorResponseCode = iota
)

/* Simple GophorError data structure to wrap another error */
type GophorError struct {
    Code ErrorCode
    Err  error
}

/* Convert error code to string */
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
        case InvalidGophermapErr:
            str = "invalid gophermap"

        case BufferReadErr:
            str = "buffer read fail"
        case CommandStartErr:
            str = "command start fail"
        case CommandExitCodeErr:
            str = "command exit code non-zero"

        default:
            str = "Unknown"
    }

    if e.Err != nil {
        return fmt.Sprintf("%s (%s)", str, e.Err.Error())
    } else {
        return fmt.Sprintf("%s", str)
    }
}

/* Convert a gophor error code to appropriate error response code */
func gophorErrorToResponseCode(code ErrorCode) ErrorResponseCode {
    switch code {
        case PathEnumerationErr:
            return ErrorResponse400
        case IllegalPathErr:
            return ErrorResponse403
        case FileStatErr:
            return ErrorResponse404
        case FileOpenErr:
            return ErrorResponse404
        case FileReadErr:
            return ErrorResponse404
        case FileTypeErr:
            /* If wrong file type, just assume file not there */
            return ErrorResponse404
        case DirListErr:
            return ErrorResponse404

        /* These are errors _while_ sending, no point trying to send error  */
        case SocketWriteErr:
            return NoResponse
        case SocketWriteCountErr:
            return NoResponse

        case InvalidRequestErr:
            return ErrorResponse400
        case EmptyItemTypeErr:
            return ErrorResponse500
        case InvalidGophermapErr:
            return ErrorResponse500

        case BufferReadErr:
            return ErrorResponse500
        case CommandStartErr:
            return ErrorResponse500
        case CommandExitCodeErr:
            return ErrorResponse500

        default:
            return ErrorResponse503
    }
}

/* Generates gopher protocol compatible error response from our code */
func generateGopherErrorResponseFromCode(code ErrorCode) []byte {
    responseCode := gophorErrorToResponseCode(code)
    if responseCode == NoResponse {
        return nil
    }
    return generateGopherErrorResponse(responseCode)
}

/* Generates gopher protocol compatible error response for response code */
func generateGopherErrorResponse(code ErrorResponseCode) []byte {
    return buildError(code.String())
}

/* Error response code to string */
func (e ErrorResponseCode) String() string {
    switch e {
        case ErrorResponse200:
            return "200 OK"
        case ErrorResponse400:
            return "400 Bad Request"
        case ErrorResponse401:
            return "401 Unauthorised"
        case ErrorResponse403:
            return "403 Forbidden"
        case ErrorResponse404:
            return "404 Not Found"
        case ErrorResponse408:
            return "408 Request Time-out"
        case ErrorResponse410:
            return "410 Gone"
        case ErrorResponse500:
            return "500 Internal Server Error"
        case ErrorResponse501:
            return "501 Not Implemented"
        case ErrorResponse503:
            return "503 Service Unavailable"
        default:
            /* Should not have reached here */
            Config.SysLog.Fatal("", "Unhandled ErrorResponseCode type\n")
            return ""
    }
}
