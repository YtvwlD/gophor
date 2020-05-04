package main

import (
    "path"
    "strings"
)

/* TODO: having 2 separate rootdir string values in Host and RootDir
 *       doesn't sit right with me. It cleans up code a lot for now
 *       but could get confusing. Figure out a more elegant way of
 *       structuring the filesystem request that gets passed around.
 */

type FileSystemRequest struct {
    /* A file system request with any possible required
     * data required. Either handled through FileSystem or to
     * direct function like listDir()
     */

    /* Virtual host and client information */
    Host       *ConnHost
    Client     *ConnClient

    /* File path information */
    RootDir    string
    Rel        string
    Abs        string

    /* Other parameters */
    Parameters []string /* CGI-bin params will be 1 length slice, shell commands populate >=1 */ 
}

func NewSanitizedFileSystemRequest(host *ConnHost, client *ConnClient, request string) *FileSystemRequest {
    /* Split dataStr into request path and parameter string (if pressent) */
    requestPath, parameters := parseRequestString(request)
    requestPath = sanitizeRequestPath(host.RootDir, requestPath)
    return NewFileSystemRequest(host, client, host.RootDir, requestPath, parameters)
}

func NewFileSystemRequest(host *ConnHost, client *ConnClient, rootDir, requestPath string, parameters []string) *FileSystemRequest {
    return &FileSystemRequest{
        host,
        client,
        rootDir,
        requestPath,
        path.Join(rootDir, requestPath),
        parameters,
    }
}

func (r *FileSystemRequest) SelectorPath() string {
    if r.Rel == "." {
        return "/"
    } else {
        return "/"+r.Rel
    }
}

func (r *FileSystemRequest) AbsPath() string {
    return r.Abs
}

func (r *FileSystemRequest) RelPath() string {
    return r.Rel
}

func (r *FileSystemRequest) JoinSelectorPath(extPath string) string {
    if r.Rel == "." {
        return path.Join("/", extPath)
    } else {
        return "/"+path.Join(r.Rel, extPath)
    }
}

func (r *FileSystemRequest) JoinAbsPath(extPath string) string {
    return path.Join(r.AbsPath(), extPath)
}

func (r *FileSystemRequest) JoinRelPath(extPath string) string {
    return path.Join(r.RelPath(), extPath)
}

func (r *FileSystemRequest) HasAbsPathPrefix(prefix string) bool {
    return strings.HasPrefix(r.AbsPath(), prefix)
}

func (r *FileSystemRequest) HasRelPathPrefix(prefix string) bool {
    return strings.HasPrefix(r.RelPath(), prefix)
}

func (r *FileSystemRequest) HasRelPathSuffix(suffix string) bool {
    return strings.HasSuffix(r.RelPath(), suffix)
}

func (r *FileSystemRequest) HasAbsPathSuffix(suffix string) bool {
    return strings.HasSuffix(r.AbsPath(), suffix)
}

func (r *FileSystemRequest) TrimRelPathSuffix(suffix string) string {
    return strings.TrimSuffix(strings.TrimSuffix(r.RelPath(), suffix), "/")
}

func (r *FileSystemRequest) TrimAbsPathSuffix(suffix string) string {
    return strings.TrimSuffix(strings.TrimSuffix(r.AbsPath(), suffix), "/")
}

func (r *FileSystemRequest) JoinPathFromRoot(extPath string) string {
    return path.Join(r.RootDir, extPath)
}

func (r *FileSystemRequest) NewStoredRequestAtRoot(relPath string, parameters []string) *FileSystemRequest {
    /* DANGER THIS DOES NOT CHECK FOR BACK-DIR TRAVERSALS */
    return NewFileSystemRequest(nil, nil, r.RootDir, relPath, parameters)
}

func (r *FileSystemRequest) NewStoredRequest() *FileSystemRequest {
    return NewFileSystemRequest(nil, nil, r.RootDir, r.RelPath(), r.Parameters)
}

/* Sanitize a request path string */
func sanitizeRequestPath(rootDir, requestPath string) string {
    /* Start with a clean :) */
    requestPath = path.Clean(requestPath)

    if path.IsAbs(requestPath) {
        /* Is absolute. Try trimming root and leading '/' */
        requestPath = strings.TrimPrefix(strings.TrimPrefix(requestPath, rootDir), "/")
    } else {
        /* Is relative. If back dir traversal, give them root */
        if strings.HasPrefix(requestPath, "..") {
            requestPath = ""
        }
    }

    return requestPath
}
