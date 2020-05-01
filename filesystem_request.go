package main

import (
    "path"
    "strings"
)

/* FileSystemRequest:
 * Makes a request to the filesystem either through
 * the FileCache or directly to a function like listDir().
 * It carries the requested filesystem path and any extra
 * needed information, for the moment just a set of details
 * about the virtual host.. Opens things up a lot more for
 * the future :)
 */
type FileSystemRequest struct {
    Path *RequestPath
    Host *ConnHost
}

type RequestPath struct {
    Root    string
    Path    string
    AbsPath string
}

func NewRequestPath(root, request string) *RequestPath {
    /* Here we must sanitize the request path. Start with a clean :) */
    requestPath := path.Clean(request)

    if path.IsAbs(requestPath) {
        /* Is absolute. Try trimming root and leading '/' */
        requestPath = strings.TrimPrefix(strings.TrimPrefix(requestPath, root), "/")
    } else {
        /* Is relative. If back dir traversal, give them root */
        if strings.HasPrefix(requestPath, "..") {
            requestPath = ""
        }
    }

    return &RequestPath{ root, requestPath, "" }
}

func (rp *RequestPath) SelectorPath() string {
    if rp.Path == "." {
        return "/"
    } else {
        return "/"+rp.Path
    }
}

func (rp *RequestPath) AbsolutePath() string {
    if rp.AbsPath == "" {
        rp.AbsPath = path.Join(rp.Root, rp.Path)
    }
    return rp.AbsPath
}

func (rp *RequestPath) RelativePath() string {
    return rp.Path
}

func (rp *RequestPath) SelectorPathJoin(extPath string) string {
    if rp.Path == "." {
        return path.Join("/", extPath)
    } else {
        return "/"+path.Join(rp.Path, extPath)
    }
}

func (rp *RequestPath) JoinAbsolutePath(extPath string) string {
    return path.Join(rp.AbsolutePath(), extPath)
}

func (rp *RequestPath) JoinRelativePath(extPath string) string {
    return path.Join(rp.Path, extPath)
}

func (rp *RequestPath) HasAbsoluteSuffix(suffix string) bool {
    return strings.HasSuffix(rp.AbsolutePath(), suffix)
}

func (rp *RequestPath) HasRelativeSuffix(suffix string) bool {
    return strings.HasSuffix(rp.Path, suffix)
}

func (rp *RequestPath) TrimRelativeSuffix(suffix string) string {
    return strings.TrimSuffix(rp.Path, suffix)
}

func (rp *RequestPath) TrimAbsoluteSuffix(suffix string) string {
    return strings.TrimSuffix(rp.AbsolutePath(), suffix)
}

func (rp *RequestPath) NewJoinedPathFromCurrent(extPath string) *RequestPath {
    /* DANGER THIS DOES NOT CHECK FOR BACK-DIR TRAVERSALS */
    return &RequestPath{ rp.Root, rp.JoinRelativePath(extPath), "" }
}

func (rp *RequestPath) NewTrimmedPathFromCurrent(trimSuffix string) *RequestPath {
    return &RequestPath{ rp.Root, rp.TrimRelativeSuffix(trimSuffix), "" }
}

func (rp *RequestPath) NewPathAtRoot(extPath string) *RequestPath {
    /* Sanitized and safe (hopefully) */
    return NewRequestPath(rp.Root, extPath)
}
