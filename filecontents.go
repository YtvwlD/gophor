package main

import (
    "bytes"
    "bufio"
    "os"
)

type FileContents interface {
    /* Interface that provides an adaptable implementation
     * for holding onto some level of information about the
     * contents of a file.
     */
    Render(*FileSystemRequest) []byte
    Load()                     *GophorError
    Clear()
}

type GeneratedFileContents struct {
    Contents []byte /* Generated file contents as byte slice */
}

func (fc *GeneratedFileContents) Render(request *FileSystemRequest) []byte {
    return fc.Contents
}

func (fc *GeneratedFileContents) Load() *GophorError {
    /* do nothing */
    return nil
}

func (fc *GeneratedFileContents) Clear() {
    /* do nothing */
}

type RegularFileContents struct {
    Request  *FileSystemRequest /* Stored filesystem request */
    Contents []byte             /* File contents as byte slice */
}

func (fc *RegularFileContents) Render(request *FileSystemRequest) []byte {
    return fc.Contents
}

func (fc *RegularFileContents) Load() *GophorError {
    /* Load the file into memory */
    var gophorErr *GophorError
    fc.Contents, gophorErr = bufferedRead(fc.Request.AbsPath())
    return gophorErr
}

func (fc *RegularFileContents) Clear() {
    fc.Contents = nil
}

type GophermapContents struct {
    Request  *FileSystemRequest /* Stored filesystem request */
    Sections []GophermapSection /* Slice to hold differing gophermap sections */
}

func (gc *GophermapContents) Render(request *FileSystemRequest) []byte {
    returnContents := make([]byte, 0)

    /* Render each of the gophermap sections into byte slices */
    for _, line := range gc.Sections {
        content, gophorErr := line.Render(request)
        if gophorErr != nil {
            content = buildInfoLine(GophermapRenderErrorStr)
        }
        returnContents = append(returnContents, content...)
    }

    /* The footer added later contains last line, don't need to worry */

    return returnContents
}

func (gc *GophermapContents) Load() *GophorError {
    /* Load the gophermap into memory as gophermap sections */
    var gophorErr *GophorError
    gc.Sections, gophorErr = readGophermap(gc.Request)
    return gophorErr
}

func (gc *GophermapContents) Clear() {
    gc.Sections = nil
}

type GophermapSection interface {
    /* Interface for storing differring types of gophermap
     * sections and render when necessary
     */

    Render(*FileSystemRequest) ([]byte, *GophorError)
}

type GophermapText struct {
    Contents []byte /* Text contents */
}

func (s *GophermapText) Render(request *FileSystemRequest) ([]byte, *GophorError) {
    return replaceStrings(string(s.Contents), request.Host), nil
}

type GophermapDirListing struct {
    Request *FileSystemRequest /* Stored filesystem request */
    Hidden  map[string]bool    /* Hidden files map parsed from gophermap */
}

func (g *GophermapDirListing) Render(request *FileSystemRequest) ([]byte, *GophorError) {
    /* Create new filesystem request from mixture of stored + supplied */
    return listDir(
        &FileSystemRequest{
            request.Host,
            request.Client,
            g.Request.RootDir,
            g.Request.RelPath(),
            g.Request.AbsPath(),
            g.Request.Parameters,
        },
        g.Hidden,
    )
}

type GophermapExecCgi struct {
    Request *FileSystemRequest /* Stored file system request */
}

func (g *GophermapExecCgi) Render(request *FileSystemRequest) ([]byte, *GophorError) {
    /* Create new filesystem request from mixture of stored + supplied */
    return executeCgi(g.Request)
}

type GophermapExecFile struct {
    Request *FileSystemRequest /* Stored file system request */
}

func (g *GophermapExecFile) Render(request *FileSystemRequest) ([]byte, *GophorError) {
    return executeCommand(g.Request)
}

type GophermapExecCommand struct {
    Request *FileSystemRequest
}

func (g *GophermapExecCommand) Render(request *FileSystemRequest) ([]byte, *GophorError) {
    return executeCommand(g.Request)
}

