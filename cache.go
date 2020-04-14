package main

import (
    "os"
    "sync"
    "time"
    "container/list"
)

const (
    FileMonitorSleepTimeMs = 30
    FileMonitorSleepTime = time.Duration(FileMonitorSleepTimeMs) * time.Second
)

func StartFileMonitor(regularCache *RegularFileCache, gophermapCache *GophermapFileCache) {
    go func() {
        for {
            /* Check regular file cache is fresh */
            regularCache.CacheMutex.RLock()
            for path := range regularCache.CacheMap {
                /* Check if file on disk has changed. */
                stat, err := os.Stat(path)
                if err != nil {
                    /* Gotta be speed, skip on error */
                    continue
                }
                timeModified := stat.ModTime().UnixNano()

                fileElement := regularCache.CacheMap[path]
                fileElement.File.Lock()
                if fileElement.File.IsFresh() && fileElement.File.LastRefresh() < timeModified {
                    fileElement.File.SetUnfresh()
                }
                fileElement.File.Unlock()
            }
            regularCache.CacheMutex.RUnlock()

            /* Check gophermap file cache is fresh */
            gophermapCache.CacheMutex.RLock()
            for path := range gophermapCache.CacheMap {
                /* Check if file on disk has changed. */
                stat, err := os.Stat(path)
                if err != nil {
                    /* Gotta be speed, skip on error */
                    continue
                }
                timeModified := stat.ModTime().UnixNano()

                fileElement := gophermapCache.CacheMap[path]
                fileElement.File.Lock()
                if fileElement.File.IsFresh() && fileElement.File.LastRefresh() < timeModified {
                    fileElement.File.SetUnfresh()
                }
                fileElement.File.Unlock()
            }
            gophermapCache.CacheMutex.RUnlock()

            /* Sleep so we don't take up all the precious CPU time :) */
            time.Sleep(FileMonitorSleepTime)
        }

        /* We shouldn't have reached here */
        logSystemFatal("FileCache monitor escaped run loop!\n")
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

type RegularFileElement struct {
    File    *RegularFile
    Element *list.Element
}

type RegularFileCache struct {
    CacheMap   map[string]*RegularFileElement
    CacheMutex sync.RWMutex
    FileList   *list.List
    ListMutex  sync.Mutex
    Size       int
}

func (fc *RegularFileCache) Init(size int) {
    fc.CacheMap = make(map[string]*RegularFileElement)
    fc.CacheMutex = sync.RWMutex{}
    fc.FileList = list.New()
    fc.FileList.Init()
    fc.ListMutex = sync.Mutex{}
    fc.Size = size
}

func (fc *RegularFileCache) Fetch(path string) ([]byte, *GophorError) {
    /* Get read lock, try get file and defer read unlock */
    fc.CacheMutex.RLock()
    fileElement, ok := fc.CacheMap[path]

    if ok {
        /* File in cache -- before doing anything get read lock */
        fileElement.File.RLock()

        /* Now check is fresh */
        if !fileElement.File.IsFresh() {
            /* File not fresh! Swap read for write-lock */
            fileElement.File.RUnlock()
            fileElement.File.Lock()

            gophorErr := fileElement.File.LoadContents()
            if gophorErr != nil {
                /* Error loading contents, unlock all mutex then return error */
                fileElement.File.Unlock()
                fc.CacheMutex.RUnlock()
                return nil, gophorErr
            }

            /* Updated! Swap back to file read lock for upcoming content read */
            fileElement.File.Unlock()
            fileElement.File.RLock()
        }
    } else {
        /* File not in cache -- Swap cache map read lock for write lock */
        fc.CacheMutex.RUnlock()
        fc.CacheMutex.Lock()

        /* New file init function */
        file := NewRegularFile(path)

        /* NOTE: file isn't in cache yet so no need to lock file mutex */
        gophorErr := file.LoadContents()
        if gophorErr != nil {
            /* Error loading contents, unlock all mutex then return error */
            fc.CacheMutex.Unlock()
            return nil, gophorErr
        }

        /* Place path in FileList to get back element */
        element := fc.FileList.PushFront(path)

        /* Create fileElement and place in map */
        fileElement = &RegularFileElement{ file, element }
        fc.CacheMap[path] = fileElement

        /* If we're at capacity, remove last item in list from map+list */
        if fc.FileList.Len() == fc.Size {
            removeElement := fc.FileList.Back()

            /* Have to perform type assertion, if error we'll exit */
            removePath, ok := removeElement.Value.(string)
            if !ok {
                logSystemFatal("Non-string found in cache list!\n")
            }

            /* Get lock to ensure no-one else using */
            fc.CacheMap[removePath].File.Lock()
            fc.CacheMap[removePath].File.Unlock()

            /* Now delete. We don't need ListMutex lock as we have cache map write lock */
            delete(fc.CacheMap, removePath)
            fc.FileList.Remove(removeElement)
        }

        /* Swap cache lock back to read */
        fc.CacheMutex.Unlock()
        fc.CacheMutex.RLock()

        /* Get file read lock for upcoming content read */
        file.RLock()
    }

    /* Read file contents into new variable for return, then unlock all */
    b := fileElement.File.Contents()
    fileElement.File.RUnlock()

    /* First get list lock, now update placement in list */
    fc.ListMutex.Lock()
    fc.FileList.MoveToFront(fileElement.Element)
    fc.ListMutex.Unlock()

    fc.CacheMutex.RUnlock()

    return b, nil
}

type GophermapFileElement struct {
    File    *GophermapFile
    Element *list.Element
}

type GophermapFileCache struct {
    CacheMap   map[string]*GophermapFileElement
    CacheMutex sync.RWMutex
    FileList   *list.List
    ListMutex  sync.Mutex
    Size       int
}

func (fc *GophermapFileCache) Init(size int) {
    fc.CacheMap = make(map[string]*GophermapFileElement)
    fc.CacheMutex = sync.RWMutex{}
    fc.FileList = list.New()
    fc.FileList.Init()
    fc.ListMutex = sync.Mutex{}
    fc.Size = size
}

func (fc *GophermapFileCache) Fetch(path string) ([]byte, *GophorError) {
    /* Get read lock, try get file and defer read unlock */
    fc.CacheMutex.RLock()
    fileElement, ok := fc.CacheMap[path]

    if ok {
        /* File in cache -- before doing anything get read lock */
        fileElement.File.RLock()

        /* Now check is fresh */
        if !fileElement.File.IsFresh() {
            /* File not fresh! Swap read for write-lock */
            fileElement.File.RUnlock()
            fileElement.File.Lock()

            gophorErr := fileElement.File.LoadContents()
            if gophorErr != nil {
                /* Error loading contents, unlock all mutex then return error */
                fileElement.File.Unlock()
                fc.CacheMutex.RUnlock()
                return nil, gophorErr
            }

            /* Updated! Swap back to file read lock for upcoming content read */
            fileElement.File.Unlock()
            fileElement.File.RLock()
        }
    } else {
        /* File not in cache -- Swap cache map read lock for write lock */
        fc.CacheMutex.RUnlock()
        fc.CacheMutex.Lock()

        /* New file init function */
        file := NewGophermapFile(path)

        /* NOTE: file isn't in cache yet so no need to lock file mutex */
        gophorErr := file.LoadContents()
        if gophorErr != nil {
            /* Error loading contents, unlock all mutex then return error */
            fc.CacheMutex.Unlock()
            return nil, gophorErr
        }

        /* Place path in FileList to get back element */
        element := fc.FileList.PushFront(path)

        /* Create fileElement and place in map */
        fileElement = &GophermapFileElement{ file, element }
        fc.CacheMap[path] = fileElement

        /* If we're at capacity, remove last item in list from map+list */
        if fc.FileList.Len() == fc.Size {
            removeElement := fc.FileList.Back()

            /* Have to perform type assertion, if error we'll exit */
            removePath, ok := removeElement.Value.(string)
            if !ok {
                logSystemFatal("Non-string found in cache list!\n")
            }

            /* Get lock to ensure no-one else using */
            fc.CacheMap[removePath].File.Lock()
            fc.CacheMap[removePath].File.Unlock()

            /* Now delete. We don't need ListMutex lock as we have cache map write lock */
            delete(fc.CacheMap, removePath)
            fc.FileList.Remove(removeElement)
        }

        /* Swap cache lock back to read */
        fc.CacheMutex.Unlock()
        fc.CacheMutex.RLock()

        /* Get file read lock for upcoming content read */
        file.RLock()
    }

    /* Read file contents into new variable for return, then unlock all */
    b := fileElement.File.Contents()
    fileElement.File.RUnlock()

    /* First get list lock, now update placement in list */
    fc.ListMutex.Lock()
    fc.FileList.MoveToFront(fileElement.Element)
    fc.ListMutex.Unlock()

    fc.CacheMutex.RUnlock()

    return b, nil
}
