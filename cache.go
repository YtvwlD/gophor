package main

import (
    "os"
    "sync"
    "time"
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
    Config.FileCache.CacheMutex.Lock()

    /* Iterate through paths in cache map to query file last modified times */
    for path := range Config.FileCache.CacheMap.Map {
        stat, err := os.Stat(path)
        if err != nil {
            /* Log file as not in cache, then delete */
            Config.LogSystemError("Failed to stat file in cache: %s\n", path)
            Config.FileCache.CacheMap.Remove(path)
            continue
        }
        timeModified := stat.ModTime().UnixNano()

        /* Get file pointer, no need for lock as we have write lock */
        file := Config.FileCache.CacheMap.Get(path)

        /* If the file is marked as fresh, but file on disk newer, mark as unfresh */
        if file.IsFresh() && file.LastRefresh() < timeModified {
            file.SetUnfresh()
        }
    }

    /* Done! We can release cache read lock */
    Config.FileCache.CacheMutex.Unlock()
}

type FileCache struct {
    CacheMap    *FixedMap
    CacheMutex  sync.RWMutex
    FileSizeMax int64
}

func (fc *FileCache) Init(size int, fileSizeMax float64) {
    fc.CacheMap    = NewFixedMap(size)
    fc.CacheMutex  = sync.RWMutex{}
    fc.FileSizeMax = int64(BytesInMegaByte * fileSizeMax)
}

func (fc *FileCache) FetchRegular(request *FileSystemRequest) ([]byte, *GophorError) {
    return fc.Fetch(request, func(path string) FileContents {
        contents := new(RegularFileContents)
        contents.path = path
        return contents
    })
}

func (fc *FileCache) FetchGophermap(request *FileSystemRequest) ([]byte, *GophorError) {
    return fc.Fetch(request, func(path string) FileContents {
        contents := new(GophermapContents)
        contents.path = path
        return contents
    })
}

func (fc *FileCache) Fetch(request *FileSystemRequest, newFileContents func(string) FileContents) ([]byte, *GophorError) {
    /* Get cache map read lock then check if file in cache map */
    fc.CacheMutex.RLock()
    file := fc.CacheMap.Get(request.Path)

    /* TODO: work on efficiency */
    if file != nil {
        /* File in cache -- before doing anything get file read lock */
        file.RLock()

        /* Check file is marked as fresh */
        if !file.IsFresh() {
            /* File not fresh! Swap file read for write-lock */
            file.RUnlock()
            file.Lock()

            /* Reload file contents from disk */
            gophorErr := file.LoadContents()
            if gophorErr != nil {
                /* Error loading contents, unlock all mutex then return error */
                file.Unlock()
                fc.CacheMutex.RUnlock()
                return nil, gophorErr
            }

            /* Updated! Swap back file write for read lock */
            file.Unlock()
            file.RLock()
        }
    } else {
        /* Before we do ANYTHING, we need to check file-size on disk */
        stat, err := os.Stat(request.Path)
        if err != nil {
            /* Error stat'ing file, unlock read mutex then return error */
            fc.CacheMutex.RUnlock()
            return nil, &GophorError{ FileStatErr, err }
        }

        /* Create new file contents object using supplied function */
        contents := newFileContents(request.Path)

        /* Create new file wrapper around contents */
        file = NewFile(contents)

        /* NOTE: file isn't in cache yet so no need to lock file write mutex
         * before loading from disk
         */
        gophorErr := file.LoadContents()
        if gophorErr != nil {
            /* Error loading contents, unlock read mutex then return error */
            fc.CacheMutex.RUnlock()
            return nil, gophorErr
        }

        /* Compare file size (in MB) to CacheFileSizeMax, if larger just get file
         * contents, unlock all mutex and don't bother caching. 
         */
        if stat.Size() > fc.FileSizeMax {
            b := file.Contents(request)
            fc.CacheMutex.RUnlock()
            return b, nil
        }

        /* File not in cache -- Swap cache map read for write lock. */
        fc.CacheMutex.RUnlock()
        fc.CacheMutex.Lock()

        /* Put file in the FixedMap */
        fc.CacheMap.Put(request.Path, file)

        /* Before unlocking cache mutex, lock file read for upcoming call to .Contents() */
        file.RLock()

        /* Swap cache lock back to read */
        fc.CacheMutex.Unlock()
        fc.CacheMutex.RLock()
    }

    /* Read file contents into new variable for return, then unlock file read lock */
    b := file.Contents(request)
    file.RUnlock()

    /* Finally we can unlock the cache map read lock, we are done :) */
    fc.CacheMutex.RUnlock()

    return b, nil
}