func readGophermap(request *FileSystemRequest) ([]GophermapSection, *GophorError) {
    /* Create return slice */
    sections := make([]GophermapSection, 0)

    /* _Create_ hidden files map now in case dir listing requested */
    hidden := make(map[string]bool)

    /* Keep track of whether we've already come across a title line (only 1 allowed!) */
    titleAlready := false

    /* Reference directory listing now in case requested */
    var dirListing *GophermapDirListing

    /* Perform buffered scan with our supplied splitter and iterators */
    gophorErr := bufferedScan(request.AbsPath(),
        func(scanner *bufio.Scanner) bool {
            line := scanner.Text()

            /* Parse the line item type and handle */
            lineType := parseLineType(line)
            switch lineType {
                case TypeInfoNotStated:
                    /* Append TypeInfo to the beginning of line */
                    sections = append(sections, &GophermapText{ buildInfoLine(line) })

                case TypeTitle:
                    /* Reformat title line to send as info line with appropriate selector */
                    if !titleAlready {
                        sections = append(sections, &GophermapText{ buildLine(TypeInfo, line[1:], "TITLE", NullHost, NullPort) })
                        titleAlready = true
                    }

                case TypeComment:
                    /* We ignore this line */
                    break

                case TypeHiddenFile:
                    /* Add to hidden files map */
                    hidden[line[1:]] = true

                case TypeSubGophermap:
                    /* Parse new requestPath and parameters (this automatically sanitizes requestPath) */
                    subRequest := parseLineRequestString(request, line[1:])

                    if !subRequest.HasAbsPathPrefix("/") {
                        if Config.CgiEnabled {
                            /* Special case here where command must be in path, return GophermapExecCommand */
                            sections = append(sections, &GophermapExecCommand{ subRequest })
                        }
                    } else if subRequest.RelPath() == "" {
                        /* path cleaning failed */
                        break
                    } else if subRequest.RelPath() == request.RelPath() {
                        /* Same as current gophermap. Recursion bad! */
                        break
                    }

                    /* Perform file stat */
                    stat, err := os.Stat(subRequest.AbsPath())
                    if (err != nil) || (stat.Mode() & os.ModeDir != 0) {
                        /* File read error or is directory */
                        break
                    }

                    /* Check if we've been supplied subgophermap or regular file */
                    if subRequest.HasAbsPathSuffix("/"+GophermapFileStr) {
                        /* If executable, store as GophermapExecutable, else readGophermap() */
                        if stat.Mode().Perm() & 0100 != 0 && Config.CgiEnabled {
                            sections = append(sections, &GophermapExecFile { subRequest })
                        } else {
                            /* Treat as any other gophermap! */
                            submapSections, gophorErr := readGophermap(subRequest)
                            if gophorErr == nil {
                                sections = append(sections, submapSections...)
                            }
                        }
                    } else {
                        /* If stored in cgi-bin store as GophermapExecutable, else read into GophermapText */
                        if subRequest.HasRelPathPrefix(CgiBinDirStr) && Config.CgiEnabled {
                            sections = append(sections, &GophermapExecCgi{ subRequest })
                        } else {
                            fileContents, gophorErr := readIntoGophermap(subRequest.AbsPath())
                            if gophorErr == nil {
                                sections = append(sections, &GophermapText{ fileContents })
                            }
                        }
                    }

                case TypeEnd:
                    /* Lastline, break out at end of loop. Interface method Contents()
                     * will append a last line at the end so we don't have to worry about
                     * that here, only stopping the loop.
                     */
                    return false

                case TypeEndBeginList:
                    /* Create GophermapDirListing object then break out at end of loop */
                    dirRequest := NewFileSystemRequest(nil, nil, request.RootDir, request.TrimRelPathSuffix(GophermapFileStr), request.Parameters)
                    dirListing = &GophermapDirListing{ dirRequest, hidden }
                    return false

                default:
                    /* Just append to sections slice as gophermap text */
                    sections = append(sections, &GophermapText{ []byte(line+DOSLineEnd) })
            }

            return true
        },
    )

    /* Check the bufferedScan didn't exit with error */
    if gophorErr != nil {
        return nil, gophorErr
    }

    return sections, nil
}

func readIntoGophermap(path string) ([]byte, *GophorError) {
    /* Create return slice */
    fileContents := make([]byte, 0)

    /* Perform buffered scan with our supplied splitter and iterators */
    gophorErr := bufferedScan(path,
        func(scanner *bufio.Scanner) bool {
            line := scanner.Text()

            if line == "" {
                fileContents = append(fileContents, buildInfoLine("")...)
                return true
            }

            /* Replace the newline character */
            line = replaceNewLines(line)

            /* Iterate through returned str, reflowing to new line
             * until all lines < PageWidth
             */
            for len(line) > 0 {
                length := minWidth(len(line))
                fileContents = append(fileContents, buildInfoLine(line[:length])...)
                line = line[length:]
            }

            return true
        },
    )

    /* Check the bufferedScan didn't exit with error */
    if gophorErr != nil {
        return nil, gophorErr
    }

    /* Check final output ends on a newline */
    if !bytes.HasSuffix(fileContents, []byte(DOSLineEnd)) {
        fileContents = append(fileContents, []byte(DOSLineEnd)...)
    }

    return fileContents, nil
}

func minWidth(w int) int {
    if w <= Config.PageWidth {
        return w
    } else {
        return Config.PageWidth
    }
}
