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
    AbsPath string /* Cache the absolute path */
}

func NewSanitizedRequestPath(root, request string) *RequestPath {
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

    return NewRequestPath(root, requestPath)
}

func NewRequestPath(root, relative string) *RequestPath {
    return &RequestPath{ root, relative, path.Join(root, relative) }
}

func (rp *RequestPath) SelectorPath() string {
    if rp.Path == "." {
        return "/"
    } else {
        return "/"+rp.Path
    }
}

func (rp *RequestPath) AbsolutePath() string {
    return rp.AbsPath
}

func (rp *RequestPath) RelativePath() string {
    return rp.Path
}

func (rp *RequestPath) JoinSelectorPath(extPath string) string {
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
    return path.Join(rp.RelativePath(), extPath)
}

func (rp *RequestPath) HasAbsolutePrefix(prefix string) bool {
    return strings.HasPrefix(rp.AbsolutePath(), prefix)
}

func (rp *RequestPath) HasRelativePrefix(prefix string) bool {
    return strings.HasPrefix(rp.RelativePath(), prefix)
}

func (rp *RequestPath) HasRelativeSuffix(suffix string) bool {
    return strings.HasSuffix(rp.RelativePath(), suffix)
}

func (rp *RequestPath) HasAbsoluteSuffix(suffix string) bool {
    return strings.HasSuffix(rp.AbsolutePath(), suffix)
}

func (rp *RequestPath) TrimRelativeSuffix(suffix string) string {
    return strings.TrimSuffix(rp.RelativePath(), suffix)
}

func (rp *RequestPath) TrimAbsoluteSuffix(suffix string) string {
    return strings.TrimSuffix(rp.AbsolutePath(), suffix)
}

func (rp *RequestPath) JoinPathFromRoot(extPath string) string {
    return path.Join(rp.Root, extPath)
}

func (rp *RequestPath) NewJoinPathFromCurrent(extPath string) *RequestPath {
    /* DANGER THIS DOES NOT CHECK FOR BACK-DIR TRAVERSALS */
    return NewRequestPath(rp.Root, rp.JoinRelativePath(extPath))
}

func (rp *RequestPath) NewTrimPathFromCurrent(trimSuffix string) *RequestPath {
    /* DANGER THIS DOES NOT CHECK FOR BACK-DIR TRAVERSALS */
    return NewRequestPath(rp.Root, rp.TrimRelativeSuffix(trimSuffix))
}

func (rp *RequestPath) NewPathAtRoot(extPath string) *RequestPath {
    /* Sanitized and safe (hopefully) */
    return NewSanitizedRequestPath(rp.Root, extPath)
}
