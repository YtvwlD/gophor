package main

import (
    "os"
    "io"
    "bufio"
)

type RegularFile struct {
    path     string
    contents []byte

    /* Implements */
    File
}

func (f *RegularFile) Contents() []byte {
    return f.contents
}

func (f *RegularFile) LoadContents() *GophorError {
    /* Clear current cache */
    f.contents = nil

    /* Reload the file */
    var gophorErr *GophorError
    f.contents, gophorErr = bufferedRead(f.path)
    return gophorErr
}

func bufferedRead(path string) ([]byte, *GophorError) {
    /* Open file */
    fd, err := os.Open(path)
    if err != nil {
        logSystemError("failed to open %s: %s\n", path, err.Error())
        return nil, &GophorError{ FileOpenErr, err }
    }
    defer fd.Close()

    /* Setup buffers */
    var count int
    contents := make([]byte, 0)
    buf := make([]byte, FileReadBufSize)

    /* Setup reader */
    reader := bufio.NewReader(fd)

    /* Read through buffer until error or null bytes! */
    for {
        count, err = reader.Read(buf)
        if err != nil {
            if err == io.EOF {
                break
            }

            logSystemError("failed to read %s: %s\n", path, err.Error())
            return nil, &GophorError{ FileReadErr, err }
        }

        contents = append(contents, buf[:count]...)

        if count < FileReadBufSize {
            break
        }
    }

    return contents, nil
}
