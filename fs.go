package main

import (
    "os"
    "sync"
    "path"
    "strings"
    "bytes"
    "time"
    "io"
    "bufio"
)

/* File:
 * Wraps around the cached contents of a file and
 * helps with management of this content by the
 * global FileCache objects.
 */
type File struct {
    contents    FileContents
    mutex       sync.RWMutex
    isFresh     bool
    lastRefresh int64
}

func NewFile(contents FileContents) *File {
    f := new(File)
    f.contents    = contents
    f.mutex       = sync.RWMutex{}
    f.isFresh     = true
    f.lastRefresh = 0
    return f
}

func (f *File) Contents() []byte {
    return f.contents.Render()
}

func (f *File) LoadContents() *GophorError {
    /* Clear current file contents */
    f.contents.Clear()

    /* Reload the file */
    gophorErr := f.contents.Load()
    if gophorErr != nil {
        return gophorErr
    }

    /* Update lastRefresh + set fresh */
    f.lastRefresh = time.Now().UnixNano()
    f.isFresh     = true

    return nil
}

func (f *File) IsFresh() bool {
    return f.isFresh
}

func (f *File) SetUnfresh() {
    f.isFresh = false
}

func (f *File) LastRefresh() int64 {
    return f.lastRefresh
}

func (f *File) Lock() {
    f.mutex.Lock()
}

func (f *File) Unlock() {
    f.mutex.Unlock()
}

func (f *File) RLock() {
    f.mutex.RLock()
}

func (f *File) RUnlock() {
    f.mutex.RUnlock()
}

/* FileContents:
 * Interface that provides an adaptable implementation
 * for holding onto some level of information about
 * the contents of a file, also methods for processing
 * and returning the results when the file contents
 * are requested.
 */
type FileContents interface {
    Render() []byte
    Load()   *GophorError
    Clear()
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

func bufferedScan(path string, scanIterator func(*bufio.Scanner) bool) *GophorError {
    /* First, read raw file contents */
    contents, gophorErr := bufferedRead(path)
    if gophorErr != nil {
        return gophorErr
    }

    /* Create reader and scanner from this */
    reader := bytes.NewReader(contents)
    scanner := bufio.NewScanner(reader)

    /* If contains DOS line-endings, use them! Else, Unix line-endings */
    if bytes.Contains(contents, []byte(DOSLineEnd)) {
        scanner.Split(dosLineEndSplitter)
    } else {
        scanner.Split(unixLineEndSplitter)
    }

    /* Scan through file contents using supplied iterator */
    for scanner.Scan() && scanIterator(scanner) {}

    /* Check scanner finished cleanly */
    if scanner.Err() != nil {
        return &GophorError{ FileReadErr, scanner.Err() }
    }

    return nil
}

func dosLineEndSplitter(data []byte, atEOF bool) (advance int, token []byte, err error) {
    if atEOF && len(data) == 0  {
        /* At EOF, no more data */
        return 0, nil, nil
    }

    if i := bytes.Index(data, []byte("\r\n")); i >= 0 {
        /* We have a full new-line terminate line */
        return i+2, data[:i], nil
    }

    /* Request more data */
    return 0, nil, nil
}

func unixLineEndSplitter(data []byte, atEOF bool) (advance int, token []byte, err error) {
    if atEOF && len(data) == 0  {
        /* At EOF, no more data */
        return 0, nil, nil
    }

    if i := bytes.Index(data, []byte("\n")); i >= 0 {
        /* We have a full new-line terminate line */
        return i+1, data[:i], nil
    }

    /* Request more data */
    return 0, nil, nil
}

func listDir(dirPath string, hidden map[string]bool) ([]byte, *GophorError) {
    /* Open directory file descriptor */
    fd, err := os.Open(dirPath)
    if err != nil {
        logSystemError("failed to open %s: %s\n", dirPath, err.Error())
        return nil, &GophorError{ FileOpenErr, err }
    }

    /* Open directory stream for reading */
    files, err := fd.Readdir(-1)
    if err != nil {
        logSystemError("failed to enumerate dir %s: %s\n", dirPath, err.Error())
        return nil, &GophorError{ DirListErr, err }
    }

    /* Walk through directory */
    dirContents := make([]byte, 0)

    /* First add a 'back' entry. GoLang Readdir() seems to miss this */
    line := buildLine(TypeDirectory, "..", path.Join(fd.Name(), ".."), *ServerHostname, *ServerPort)
    dirContents = append(dirContents, line...)

    /* Iterate through files :) */
    for _, file := range files {
        /* Skip dotfiles + gophermap file + requested hidden */
        if file.Name()[0] == '.' || strings.HasSuffix(file.Name(), GophermapFileStr) {
            continue
        }

        /* Handle file, directory or ignore others */
        switch {
            case file.Mode() & os.ModeDir != 0:
                /* Directory -- create directory listing */
                itemPath := path.Join(fd.Name(), file.Name())
                line = buildLine(TypeDirectory, file.Name(), itemPath, *ServerHostname, *ServerPort)
                dirContents = append(dirContents, line...)

            case file.Mode() & os.ModeType == 0:
                /* Regular file -- find item type and creating listing */
                itemPath := path.Join(fd.Name(), file.Name())
                itemType := getItemType(itemPath)
                if itemType == TypeBanned {
                    /* Banned file extension, skipping! */
                    break
                }
                line = buildLine(itemType, file.Name(), itemPath, *ServerHostname, *ServerPort)
                dirContents = append(dirContents, line...)

            default:
                /* Ignore */
        }
    }

    return dirContents, nil
}
