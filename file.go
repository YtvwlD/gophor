package main

import (
    "os"
    "io"
    "bufio"
    "sync"
    "time"
)

type RegularFile struct {
    path        string
    contents    []byte
    mutex       sync.RWMutex
    isFresh     bool
    lastRefresh int64
}

func NewRegularFile(path string) *RegularFile {
    f := new(RegularFile)
    f.path = path
    f.mutex = sync.RWMutex{}
    return f
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

    /* Update lastRefresh time + set fresh */
    f.lastRefresh = time.Now().UnixNano()
    f.isFresh = true

    return gophorErr
}

func (f *RegularFile) IsFresh() bool {
    return f.isFresh
}

func (f *RegularFile) SetUnfresh() {
    f.isFresh = false
}

func (f *RegularFile) LastRefresh() int64 {
    return f.lastRefresh
}

func (f *RegularFile) Lock() {
    f.mutex.Lock()
}

func (f *RegularFile) Unlock() {
    f.mutex.Unlock()
}

func (f *RegularFile) RLock() {
    f.mutex.RLock()
}

func (f *RegularFile) RUnlock() {
    f.mutex.RUnlock()
}

func bufferedRead(path string) ([]byte, *GophorError) {
    /* Open file */
    fd, err := os.Open(path)
    if err != nil {
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

            return nil, &GophorError{ FileReadErr, err }
        }

        contents = append(contents, buf[:count]...)

        if count < FileReadBufSize {
            break
        }
    }

    return contents, nil
}
