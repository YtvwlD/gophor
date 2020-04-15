package main

import (
    "os"
    "sync"
    "time"
    "container/list"
)

const (
    BytesInMegaByte = 1048576.0
)

var (
    FileMonitorSleepTime = time.Duration(*CacheCheckFreq) * time.Second

    /* Global file caches */
    GophermapCache *FileCache
    RegularCache   *FileCache
)

func startFileCaching() {
    /* Create gophermap file cache */
    GophermapCache = new(FileCache)
    GophermapCache.Init(*CacheSize, func(path string) File {
        return NewGophermapFile(path)
    })

    /* Create regular file cache */
    RegularCache = new(FileCache)
    RegularCache.Init(*CacheSize, func(path string) File {
        return NewRegularFile(path)
    })

    /* Start file monitor in separate goroutine */
    go startFileMonitor()
}

func startFileMonitor() {
    go func() {
        for {
            /* Sleep so we don't take up all the precious CPU time :) */
            time.Sleep(FileMonitorSleepTime)

            /* Check regular cache freshness */
            checkCacheFreshness(RegularCache)

            /* Check gophermap cache freshness */
            checkCacheFreshness(GophermapCache)
        }

        /* We shouldn't have reached here */
        logSystemFatal("FileCache monitor escaped run loop!\n")
    }()
}

func checkCacheFreshness(cache *FileCache) {
    /* Before anything, get cache read lock */
    cache.CacheMutex.RLock()

    /* Iterate through paths in cache map to query file last modified times */
    for path := range cache.CacheMap {
        stat, err := os.Stat(path)
        if err != nil {
            /* Gotta be speedy, skip on error */
            logSystemError("failed to stat file in cache: %s\n", path)
            continue
        }
        timeModified := stat.ModTime().UnixNano()

        /* Get file pointer and immediately get write lock */
        file := cache.CacheMap[path].File
        file.Lock()

        /* If the file is marked as fresh, but file on disk newer, mark as unfresh */
        if file.IsFresh() && file.LastRefresh() < timeModified {
            file.SetUnfresh()
        }

        /* Done with file, we can release write lock */
        file.Unlock()
    }

    /* Done! We can release cache read lock */
    cache.CacheMutex.RUnlock()
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

type FileElement struct {
    File    File
    Element *list.Element
}

type FileCache struct {
    CacheMap   map[string]*FileElement
    CacheMutex sync.RWMutex
    FileList   *list.List
    ListMutex  sync.Mutex
    Size       int

    NewFile    func(path string) File
}

func (fc *FileCache) Init(size int, newFileFunc func(path string) File) {
    fc.CacheMap = make(map[string]*FileElement)
    fc.CacheMutex = sync.RWMutex{}
    fc.FileList = list.New()
    fc.FileList.Init()
    fc.ListMutex = sync.Mutex{}
    fc.Size = size
    fc.NewFile = newFileFunc
}

func (fc *FileCache) Fetch(path string) ([]byte, *GophorError) {
    /* Get cache map read lock then check if file in cache map */
    fc.CacheMutex.RLock()
    fileElement, ok := fc.CacheMap[path]

    if ok {
        /* File in cache -- before doing anything get file read lock */
        fileElement.File.RLock()

        /* Check file is marked as fresh */
        if !fileElement.File.IsFresh() {
            /* File not fresh! Swap file read for write-lock */
            fileElement.File.RUnlock()
            fileElement.File.Lock()

            /* Reload file contents from disk */
            gophorErr := fileElement.File.LoadContents()
            if gophorErr != nil {
                /* Error loading contents, unlock all mutex then return error */
                fileElement.File.Unlock()
                fc.CacheMutex.RUnlock()
                return nil, gophorErr
            }

            /* Updated! Swap back file write for read lock */
            fileElement.File.Unlock()
            fileElement.File.RLock()
        }
    } else {
        /* Before we do ANYTHING, we need to check file-size on disk */
        stat, err := os.Stat(path)
        if err != nil {
            return nil, &GophorError{ FileStatErr, err }
        }

        /* Use supplied new file function */
        file := fc.NewFile(path)

        /* NOTE: file isn't in cache yet so no need to lock file write mutex
         * before loading from disk
         */
        gophorErr := file.LoadContents()
        if gophorErr != nil {
            /* Error loading contents, unlock all mutex then return error */
            fc.CacheMutex.Unlock()
            return nil, gophorErr
        }

        /* Compare file size (in MB) to CacheFileSizeMax, if larger just get file
         * contents, unlock all mutex and don't bother caching. 
         */
        if float64(stat.Size()) / BytesInMegaByte > *CacheFileSizeMax {
            logSystem("File too big for cache, skipping: %s\n", path)
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

        /* Place path in FileList to get back element */
        element := fc.FileList.PushFront(path)

        /* Create fileElement and place in map */
        fileElement = &FileElement{ file, element }
        fc.CacheMap[path] = fileElement

        /* If we're at capacity, remove last item in list from cachemap + cache list */
        if fc.FileList.Len() == fc.Size {
            removeElement := fc.FileList.Back()

            /* Have to perform type assertion even if we know value will always be string.
             * If not, we may as well os.Exit(1) out since error is fatal
             */
            removePath, ok := removeElement.Value.(string)
            if !ok {
                logSystemFatal("non-string found in cache list!\n")
            }

            /* Now delete. We don't need ListMutex lock as we have cache map write lock */
            delete(fc.CacheMap, removePath)
            fc.FileList.Remove(removeElement)
        }

        /* Before unlocking cache mutex, lock file read for upcoming call to .Contents() */
        file.RLock()

        /* Swap cache lock back to read */
        fc.CacheMutex.Unlock()
        fc.CacheMutex.RLock()
    }

    /* Get list lock, ready to update placement in list */
    fc.ListMutex.Lock()

    /* Read file contents into new variable for return, then unlock file read lock */
    b := fileElement.File.Contents()
    fileElement.File.RUnlock()

    /* Update placement in list then unlock */
    fc.FileList.MoveToFront(fileElement.Element)
    fc.ListMutex.Unlock()

    /* Finally we can unlock the cache map read lock, we are done :) */
    fc.CacheMutex.RUnlock()

    return b, nil
}
