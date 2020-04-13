package main

import (
    "os"
    "sync"
    "time"
)

const (
    FileMonitorSleepTimeMs = 30
    FileMonitorSleepTime = time.Duration(FileMonitorSleepTimeMs) * time.Second
)

func StartFileMonitor(regularCache *RegularFileCache, gophermapCache *GophermapFileCache) {
    go func() {
        for {
            /* Check regular file cache is fresh */
            regularCache.Mutex.RLock()
            for path := range regularCache.CacheMap {
                /* Check if file on disk has changed. */
                stat, err := os.Stat(path)
                if err != nil {
                    /* Gotta be speed, skip on error */
                    continue
                }
                timeModified := stat.ModTime().UnixNano()

                file := regularCache.CacheMap[path]
                file.Lock()
                if file.IsFresh() && file.LastRefresh() < timeModified {
                    file.SetUnfresh()
                }
                file.Unlock()
            }
            regularCache.Mutex.RUnlock()

            /* Check gophermap file cache is fresh */
            gophermapCache.Mutex.RLock()
            for path := range gophermapCache.CacheMap {
                /* Check if file on disk has changed. */
                stat, err := os.Stat(path)
                if err != nil {
                    /* Gotta be speed, skip on error */
                    continue
                }
                timeModified := stat.ModTime().UnixNano()

                file := gophermapCache.CacheMap[path]
                file.Lock()
                if file.IsFresh() && file.LastRefresh() < timeModified {
                    file.SetUnfresh()
                }
                file.Unlock()
            }
            gophermapCache.Mutex.RUnlock()

            /* Sleep so we don't take up all the precious CPU time :) */
            time.Sleep(FileMonitorSleepTime)
        }

        /* We shouldn't have reached here */
        logSystemFatal("FileCache monitor crashed!\n")
    }()
}

type File interface {
    /* File contents */
    Contents()     []byte
    LoadContents() *GophorError

    /* Cache state */
    IsFresh()      bool
    SetUnfresh()
    LastRefresh()  int64

    /* Mutex */
    Lock()
    Unlock()
    RLock()
    RUnlock()
}

type RegularFileCache struct {
    CacheMap map[string]*RegularFile
    Mutex    sync.RWMutex
}

func (fc *RegularFileCache) Init() {
    fc.CacheMap = make(map[string]*RegularFile)
    fc.Mutex = sync.RWMutex{}
}

func (fc *RegularFileCache) Fetch(path string) ([]byte, *GophorError) {
    /* Get read lock, try get file and defer read unlock */
    fc.Mutex.RLock()
    file, ok := fc.CacheMap[path]

    if ok {
        /* File in cache, check is fresh */
        fc.CacheMap[path].Lock()
        if !file.IsFresh() {
            /* File not fresh, get write lock and update cache */
            gophorErr := fc.CacheMap[path].LoadContents()
            if gophorErr != nil {
                /* Error loading contents, unlock write, return error */
                fc.CacheMap[path].Unlock()
                fc.Mutex.RUnlock()
                return nil, gophorErr
            }
        }
        fc.CacheMap[path].Unlock()
    } else {
        /* File not in cache, get write lock then we need to load the file */
        fc.Mutex.RUnlock()
        fc.Mutex.Lock()

        file = NewRegularFile(path)

        /* We don't need to lock the file here before loading contents as we have a map lock */
        gophorErr := file.LoadContents()
        if gophorErr != nil {
            /* Error loading contents, unlock write, return error */
            fc.Mutex.Unlock()
            return nil, gophorErr
        }

        /* Place the file in the cache then unlock :) */
        fc.CacheMap[path] = file
        fc.Mutex.Unlock()
        fc.Mutex.RLock()
    }

    file.RLock()
    b := file.Contents()
    file.RUnlock()
    fc.Mutex.RUnlock()

    return b, nil
}

type GophermapFileCache struct {
    CacheMap map[string]*GophermapFile
    Mutex    sync.RWMutex
}

func (fc *GophermapFileCache) Init() {
    fc.CacheMap = make(map[string]*GophermapFile)
    fc.Mutex = sync.RWMutex{}
}

func (fc *GophermapFileCache) Fetch(path string) ([]byte, *GophorError) {
    /* Get read lock, try get file and defer read unlock */
    fc.Mutex.RLock()
    file, ok := fc.CacheMap[path]

    if ok {
        /* File in cache, check is fresh */
        fc.CacheMap[path].Lock()
        if !file.IsFresh() {
            /* File not fresh, get write lock and update cache */
            gophorErr := fc.CacheMap[path].LoadContents()
            if gophorErr != nil {
                /* Error loading contents, unlock write, return error */
                fc.CacheMap[path].Unlock()
                fc.Mutex.RUnlock()
                return nil, gophorErr
            }
        }
        fc.CacheMap[path].Unlock()
    } else {
        /* File not in cache, get write lock then we need to load the file */
        fc.Mutex.RUnlock()
        fc.Mutex.Lock()

        file = NewGophermapFile(path)

        /* We don't need to lock the file here before loading contents as we have a map lock */
        gophorErr := file.LoadContents()
        if gophorErr != nil {
            /* Error loading contents, unlock write, return error */
            fc.Mutex.Unlock()
            return nil, gophorErr
        }

        /* Place the file in the cache then unlock :) */
        fc.CacheMap[path] = file
        fc.Mutex.Unlock()
        fc.Mutex.RLock()
    }

    file.RLock()
    b := file.Contents()
    file.RUnlock()
    fc.Mutex.RUnlock()

    return b, nil
}
