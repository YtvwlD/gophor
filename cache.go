package main

import (
    "sync"
)

type File interface {
    Contents()     []byte
    LoadContents() *GophorError
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
    /* Try get file */
    file, ok := fc.GetFile(path)
    
    if !ok {
        /* File not in cache, we need to load the file */
        var gophorErr *GophorError
        file, gophorErr = fc.Load(path)
        if gophorErr != nil {
            return nil, gophorErr
        }
    }

    return file.Contents(), nil
}

func (fc *RegularFileCache) GetFile(path string) (*RegularFile, bool) {
    /* Get read lock, try get file from the cache */
    fc.Mutex.RLock()
    file, ok := fc.CacheMap[path]
    fc.Mutex.RUnlock()
    return file, ok
}

func (fc *RegularFileCache) Load(path string) (*RegularFile, *GophorError) {
    /* Create new file object for path, load contents */
    file := new(RegularFile)
    file.path = path

    gophorErr := file.LoadContents()
    if gophorErr != nil {
        return nil, gophorErr
    }

    /* Get lock, add file object to cache */
    fc.Mutex.Lock()
    fc.CacheMap[path] = file
    fc.Mutex.Unlock()

    return file, nil
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
    /* Try get file */
    file, ok := fc.GetFile(path)
    
    if !ok {
        /* File not in cache, we need to load the file */
        var gophorErr *GophorError
        file, gophorErr = fc.Load(path)
        if gophorErr != nil {
            return nil, gophorErr
        }
    }

    return file.Contents(), nil
}

func (fc *GophermapFileCache) GetFile(path string) (*GophermapFile, bool) {
    /* Get read lock, try get file from the cache */
    fc.Mutex.RLock()
    file, ok := fc.CacheMap[path]
    fc.Mutex.RUnlock()
    return file, ok
}

func (fc *GophermapFileCache) Load(path string) (*GophermapFile, *GophorError) {
    /* Create new file object for path, load contents */
    file := new(GophermapFile)
    file.path = path

    gophorErr := file.LoadContents()
    if gophorErr != nil {
        return nil, gophorErr
    }

    /* Get lock, add file object to cache. */
    fc.Mutex.Lock()
    fc.CacheMap[path] = file
    fc.Mutex.Unlock()

    return file, nil
}
