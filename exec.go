package main

import (
    "os/exec"
    "syscall"
    "bytes"
    "strconv"
    "io"
)

type ExecType int
const (
    /* Executable type */
    ExecTypeCgi     ExecType = iota
    ExecTypeRegular ExecType = iota

    SafeExecPath = "/usr/bin:/bin"
    ReaderBufSize = 1024
)

func setupExecEnviron() []string {
    return []string {
        envKeyValue("PATH", SafeExecPath),
    }
}

func setupInitialCgiEnviron() []string {
    return []string{
        /* RFC 3875 standard */
        envKeyValue("GATEWAY_INTERFACE",  "CGI/1.1"), /* MUST be set to the dialect of CGI being used by the server */
        envKeyValue("SERVER_SOFTWARE",    "gophor/"+GophorVersion), /* MUST be set to name and version of server software serving this request */
        envKeyValue("SERVER_PROTOCOL",    "RFC1436"), /* MUST be set to name and version of application protocol used for this request */
        envKeyValue("CONTENT_LENGTH",     "0"), /* Contains size of message-body attached (always 0 so we set here) */
        envKeyValue("REQUEST_METHOD",     "GET"), /* MUST be set to method by which script should process request. Always GET */

        /* Non-standard */
        envKeyValue("PATH",               SafeExecPath),
        envKeyValue("COLUMNS",            strconv.Itoa(Config.PageWidth)),
        envKeyValue("GOPHER_CHARSET",     Config.CharSet),
    }
}

func executeCgi(request *FileSystemRequest) ([]byte, *GophorError) {
    /* Get initial CgiEnv variables */
    cgiEnv := Config.CgiEnv

    /* RFC 3875 standard */
    cgiEnv = append(cgiEnv, envKeyValue("SERVER_NAME",     request.Host.Name)) /* MUST be set to name of server host client is connecting to */
    cgiEnv = append(cgiEnv, envKeyValue("SERVER_PORT",     request.Host.Port)) /* MUST be set to the server port that client is connecting to */
    cgiEnv = append(cgiEnv, envKeyValue("REMOTE_ADDR",     request.Client.Ip)) /* Remote client addr, MUST be set */

    /* This way we get query string without initial delimiter */
    var queryString string
    if len(request.Parameters[0]) > 0 {
        queryString = request.Parameters[0][1:]
    } else {
        queryString = request.Parameters[0]
    }
    cgiEnv = append(cgiEnv, envKeyValue("QUERY_STRING",    queryString)) /* URL encoded search or parameter string, MUST be set even if empty */

    cgiEnv = append(cgiEnv, envKeyValue("PATH_INFO",       "")) /* Sub-resource to be fetched by script, derived from path hierarch portion of URI. NOT URL encoded */
    cgiEnv = append(cgiEnv, envKeyValue("PATH_TRANSLATED", request.AbsPath())) /* Take PATH_INFO, parse as local URI and append root dir */
    cgiEnv = append(cgiEnv, envKeyValue("SCRIPT_NAME",     "/"+request.RelPath())) /* URI path (not URL encoded) which could identify the CGI script (rather than script's output) */

/* We ignore these due to just CBA and we're not implementing authorization yet */
//    cgiEnv = append(cgiEnv, envKeyValue("AUTH_TYPE",       "")) /* Any method used my server to authenticate user, MUST be set if auth'd */
//    cgiEnv = append(cgiEnv, envKeyValue("CONTENT_TYPE",    "")) /* Only a MUST if HTTP content-type set (so never for gopher) */
//    cgiEnv = append(cgiEnv, envKeyValue("REMOTE_IDENT",    "")) /* Remote client identity information */
//    cgiEnv = append(cgiEnv, envKeyValue("REMOTE_HOST",     "")) /* Remote client domain name */
//    cgiEnv = append(cgiEnv, envKeyValue("REMOTE_USER",     "")) /* Remote user ID, if AUTH_TYPE, MUST be set */

    /* Non-standard */
    cgiEnv = append(cgiEnv, envKeyValue("SELECTOR",        request.SelectorPath()))
    cgiEnv = append(cgiEnv, envKeyValue("SCRIPT_FILENAME", request.AbsPath()))
    cgiEnv = append(cgiEnv, envKeyValue("DOCUMENT_ROOT",   request.RootDir))
    cgiEnv = append(cgiEnv, envKeyValue("REQUEST_URI",     "/"+request.RelPath()+request.Parameters[0]))

    return execute(cgiEnv, request.AbsPath(), nil)
}

func executeFile(request *FileSystemRequest) ([]byte, *GophorError) {
    return execute(Config.Env, request.AbsPath(), request.Parameters)
}

func executeCommand(request *FileSystemRequest) ([]byte, *GophorError) {
    return execute(Config.Env, request.AbsPath(), request.Parameters)
}

func execute(env []string, path string, args []string) ([]byte, *GophorError) {
    /* Create stdout, stderr buffers */
    outBuffer := &bytes.Buffer{}

    /* Setup command */
    var cmd *exec.Cmd
    if args != nil {
        cmd = exec.Command(path, args...)
    } else {
        cmd = exec.Command(path)
    }

    /* Setup cmd env */
    cmd.Env = env

    /* Setup out buffer */
    cmd.Stdout = outBuffer

    /* Start executing! */
    err := cmd.Start()
    if err != nil {
        return nil, &GophorError{ CommandStartErr, err }
    }

    /* Wait for command to finish, get exit code */
    err = cmd.Wait()
    exitCode := 0
    if err != nil {
        /* Error, try to get exit code */
        exitError, ok := err.(*exec.ExitError)
        if ok {
            waitStatus := exitError.Sys().(syscall.WaitStatus)
            exitCode = waitStatus.ExitStatus()
        } else {
            exitCode = 1
        }
    } else {
        /* No error! Get exit code direct from command */
        waitStatus := cmd.ProcessState.Sys().(syscall.WaitStatus)
        exitCode = waitStatus.ExitStatus()
    }
    
    if exitCode != 0 {
        /* If non-zero exit code return error */
        //errContents, gophorErr := readBuffer(errBuffer)
        Config.SysLog.Error("", "Error executing: %s\n", cmd.String())
        return nil, &GophorError{ CommandExitCodeErr, err }
    } else {
        /* If zero exit code try return outContents and no error */
        outContents, gophorErr := readBuffer(outBuffer)
        if gophorErr != nil {
            /* Failed fetching outContents, return error */
            return nil, gophorErr
        }

        return outContents, nil
    }
}

func readBuffer(reader *bytes.Buffer) ([]byte, *GophorError) {
    var err error
    var count int
    contents := make([]byte, 0)
    buf := make([]byte, ReaderBufSize)

    for {
        count, err = reader.Read(buf)
        if err != nil {
            if err == io.EOF {
                break
            }

            return nil, &GophorError{ BufferReadErr, err }
        }

        contents = append(contents, buf[:count]...)

        if count < ReaderBufSize {
            break
        }
    }

    return contents, nil
}


func envKeyValue(key, value string) string {
    return key+"="+value
}
