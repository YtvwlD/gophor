package main

import (
    "strings"
    "path"
    "net/url"
)

/* Parse line type from contents */
func parseLineType(line string) ItemType {
    lineLen := len(line)

    if lineLen == 0 {
        return TypeInfoNotStated
    } else if lineLen == 1 {
        /* The only accepted types for a length 1 line */
        switch ItemType(line[0]) {
            case TypeEnd:
                return TypeEnd
            case TypeEndBeginList:
                return TypeEndBeginList
            case TypeComment:
                return TypeComment
            case TypeInfo:
                return TypeInfo
            case TypeTitle:
                return TypeTitle
            default:
                return TypeUnknown
        }
    } else if !strings.Contains(line, string(Tab)) {
        /* The only accepted types for a line with no tabs */
        switch ItemType(line[0]) {
            case TypeComment:
                return TypeComment
            case TypeTitle:
                return TypeTitle
            case TypeInfo:
                return TypeInfo
            case TypeHiddenFile:
                return TypeHiddenFile
            case TypeSubGophermap:
                return TypeSubGophermap
            case TypeExec:
                return TypeExec
            default:
                return TypeInfoNotStated
        }
    }

    return ItemType(line[0])
}

/* Parses a line in a gophermap into a filesystem request path and a string slice of arguments */
func parseLineFileSystemRequest(rootDir, requestStr string) (*RequestPath, []string) {
    if path.IsAbs(requestStr) {
        /* This is an absolute path, assume it must be within gopher directory */
        args := splitLineStringArgs(requestStr)
        requestPath := NewSanitizedRequestPath(rootDir, args[0])

        if len(args) > 1 {
            return requestPath, args[1:]
        } else {
            return requestPath, nil
        }
    } else {
        /* Not an absolute path, if starts with cgi-bin treat as within gopher directory, else as command in path */
        if strings.HasPrefix(requestStr, CgiBinDirStr) {
            args := splitLineStringArgs(requestStr)
            requestPath := NewSanitizedRequestPath(rootDir, args[0])

            if len(args) > 1 {
                return requestPath, args[1:]
            } else {
                return requestPath, nil
            }
        } else {
            args := splitLineStringArgs(requestStr)

            /* Manually create specialised request path */
            requestPath := NewRequestPath(args[0], "")

            if len(args) > 1 {
                return requestPath, args[1:]
            } else {
                return requestPath, nil
            }
        }
    }
}

/* Parses a gopher request string into a filesystem request path and string slice of arguments */
func parseFileSystemRequest(rootDir, requestStr string) (*RequestPath, []string, *GophorError) {
    /* Split the request string */
    args := splitRequestStringArgs(requestStr)

    /* Now URL decode all the parts. */
    var err error
    for i := range args {
        args[i], err = url.QueryUnescape(args[i])
        if err != nil {
            return nil, nil, &GophorError{ InvalidRequestErr, err }
        }
    }

    /* Create request path */
    requestPath := NewSanitizedRequestPath(rootDir, args[0])

    /* Return request path and args if precent */
    if len(args) > 1 {
        return requestPath, args[1:], nil
    } else {
        return requestPath, nil, nil
    }
}

/* Parse new-line separated string of environment variables into a slice */
func parseEnvironmentString(env string) []string {
    return splitStringByRune(env, '\n')
}

/* Splits a request string into it's arguments with the '?' delimiter */
func splitRequestStringArgs(requestStr string) []string {
    return splitStringByRune(requestStr, '?')
}

/* Splits a line string into it's arguments with standard space delimiter */
func splitLineStringArgs(requestStr string) []string {
    split := Config.CmdParseLineRegex.Split(requestStr, -1)
    if split == nil {
        return []string{ requestStr }
    } else {
        return split
    }
}

/* Split a string according to a rune, that supports delimiting with '\' */
func splitStringByRune(str string, r rune) []string {
    ret := make([]string, 0)
    buf := ""
    delim := false
    for _, c := range str {
        switch c {
            case r:
                if !delim {
                    ret = append(ret, buf)
                    buf = ""
                } else {
                    buf += string(c)
                    delim = false
                }

            case '\\':
                if !delim {
                    delim = true
                } else {
                    buf += string(c)
                    delim = false
                }

            default:
                if !delim {
                    buf += string(c)
                } else {
                    buf += "\\"+string(c)
                    delim = false
                }
        }
    }

    if len(buf) > 0 || len(ret) == 0 {
        ret = append(ret, buf)
    }

    return ret
}
