package main

import (
    "os"
    "sync"
    "time"
)

const (
    /* Help converting file size stat to supplied size in megabytes */
    BytesInMegaByte = 1048576.0

    /* Filename constants */
    CgiBinDirStr     = "cgi-bin"
    GophermapFileStr = "gophermap"
)

type FileSystem struct {
    /* Holds and helps manage our file cache, as well as managing
     * access and responding to filesystem requests submitted by
     * a worker instance.
     */

    CacheMap     *FixedMap    /* Fixed size cache map */
    CacheMutex   sync.RWMutex /* RWMutex for safe cachemap access */
    CacheFileMax int64        /* Cache file size max */
}

func (fs *FileSystem) Init(size int, fileSizeMax float64) {
    fs.CacheMap     = NewFixedMap(size)
    fs.CacheMutex   = sync.RWMutex{}
    fs.CacheFileMax = int64(BytesInMegaByte * fileSizeMax)
}

func (fs *FileSystem) HandleRequest(request *FileSystemRequest) ([]byte, *GophorError) {
    /* Get filesystem stat, check it exists! */
    stat, err := os.Stat(request.AbsPath())
    if err != nil {
        /* Check file isn't in cache before throwing in the towel */
        fs.CacheMutex.RLock()
        file := fs.CacheMap.Get(request.AbsPath())
        if file == nil {
            fs.CacheMutex.RUnlock()
            return nil, &GophorError{ FileStatErr, err }
        }

        /* It's there! Get contents, unlock and return */
        file.Mutex.RLock()
        b := file.Contents(request)
        file.Mutex.RUnlock()

        fs.CacheMutex.RUnlock()
        return b, nil
    }

    /* Handle file type */
    switch {
        /* Directory */
        case stat.Mode() & os.ModeDir != 0:
            /* Ignore cgi-bin directory */
            if request.HasRelPathPrefix(CgiBinDirStr) {
                return nil, &GophorError{ IllegalPathErr, nil }
            }

            /* Check Gophermap exists */
            gophermapRequest := NewFileSystemRequest(request.Host, request.Client, request.RootDir, request.JoinRelPath(GophermapFileStr), request.Parameters)
            stat, err = os.Stat(gophermapRequest.AbsPath())

            var output []byte
            var gophorErr *GophorError
            if err == nil {
                /* Gophermap exists! If executable execute, else serve. */
                if stat.Mode().Perm() & 0100 != 0 {
                    output, gophorErr = executeFile(gophermapRequest)
                } else {
                    output, gophorErr = fs.FetchFile(gophermapRequest)
                }
            } else {
                /* No gophermap, serve directory listing */
                output, gophorErr = listDir(request, map[string]bool{})
            }

            if gophorErr != nil {
                /* Fail out! */
                return nil, gophorErr
            }

            /* Append footer text (contains last line) and return */
            output = append(output, Config.FooterText...)

            return output, nil

        /* Regular file */
        case stat.Mode() & os.ModeType == 0:
            /* If cgi-bin, return executed contents. Else, fetch */
            if request.HasRelPathPrefix(CgiBinDirStr) {
                return executeCgi(request)
            } else {
                return fs.FetchFile(request)
            }

        /* Unsupported type */
        default:
            return nil, &GophorError{ FileTypeErr, nil }
    }
}

