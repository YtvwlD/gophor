package main

import (
    "os"
    "io"
    "bufio"
    "bytes"
    "strings"
    "sync"
    "time"
)

const GophermapFileStr = "gophermap"

type GophermapSection interface {
    Render() ([]byte, *GophorError)
}

type GophermapText struct {
    contents []byte

    /* Implements */
    GophermapSection
}

func NewGophermapText(contents string) *GophermapText {
    s := new(GophermapText)
    s.contents = []byte(contents)
    return s
}

func (s *GophermapText) Render() ([]byte, *GophorError) {
    return s.contents, nil
}

type GophermapDirListing struct {
    path   string
    Hidden map[string]bool

    /* Implements */
    GophermapSection
}

func NewGophermapDirListing(path string) *GophermapDirListing {
    s := new(GophermapDirListing)
    s.path = path
    return s
}

func (s *GophermapDirListing) Render() ([]byte, *GophorError) {
    return listDir(s.path, s.Hidden)
}

type GophermapFile struct {
    path        string
    lines       []GophermapSection
    mutex       *sync.RWMutex
    isFresh     bool
    lastRefresh int64

    /* Implements */
    File
}

func NewGophermapFile(path string) *GophermapFile {
    f := new(GophermapFile)
    f.path = path
    f.mutex = new(sync.RWMutex)
    return f
}

func (f *GophermapFile) Contents() []byte {
    /* We don't just want to read the contents,
     * but also execute any included gophermap
     * execute lines.
     */
    contents := make([]byte, 0)
    for _, line := range f.lines {
        content, gophorErr := line.Render()
        if gophorErr != nil {
            content = []byte(string(TypeInfo)+"Error rendering gophermap section."+CrLf)
        }
        contents = append(contents, content...)
    }

    return contents
}

func (f *GophermapFile) LoadContents() *GophorError {
    /* Clear the current cache */
    f.lines = nil

    /* Reload the file */
    var gophorErr *GophorError
    f.lines, gophorErr = f.readGophermap(f.path)

    /* Update lastRefresh + set fresh */
    f.lastRefresh = time.Now().UnixNano()
    f.isFresh = true

    return gophorErr
}

func (f *GophermapFile) IsFresh() bool {
    return f.isFresh
}

func (f *GophermapFile) SetUnfresh() {
    f.isFresh = false
}

func (f *GophermapFile) LastRefresh() int64 {
    return f.lastRefresh
}

func (f *GophermapFile) Lock() {
    f.mutex.Lock()
}

func (f *GophermapFile) Unlock() {
    f.mutex.Unlock()
}

func (f *GophermapFile) RLock() {
    f.mutex.RLock()
}

func (f *GophermapFile) RUnlock() {
    f.mutex.RUnlock()
}

func (f *GophermapFile) readGophermap(path string) ([]GophermapSection, *GophorError) {
    /* First, read raw file contents */
    contents, gophorErr := bufferedRead(path)
    if gophorErr != nil {
        return nil, gophorErr
    }

    /* Create reader and scanner from this */
    reader := bytes.NewReader(contents)
    scanner := bufio.NewScanner(reader)

    /* Setup scanner to split on CrLf */
    scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
        if atEOF && len(data) == 0  {
            /* At EOF, no more data */
            return 0, nil, nil
        }

        if i := bytes.Index(data, []byte{ '\r', '\n' }); i >= 0 {
            /* We have a full new-line terminate line */
            return i+2, data[0:i], nil
        }

        /* Request more data */
        return 0, nil, nil
    })

    /* Create return slice + hidden files map in case dir listing requested */
    sections := make([]GophermapSection, 0)
    hidden := make(map[string]bool)
    var dirListing *GophermapDirListing

    /* Scan, format each token and add to parsedContents */
    doEnd := false
    for scanner.Scan() {
        line := scanner.Text()

        /* Parse the line item type and handle */
        lineType := parseLineType(line)
        switch lineType {
            case TypeInfoNotStated:
                /* Append TypeInfo to the beginning of line */
                sections = append(sections, NewGophermapText(string(TypeInfo)+line+CrLf))

            case TypeComment:
                /* We ignore this line */
                continue

            case TypeHiddenFile:
                /* Add to hidden files map */
                hidden[line[1:]] = true

            case TypeSubGophermap:
                /* Check if we've been supplied subgophermap or regular file */
                if strings.HasSuffix(line[1:], GophermapFileStr) {
                    /* Ensure we haven't been passed the current gophermap. Recursion bad! */
                    if line[1:] == path {
                        continue
                    }

                    /* Treat as any other gopher map! */
                    submapSections, gophorErr := f.readGophermap(line[1:])
                    if gophorErr != nil {
                        /* Failed to read subgophermap, insert error line */
                        sections = append(sections, NewGophermapText(string(TypeInfo)+"Error reading subgophermap: "+line[1:]+CrLf))
                    } else {
                        sections = append(sections, submapSections...)
                    }
                } else {
                    /* Treat as regular file, but we need to replace Unix line endings
                     * with gophermap line endings
                     */
                    fileContents, gophorErr := bufferedReadAsGophermap(line[1:])
                    if gophorErr != nil {
                        /* Failed to read file, insert error line */
                        sections = append(sections, NewGophermapText(string(TypeInfo)+"Error reading subgophermap: "+line[1:]+CrLf))
                    } else {
                        sections = append(sections, NewGophermapText(string(fileContents)))
                    }
                }

            case TypeExec:
                /* Try executing supplied line */
                sections = append(sections, NewGophermapText(string(TypeInfo)+"Error: inline shell commands not yet supported"+CrLf))

            case TypeEnd:
                /* Lastline, break out at end of loop. Interface method Contents()
                 * will append a last line at the end so we don't have to worry about
                 * that here, only stopping the loop.
                 */
                doEnd = true

            case TypeEndBeginList:
                /* Create GophermapDirListing object then break out at end of loop */
                doEnd = true
                dirListing = NewGophermapDirListing(strings.TrimSuffix(path, GophermapFileStr))

            default:
                sections = append(sections, NewGophermapText(line+CrLf))
        }

        /* Break out of read loop if requested */
        if doEnd {
            break
        }
    }

    /* If scanner didn't finish cleanly, return nil and error */
    if scanner.Err() != nil {
        return nil, &GophorError{ FileReadErr, scanner.Err() }
    }

    /* If dir listing requested, append the hidden files map then add
     * to sections slice. We can do this here as the TypeEndBeginList item
     * type ALWAYS comes last, at least in the gophermap handled by this context.
     */
    if dirListing != nil {
        dirListing.Hidden = hidden
        sections = append(sections, dirListing)
    }

    return sections, nil
}

func bufferedReadAsGophermap(path string) ([]byte, *GophorError) {
    /* Open file */
    fd, err := os.Open(path)
    if err != nil {
        logSystemError("failed to open %s: %s\n", path, err.Error())
        return nil, &GophorError{ FileOpenErr, err }
    }
    defer fd.Close()

    /* Create reader and scanner from this */
    reader := bufio.NewReader(fd)
    fileContents := make([]byte, 0)

    for {
        str, err := reader.ReadString('\n')
        if err != nil {
            if err == io.EOF {
                /* Reached EOF */
                break
            }
            
            return nil, &GophorError{ FileReadErr, nil }
        }

        str = string(TypeInfo) + strings.Replace(str, "\n", CrLf, -1)
        fileContents = append(fileContents, []byte(str)...)
    }

    if !bytes.HasSuffix(fileContents, []byte(CrLf)) {
        fileContents = append(fileContents, []byte(CrLf)...)
    }

    return fileContents, nil
}
