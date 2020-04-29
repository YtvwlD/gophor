package main

import (
    "os"
    "sync"
    "time"
    "path"
    "strings"
)

func startFileMonitor(sleepTime time.Duration) {
    go func() {
        for {
            /* Sleep so we don't take up all the precious CPU time :) */
            time.Sleep(sleepTime)

            /* Check global file cache freshness */
            checkCacheFreshness()
        }

        /* We shouldn't have reached here */
        Config.LogSystemFatal("FileCache monitor escaped run loop!\n")
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
            Config.LogSystemError("Failed to stat file in cache: %s\n", path)
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
    switch file.contents.(type) {
       case *GeneratedFileContents:
           return true
       default:
           return false 
    }
}

/* FileCache:
 * Object to hold and help manage our file cache. Uses a fixed map
 * as a means of easily collecting files by path, but also being able
 * to remove cached files in a LRU style. Uses a RW mutex to lock the
 * cache map for appropriate functions and ensure thread safety.
 */
type FileSystem struct {
    CacheMap     *FixedMap
    CacheMutex   sync.RWMutex
    CacheFileMax int64
}

func (fs *FileSystem) Init(size int, fileSizeMax float64) {
    fs.CacheMap     = NewFixedMap(size)
    fs.CacheMutex   = sync.RWMutex{}
    fs.CacheFileMax = int64(BytesInMegaByte * fileSizeMax)
}

func (fs *FileSystem) HandleRequest(requestPath string, host *ConnHost) ([]byte, *GophorError) {
    /* Stat filesystem for request file type */
    fileType := FileTypeDir;
    if requestPath != "/" {
        stat, err := os.Stat(requestPath)
        if err != nil {
            /* Check file isn't in cache before throwing in the towel */
            fs.CacheMutex.RLock()
            file := fs.CacheMap.Get(requestPath)
            if file == nil {
                fs.CacheMutex.RUnlock()
                return nil, &GophorError{ FileStatErr, err }
            }

            /* It's there! Get contents, unlock and return */
            file.Mutex.RLock()
            b := file.Contents(&FileSystemRequest{ requestPath, host })
            file.Mutex.RUnlock()

            fs.CacheMutex.RUnlock()
            return b, nil
        }

        /* Set file type for later handling */
        switch {
            case stat.Mode() & os.ModeDir != 0:
                /* do nothing, already set :) */
                break

            case stat.Mode() & os.ModeType == 0:
                fileType = FileTypeRegular

            default:
                fileType = FileTypeBad
        }
    }

    switch fileType {
        /* Directory */
        case FileTypeDir:
            gophermapPath := path.Join(requestPath, GophermapFileStr)
            _, err := os.Stat(gophermapPath)
            if err == nil {
                /* Gophermap exists, serve this! */
                return fs.FetchFile(&FileSystemRequest{ gophermapPath, host })
            } else {
                /* No gophermap, serve directory listing */
                return listDir(&FileSystemRequest{ requestPath, host }, map[string]bool{})
            }

        /* Regular file */
        case FileTypeRegular:
            return fs.FetchFile(&FileSystemRequest{ requestPath, host })

        /* Unsupported type */
        default:
            return nil, &GophorError{ FileTypeErr, nil }
    }
}

func (fs *FileSystem) FetchFile(request *FileSystemRequest) ([]byte, *GophorError) {
    /* Get cache map read lock then check if file in cache map */
    fs.CacheMutex.RLock()
    file := fs.CacheMap.Get(request.Path)

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
        stat, err := os.Stat(request.Path)
        if err != nil {
            /* Error stat'ing file, unlock read mutex then return error */
            fs.CacheMutex.RUnlock()
            return nil, &GophorError{ FileStatErr, err }
        }

        /* Create new file contents object using supplied function */
        var contents FileContents
        if strings.HasSuffix(request.Path, "/"+GophermapFileStr) {
            contents = &GophermapContents{ request.Path, nil }
        } else {
            contents = &RegularFileContents{ request.Path, nil }
        }

        /* Create new file wrapper around contents */
        file = NewFile(contents)

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
        fs.CacheMap.Put(request.Path, file)

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
