package main

import (
    "io"
    "path"
    "strings"
)

type RequestPath struct {
    /* Path structure to allow hosts at
     * different roots while maintaining relative
     * and absolute path names for returned values
     * and filesystem reading
     */

    Root string
    Rel  string
    Abs  string
}

func NewRequestPath(rootDir, relPath string) *RequestPath {
    return &RequestPath{ rootDir, relPath, path.Join(rootDir, strings.TrimSuffix(relPath, "/")) }
}

func (rp *RequestPath) RootDir() string {
    return rp.Root
}

func (rp *RequestPath) Relative() string {
    return rp.Rel
}

func (rp *RequestPath) Absolute() string {
    return rp.Abs
}

func (rp *RequestPath) Selector() string {
    if rp.Rel == "." {
        return "/"
    } else {
        return "/"+rp.Rel
    }
}

type Request struct {
    /* A gophor request containing any data necessary.
     * Either handled through FileSystem or to direct function like listDir().
     */

    /* Can be nil */
    Host       *ConnHost
    Client     *ConnClient

    /* MUST be set */
    Writer     io.Writer
    Path       *RequestPath
    Parameters []string /* CGI-bin params will be 1 length slice, shell commands populate >=1 */ 
}

func NewSanitizedRequest(conn *GophorConn, requestStr string) *Request {
    /* Split dataStr into request path and parameter string (if pressent) */
    relPath, parameters := parseRequestString(requestStr)
    relPath = sanitizeRelativePath(conn.HostRoot(), relPath)
    return NewRequest(conn.Host, conn.Client, conn.Conn, NewRequestPath(conn.HostRoot(), relPath), parameters)
}

func NewRequest(host *ConnHost, client *ConnClient, writer io.Writer, path *RequestPath, parameters []string) *Request {
    return &Request{
        host,
        client,
        writer,
        path,
        parameters,
    }
}

func (r *Request) AccessLogInfo(format string, args ...interface{}) {
    /* You HAVE to be sure that r.Conn is NOT nil before calling this */
    Config.AccLog.Info("("+r.Client.Ip+") ", format, args...)
}

func (r *Request) AccessLogError(format string, args ...interface{}) {
    /* You HAVE to be sure that r.Conn is NOT nil before calling this */
    Config.AccLog.Error("("+r.Client.Ip+") ", format, args...)
}

func (r *Request) WriteRaw(reader io.Reader) *GophorError {
    /* You HAVE to be sure that r.Conn is NOT nil before calling this */
    _, err := io.Copy(r.Writer, reader)
    if err != nil {
        return &GophorError{ RequestWriteErr, err }
    } else {
        return nil
    }
}

func (r *Request) Write(data []byte) *GophorError {
    count, err := r.Writer.Write(data)
    if err != nil {
        return &GophorError{ RequestWriteErr, err }
    } else if count != len(data) {
        return &GophorError{ RequestWriteCountErr, nil }
    } else {
        return nil
    }
}

func (r *Request) RootDir() string {
    return r.Path.RootDir()
}

func (r *Request) AbsPath() string {
    return r.Path.Absolute()
}

func (r *Request) RelPath() string {
    return r.Path.Relative()
}

func (r *Request) SelectorPath() string {
    return r.Path.Selector()
}

func (r *Request) PathJoinSelector(extPath string) string {
    return path.Join(r.SelectorPath(), extPath)
}

func (r *Request) PathJoinAbs(extPath string) string {
    return path.Join(r.AbsPath(), extPath)
}

func (r *Request) PathJoinRel(extPath string) string {
    return path.Join(r.RelPath(), extPath)
}

func (r *Request) PathHasAbsPrefix(prefix string) bool {
    return strings.HasPrefix(r.AbsPath(), prefix)
}

func (r *Request) PathHasRelPrefix(prefix string) bool {
    return strings.HasPrefix(r.RelPath(), prefix)
}

func (r *Request) PathHasRelSuffix(suffix string) bool {
    return strings.HasSuffix(r.RelPath(), suffix)
}

func (r *Request) PathHasAbsSuffix(suffix string) bool {
    return strings.HasSuffix(r.AbsPath(), suffix)
}

func (r *Request) PathTrimRelSuffix(suffix string) string {
    return strings.TrimSuffix(strings.TrimSuffix(r.RelPath(), suffix), "/")
}

func (r *Request) PathTrimAbsSuffix(suffix string) string {
    return strings.TrimSuffix(strings.TrimSuffix(r.AbsPath(), suffix), "/")
}

func (r *Request) PathJoinRootDir(extPath string) string {
    return path.Join(r.Path.RootDir(), extPath)
}

func (r *Request) CachedRequest() *Request {
    return NewRequest(nil, nil, nil, r.Path, r.Parameters)
}

/* Sanitize a request path string */
func sanitizeRelativePath(rootDir, relPath string) string {
    /* Start with a clean :) */
    relPath = path.Clean(relPath)

    if path.IsAbs(relPath) {
        /* Is absolute. Try trimming root and leading '/' */
        relPath = strings.TrimPrefix(strings.TrimPrefix(relPath, rootDir), "/")
    } else {
        /* Is relative. If back dir traversal, give them root */
        if strings.HasPrefix(relPath, "..") {
            relPath = ""
        }
    }

    return relPath
}
