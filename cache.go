package main

import (
    "os"
    "sync"
    "time"
)

var (
    FileMonitorSleepTime = time.Duration(*CacheCheckFreq) * time.Second

    /* Global file caches */
    GlobalFileCache *FileCache
)

func startFileCaching() {
    /* Create gophermap file cache */
    GlobalFileCache = new(FileCache)
    GlobalFileCache.Init(*CacheSize)

    /* Start file monitor in separate goroutine */
    go startFileMonitor()
}

func startFileMonitor() {
    go func() {
        for {
            /* Sleep so we don't take up all the precious CPU time :) */
            time.Sleep(5 * time.Second)

            /* Check global file cache freshness */
            checkCacheFreshness()
        }

        /* We shouldn't have reached here */
        logSystemFatal("FileCache monitor escaped run loop!\n")
    }()
}

func checkCacheFreshness() {
    /* Before anything, get cache write lock (in case we have to delete) */
    GlobalFileCache.CacheMutex.Lock()

    /* Iterate through paths in cache map to query file last modified times */
    for path := range GlobalFileCache.CacheMap.Map {
        stat, err := os.Stat(path)
        if err != nil {
            /* Log file as not in cache, then delete */
            logSystemError("Failed to stat file in cache: %s\n", path)
            GlobalFileCache.CacheMap.Remove(path)
            continue
        }
        timeModified := stat.ModTime().UnixNano()

        /* Get file pointer, no need for lock as we have write lock */
        file := GlobalFileCache.CacheMap.Get(path)

        /* If the file is marked as fresh, but file on disk newer, mark as unfresh */
        if file.IsFresh() && file.LastRefresh() < timeModified {
            file.SetUnfresh()
        }
    }

    /* Done! We can release cache read lock */
    GlobalFileCache.CacheMutex.Unlock()
}

/* TODO: see if there is more efficienct setup */
type FileCache struct {
    CacheMap   *FixedMap
    CacheMutex sync.RWMutex
}

func (fc *FileCache) Init(size int) {
    fc.CacheMap = NewFixedMap(size)
    fc.CacheMutex = sync.RWMutex{}
}

func (fc *FileCache) FetchRegular(path string) ([]byte, *GophorError) {
    return fc.Fetch(path, func(path string) FileContents {
        contents := new(RegularFileContents)
        contents.path = path
        return contents
    })
}

func (fc *FileCache) FetchGophermap(path string) ([]byte, *GophorError) {
    return fc.Fetch(path, func(path string) FileContents {
        contents := new(GophermapContents)
        contents.path = path
        return contents
    })
}

func (fc *FileCache) Fetch(path string, newFileContents func(string) FileContents) ([]byte, *GophorError) {
    /* Get cache map read lock then check if file in cache map */
    fc.CacheMutex.RLock()
    file := fc.CacheMap.Get(path)

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
        stat, err := os.Stat(path)
        if err != nil {
            /* Error stat'ing file, unlock read mutex then return error */
            fc.CacheMutex.RUnlock()
            return nil, &GophorError{ FileStatErr, err }
        }

        /* Create new file contents object using supplied function */
        contents := newFileContents(path)

        /* Create new file wrapper around contents */
        file := NewFile(contents)

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
        if float64(stat.Size()) / BytesInMegaByte > *CacheFileSizeMax {
            b := file.Contents()
            fc.CacheMutex.RUnlock()
            return b, nil
        }

        /* File not in cache -- Swap cache map read for write lock.
         * NOTE: Here we don't need a list mutex lock as it is impossible
         *       for any other goroutine to get this lock while we have a cache
         *       _write_ lock. Due to the way this Fetch() function is written
         */
        fc.CacheMutex.RUnlock()
        fc.CacheMutex.Lock()

        /* Put file in the FixedMap */
        fc.CacheMap.Put(path, file)

        /* Before unlocking cache mutex, lock file read for upcoming call to .Contents() */
        file.RLock()

        /* Swap cache lock back to read */
        fc.CacheMutex.Unlock()
        fc.CacheMutex.RLock()
    }

    /* Read file contents into new variable for return, then unlock file read lock */
    b := file.Contents()
    file.RUnlock()

    /* Finally we can unlock the cache map read lock, we are done :) */
    fc.CacheMutex.RUnlock()

    return b, nil
}
