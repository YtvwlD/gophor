package main

import (
    "os/exec"
    "syscall"
    "bytes"
    "runtime"
    "io"
)

const (
    SafeExecPath = ""

    ReaderBufSize = 1024
)

func setupInitialCgiEnviron(description string) []string {
    return []string{
        /* RFC 3875 standard */
        envKeyValue("PATH",               SafeExecPath),
        envKeyValue("GATEWAY_INTERFACE",  ""),
        envKeyValue("SERVER_SOFTWARE",    "gophor "+GophorVersion),
        envKeyValue("SERVER_ARCH",        runtime.GOARCH),
        envKeyValue("SERVER_DESCRIPTION", description),
        envKeyValue("SERVER_VERSION",     GophorVersion),
        envKeyValue("SERVER_PROTOCOL",    "RFC1436"),
        envKeyValue("COLUMNS",            Config.PageWidth),
        envKeyValue("GOPHER_CHARSET",     Config.CharSet),
        envKeyValue("SERVER_CODENAME",    ""),
    }
}

func executeFile(requestPath *RequestPath, args []string) ([]byte, *GophorError) {
    /* Create stdout, stderr buffers */
    outBuffer := &bytes.Buffer{}
    errBuffer := &bytes.Buffer{}

    /* Setup command */
    var cmd *exec.Cmd
    if args != nil {
        cmd = exec.Command(requestPath.AbsolutePath(), args...)
    } else {
        cmd = exec.Command(requestPath.AbsolutePath())
    }

    /* Setup remaining CGI spec environment values */
    cmd.Env = Config.CgiEnv

    /* RFC 1436 standard */
    cmd.Env = append(cmd.Env, envKeyValue("CONTENT_LENGTH",  ""),
    cmd.Env = append(cmd.Env, envKeyValue("SERVER_NAME",     ""))
    cmd.Env = append(cmd.Env, envKeyValue("SERVER_PORT",     ""))
    cmd.Env = append(cmd.Env, envKeyValue("REQUEST_METHOD",  ""))
    cmd.Env = append(cmd.Env, envKeyValue("DOCUMENT_ROOT",   ""))
    cmd.Env = append(cmd.Env, envKeyValue("SCRIPT_NAME",     ""))
    cmd.Env = append(cmd.Env, envKeyValue("SCRIPT_FILENAME", ""))
    cmd.Env = append(cmd.Env, envKeyValue("LOCAL_ADDR",      ""))
    cmd.Env = append(cmd.Env, envKeyValue("REMOTE_ADDR",     ""))
    cmd.Env = append(cmd.Env, envKeyValue("SESSION_ID",      ""))
    cmd.Env = append(cmd.Env, envKeyValue("GOPHER_FILETYPE", ""))
    cmd.Env = append(cmd.Env, envKeyValue("GOPHER_REFERER",  ""))
    cmd.Env = append(cmd.Env, envKeyValue("SERVER_HOST",     ""))
    cmd.Env = append(cmd.Env, envKeyValue("REQUEST",         ""))
    cmd.Env = append(cmd.Env, envKeyValue("SEARCHREQUEST",   ""))
    cmd.Env = append(cmd.Env, envKeyValue("QUERY_STRING",    ""))
//    environ = append(environ, envKeyValue("TLS", "")
//    environ = append(environ, envKeyValue("SERVER_TLS_PORT", ""),
//    environ = append(environ, envKeyValue("HTTP_ACCEPT_CHARSET", "")
//    environ = append(environ, envKeyValue("HTTP_REFERER", "")

    /* Non-standard */
    environ = append(environ, envKeyValue("SELECTOR", ""))

    /* Set buffers*/
    cmd.Stdout = outBuffer
    cmd.Stderr = errBuffer

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
        /* If non-zero exit code return error, print stderr to sys log */
        errContents, gophorErr := readBuffer(errBuffer)
        if gophorErr == nil {
            /* Only print if we successfully fetched errContents */
            Config.SysLog.Error("", "Error executing: %s %v\n%s\n", requestPath.AbsolutePath(), args, errContents)
        }

        return nil, &GophorError{  }
    } else {
        /* If zero exit code try return outContents and no error */
        outContents, gophorErr := readBuffer(outBuffer)
        if gophorErr != nil {
            /* Failed fetching outContents, return error */
            return nil, &GophorError{ }
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
