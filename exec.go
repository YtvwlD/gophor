package main

import (
    "os/exec"
    "syscall"
    "bytes"
    "io"
)

const (
    /* Reader buf size */
    ReaderBufSize = 1024
)

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

    /* Setup operating environment */
    cmd.Env = Config.Env /* User defined cgi-bin environment */

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
