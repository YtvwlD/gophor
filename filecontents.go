package main

import (
    "bytes"
    "bufio"
    "strings"
)

/* RegularFileContents:
 * Very simple implementation of FileContents that just
 * buffered reads from the stored file path, stores the
 * read bytes in a slice and returns when requested.
 */
type RegularFileContents struct {
    path     string
    contents []byte
}

func (fc *RegularFileContents) Render(request *FileSystemRequest) []byte {
    /* Here we can ignore the extra data in request.
     * We are but a simple cache'd file
     */
    return fc.contents
}

func (fc *RegularFileContents) Load() *GophorError {
    /* Load the file into memory */
    var gophorErr *GophorError
    fc.contents, gophorErr = bufferedRead(fc.path)
    return gophorErr
}

func (fc *RegularFileContents) Clear() {
    fc.contents = nil
}

/* GophermapContents:
 * Implementation of FileContents that reads and
 * parses a gophermap file into a slice of gophermap
 * sections, then renders and returns these sections
 * when requested.
 */
type GophermapContents struct {
    path     string
    sections []GophermapSection
}

func (gc *GophermapContents) Render(request *FileSystemRequest) []byte {
    /* We don't just want to read the contents, each section
     * in the sections slice needs a call to render() to
     * perform their own required actions in producing a
     * sendable byte slice.
     */
    returnContents := make([]byte, 0)
    for _, line := range gc.sections {
        content, gophorErr := line.Render(request)
        if gophorErr != nil {
            content = buildInfoLine(GophermapRenderErrorStr)
        }
        returnContents = append(returnContents, content...)
    }

    /* Finally we end render with last line */
    returnContents = append(returnContents, []byte(LastLine)...)

    return returnContents
}

func (gc *GophermapContents) Load() *GophorError {
    /* Load the gophermap into memory as gophermap sections */
    var gophorErr *GophorError
    gc.sections, gophorErr = readGophermap(gc.path)
    return gophorErr
}

func (gc *GophermapContents) Clear() {
    gc.sections = nil
}

/* GophermapSection:
 * Provides an interface for different stored sections
 * of a gophermap file, whether it's static text that we
 * may want stored as-is, or the data required for a dir
 * listing or command executed that we may want updated
 * upon each file cache request.
 */
type GophermapSection interface {
    Render(*FileSystemRequest) ([]byte, *GophorError)
}

/* GophermapText:
 * Simple implementation of GophermapSection that holds
 * onto a static section of text as a slice of bytes.
 */
type GophermapText struct {
    contents []byte
}

func NewGophermapText(contents []byte) *GophermapText {
    s := new(GophermapText)
    s.contents = contents
    return s
}

func (s *GophermapText) Render(request *FileSystemRequest) ([]byte, *GophorError) {
    return replaceStrings(string(s.contents), request.Host), nil
}

/* GophermapDirListing:
 * An implementation of GophermapSection that holds onto a
 * path and a requested list of hidden files, then enumerates
 * the supplied paths (ignoring hidden files) when the content
 * Render() call is received.
 */
type GophermapDirListing struct {
    path   string
    Hidden map[string]bool
}

func NewGophermapDirListing(path string) *GophermapDirListing {
    s := new(GophermapDirListing)
    s.path = path
    return s
}

func (s *GophermapDirListing) Render(request *FileSystemRequest) ([]byte, *GophorError) {
    /* We could just pass the request directly, but in case the request
     * path happens to differ for whatever reason we create a new one
     */
    return listDir(&FileSystemRequest{ s.path, request.Host }, s.Hidden)
}

func readGophermap(path string) ([]GophermapSection, *GophorError) {
    /* Create return slice */
    sections := make([]GophermapSection, 0)

    /* _Create_ hidden files map now in case dir listing requested */
    hidden := make(map[string]bool)

    /* Keep track of whether we've already come across a title line (only 1 allowed!) */
    titleAlready := false

    /* Reference directory listing now in case requested */
    var dirListing *GophermapDirListing

    /* Perform buffered scan with our supplied splitter and iterators */
    gophorErr := bufferedScan(path,
        func(scanner *bufio.Scanner) bool {
            line := scanner.Text()

            /* Parse the line item type and handle */
            lineType := parseLineType(line)
            switch lineType {
                case TypeInfoNotStated:
                    /* Append TypeInfo to the beginning of line */
                    sections = append(sections, NewGophermapText(buildInfoLine(line)))

                case TypeTitle:
                    /* Reformat title line to send as info line with appropriate selector */
                    if !titleAlready {
                        sections = append(sections, NewGophermapText(buildLine(TypeInfo, line[1:], "TITLE", NullHost, NullPort)))
                        titleAlready = true
                    }

                case TypeComment:
                    /* We ignore this line */
                    break

                case TypeHiddenFile:
                    /* Add to hidden files map */
                    hidden[line[1:]] = true

                case TypeSubGophermap:
                    /* Check if we've been supplied subgophermap or regular file */
                    if strings.HasSuffix(line[1:], GophermapFileStr) {
                        /* Ensure we haven't been passed the current gophermap. Recursion bad! */
                        if line[1:] == path {
                            break
                        }

                        /* Treat as any other gopher map! */
                        submapSections, gophorErr := readGophermap(line[1:])
                        if gophorErr != nil {
                            /* Failed to read subgophermap, insert error line */
                            sections = append(sections, NewGophermapText(buildInfoLine("Error reading subgophermap: "+line[1:])))
                        } else {
                            sections = append(sections, submapSections...)
                        }
                    } else {
                        /* Treat as regular file, but we need to replace Unix line endings
                         * with gophermap line endings
                         */
                        fileContents, gophorErr := readIntoGophermap(line[1:])
                        if gophorErr != nil {
                            /* Failed to read file, insert error line */
                            Config.LogSystem("Error: %s\n", gophorErr)
                            sections = append(sections, NewGophermapText(buildInfoLine("Error reading subgophermap: "+line[1:])))
                        } else {
                            sections = append(sections, NewGophermapText(fileContents))
                        }
                    }

                case TypeExec:
                    /* Try executing supplied line */
                    sections = append(sections, NewGophermapText(buildInfoLine("Error: inline shell commands not yet supported")))

                case TypeEnd:
                    /* Lastline, break out at end of loop. Interface method Contents()
                     * will append a last line at the end so we don't have to worry about
                     * that here, only stopping the loop.
                     */
                    return false

                case TypeEndBeginList:
                    /* Create GophermapDirListing object then break out at end of loop */
                    dirListing = NewGophermapDirListing(strings.TrimSuffix(path, GophermapFileStr))
                    return false

                default:
                    /* Just append to sections slice as gophermap text */
                    sections = append(sections, NewGophermapText([]byte(line+DOSLineEnd)))
            }
            
            return true
        },
    )

    /* Check the bufferedScan didn't exit with error */
    if gophorErr != nil {
        return nil, gophorErr
    }

    /* If dir listing requested, append the hidden files map then add
     * to sections slice. We can do this here as the TypeEndBeginList item
     * type ALWAYS comes last, at least in the gophermap handled by this call
     * to readGophermap().
     */
    if dirListing != nil {
        dirListing.Hidden = hidden
        sections = append(sections, dirListing)
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
            line = strings.Replace(line, "\n", "", -1)

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

func replaceStrings(str string, connHost *ConnHost) []byte {
    str = strings.Replace(str, ReplaceStrHostname, connHost.Name, -1)
    str = strings.Replace(str, ReplaceStrPort, connHost.Port, -1)
    return []byte(str)
}
