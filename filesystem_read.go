package main

import (
    "os"
    "bytes"
    "io"
    "sort"
    "bufio"
)

/* Perform simple buffered read on a file at path */
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

/* Perform buffered read on file at path, then scan through with supplied iterator func */
func bufferedScan(path string, scanIterator func(*bufio.Scanner) bool) *GophorError {
    /* First, read raw file contents */
    contents, gophorErr := bufferedRead(path)
    if gophorErr != nil {
        return gophorErr
    }

    /* Create reader and scanner from this */
    reader := bytes.NewReader(contents)
    scanner := bufio.NewScanner(reader)

    /* If contains DOS line-endings, split by DOS! Else, split by Unix */
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

/* listDir():
 * Here we use an empty function pointer, and set the correct
 * function to be used during the restricted files regex parsing.
 * This negates need to check if RestrictedFilesRegex is nil every
 * single call.
 */
var listDir func(request *FileSystemRequest, hidden map[string]bool) ([]byte, *GophorError)

func _listDir(request *FileSystemRequest, hidden map[string]bool) ([]byte, *GophorError) {
    return _listDirBase(request, func(dirContents *[]byte, file os.FileInfo) {
        /* If requested hidden */
        if _, ok := hidden[file.Name()]; ok {
            return
        }

        /* Handle file, directory or ignore others */
        switch {
            case file.Mode() & os.ModeDir != 0:
                /* Directory -- create directory listing */
                itemPath := request.Path.JoinSelectorPath(file.Name())
                *dirContents = append(*dirContents, buildLine(TypeDirectory, file.Name(), itemPath, request.Host.Name, request.Host.Port)...)

            case file.Mode() & os.ModeType == 0:
                /* Regular file -- find item type and creating listing */
                itemPath := request.Path.JoinSelectorPath(file.Name())
                itemType := getItemType(itemPath)
                *dirContents = append(*dirContents, buildLine(itemType, file.Name(), itemPath, request.Host.Name, request.Host.Port)...)

            default:
                /* Ignore */
        }
    })
}

func _listDirRegexMatch(request *FileSystemRequest, hidden map[string]bool) ([]byte, *GophorError) {
    return _listDirBase(request, func(dirContents *[]byte, file os.FileInfo) {
        /* If regex match in restricted files || requested hidden */
        if isRestrictedFile(file.Name()) {
            return
        } else if _, ok := hidden[file.Name()]; ok {
            return
        }

        /* Handle file, directory or ignore others */
        switch {
            case file.Mode() & os.ModeDir != 0:
                /* Directory -- create directory listing */
                itemPath := request.Path.JoinSelectorPath(file.Name())
                *dirContents = append(*dirContents, buildLine(TypeDirectory, file.Name(), itemPath, request.Host.Name, request.Host.Port)...)

            case file.Mode() & os.ModeType == 0:
                /* Regular file -- find item type and creating listing */
                itemPath := request.Path.JoinSelectorPath(file.Name())
                itemType := getItemType(itemPath)
                *dirContents = append(*dirContents, buildLine(itemType, file.Name(), itemPath, request.Host.Name, request.Host.Port)...)

            default:
                /* Ignore */
        }
    })
}

func _listDirBase(request *FileSystemRequest, iterFunc func(dirContents *[]byte, file os.FileInfo)) ([]byte, *GophorError) {
    /* Open directory file descriptor */
    fd, err := os.Open(request.Path.AbsolutePath())
    if err != nil {
        Config.SysLog.Error("", "failed to open %s: %s\n", request.Path, err.Error())
        return nil, &GophorError{ FileOpenErr, err }
    }

    /* Read files in directory */
    files, err := fd.Readdir(-1)
    if err != nil {
        Config.SysLog.Error("", "failed to enumerate dir %s: %s\n", request.Path, err.Error())
        return nil, &GophorError{ DirListErr, err }
    }

    /* Sort the files by name */
    sort.Sort(byName(files))

    /* Create directory content slice, ready */
    dirContents := make([]byte, 0)

    /* First add a title + a space */
    dirContents = append(dirContents, buildLine(TypeInfo, "[ "+request.Host.Name+request.Path.SelectorPath()+" ]", "TITLE", NullHost, NullPort)...)
    dirContents = append(dirContents, buildInfoLine("")...)

    /* Add a 'back' entry. GoLang Readdir() seems to miss this */
    dirContents = append(dirContents, buildLine(TypeDirectory, "..", request.Path.JoinRelativePath(".."), request.Host.Name, request.Host.Port)...)

    /* Walk through files :D */
    for _, file := range files { iterFunc(&dirContents, file) }

    return dirContents, nil
}

/* Took a leaf out of go-gopher's book here. */
type byName []os.FileInfo
func (s byName) Len() int           { return len(s) }
func (s byName) Less(i, j int) bool { return s[i].Name() < s[j].Name() }
func (s byName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