func (fs *FileSystem) FetchFile(request *FileSystemRequest) ([]byte, *GophorError) {
    /* Get cache map read lock then check if file in cache map */
    fs.CacheMutex.RLock()
    file := fs.CacheMap.Get(request.AbsPath())

    if file != nil {
        /* File in cache -- before doing anything get file read lock */
        file.Mutex.RLock()

        /* Check file is marked as fresh */
        if !file.Fresh {
            /* File not fresh! Swap file read for write-lock */
            file.Mutex.RUnlock()
            file.Mutex.Lock()

            /* Reload file contents from disk */
            gophorErr := file.LoadContents()
            if gophorErr != nil {
                /* Error loading contents, unlock all mutex then return error */
                file.Mutex.Unlock()
                fs.CacheMutex.RUnlock()
                return nil, gophorErr
            }

            /* Updated! Swap back file write for read lock */
            file.Mutex.Unlock()
            file.Mutex.RLock()
        }
    } else {
        /* Perform filesystem stat ready for checking file size later.
         * Doing this now allows us to weed-out non-existent files early
         */
        stat, err := os.Stat(request.AbsPath())
        if err != nil {
            /* Error stat'ing file, unlock read mutex then return error */
            fs.CacheMutex.RUnlock()
            return nil, &GophorError{ FileStatErr, err }
        }

        /* Create new file contents */
        var contents FileContents
        if request.HasAbsPathSuffix("/"+GophermapFileStr) {
            contents = &GophermapContents{ request, nil }
        } else {
            contents = &RegularFileContents{ request, nil }
        }

        /* Create new file wrapper around contents */
        file = &File{ contents, sync.RWMutex{}, true, time.Now().UnixNano() }

        /* File isn't in cache yet so no need to get file lock mutex */
        gophorErr := file.LoadContents()
        if gophorErr != nil {
            /* Error loading contents, unlock read mutex then return error */
            fs.CacheMutex.RUnlock()
            return nil, gophorErr
        }

        /* Compare file size (in MB) to CacheFileSizeMax, if larger just get file
         * contents, unlock all mutex and don't bother caching. 
         */
        if stat.Size() > fs.CacheFileMax {
            b := file.Contents(request)
            fs.CacheMutex.RUnlock()
            return b, nil
        }

        /* File not in cache -- Swap cache map read for write lock. */
        fs.CacheMutex.RUnlock()
        fs.CacheMutex.Lock()

        /* Put file in the FixedMap */
        fs.CacheMap.Put(request.AbsPath(), file)

        /* Before unlocking cache mutex, lock file read for upcoming call to .Contents() */
        file.Mutex.RLock()

        /* Swap cache lock back to read */
        fs.CacheMutex.Unlock()
        fs.CacheMutex.RLock()
    }

    /* Read file contents into new variable for return, then unlock file read lock */
    b := file.Contents(request)
    file.Mutex.RUnlock()

    /* Finally we can unlock the cache map read lock, we are done :) */
    fs.CacheMutex.RUnlock()

    return b, nil
}

type File struct {
    /* Wraps around the cached contents of a file
     * helping with management of this content by
     * a FileSystem instance.
     */

    Content     FileContents
    Mutex       sync.RWMutex
    Fresh       bool
    LastRefresh int64
}

func (f *File) Contents(request *FileSystemRequest) []byte {
    return f.Content.Render(request)
}

func (f *File) LoadContents() *GophorError {
    /* Clear current file contents */
    f.Content.Clear()

    /* Reload the file */
    gophorErr := f.Content.Load()
    if gophorErr != nil {
        return gophorErr
    }

    /* Update lastRefresh, set fresh, unset deletion (not likely set) */
    f.LastRefresh = time.Now().UnixNano()
    f.Fresh       = true

    return nil
}

func startFileMonitor(sleepTime time.Duration) {
    go func() {
        for {
            /* Sleep so we don't take up all the precious CPU time :) */
            time.Sleep(sleepTime)

            /* Check global file cache freshness */
            checkCacheFreshness()
        }

        /* We shouldn't have reached here */
        Config.SysLog.Fatal("", "FileCache monitor escaped run loop!\n")
    }()
}

func checkCacheFreshness() {
    /* Before anything, get cache write lock (in case we have to delete) */
    Config.FileSystem.CacheMutex.Lock()

    /* Iterate through paths in cache map to query file last modified times */
    for path := range Config.FileSystem.CacheMap.Map {
        /* Get file pointer, no need for lock as we have write lock */
        file := Config.FileSystem.CacheMap.Get(path)

        /* If this is a generated file, we skip */
        if isGeneratedType(file) {
            continue
        }

        stat, err := os.Stat(path)
        if err != nil {
            /* Log file as not in cache, then delete */
            Config.SysLog.Error("", "Failed to stat file in cache: %s\n", path)
            Config.FileSystem.CacheMap.Remove(path)
            continue
        }
        timeModified := stat.ModTime().UnixNano()

        /* If the file is marked as fresh, but file on disk newer, mark as unfresh */
        if file.Fresh && file.LastRefresh < timeModified {
            file.Fresh = false
        }
    }

    /* Done! We can release cache read lock */
    Config.FileSystem.CacheMutex.Unlock()
}

func isGeneratedType(file *File) bool {
    /* Just a helper function to neaten-up checking if file contents is of generated type */
    switch file.Content.(type) {
       case *GeneratedFileContents:
           return true
       default:
           return false 
    }
}
