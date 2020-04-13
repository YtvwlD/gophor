package main

import (
    "os"
    "path"
    "strings"
)

func listDir(dirPath string, hidden map[string]bool) ([]byte, *GophorError) {
    /* Open directory file descriptor */
    fd, err := os.Open(dirPath)
    if err != nil {
        logSystemError("failed to open %s: %s\n", dirPath, err.Error())
        return nil, &GophorError{ FileOpenErr, err }
    }

    /* Open directory stream for reading */
    files, err := fd.Readdir(-1)
    if err != nil {
        logSystemError("failed to enumerate dir %s: %s\n", dirPath, err.Error())
        return nil, &GophorError{ DirListErr, err }
    }

    var entity *DirEntity
    dirContents := make([]byte, 0)

    /* Walk through directory */
    for _, file := range files {
        /* Skip dotfiles + gophermap file + requested hidden */
        if file.Name()[0] == '.' || strings.HasSuffix(file.Name(), GophermapFileStr) {
            continue
        } else if _, ok := hidden[file.Name()]; ok {
            continue
        }

        /* Handle file, directory or ignore others */
        switch {
            case file.Mode() & os.ModeDir != 0:
                /* Directory -- create directory listing */
                itemPath := path.Join(fd.Name(), file.Name())
                entity = newDirEntity(TypeDirectory, file.Name(), "/"+itemPath, *ServerHostname, *ServerPort)
                dirContents = append(dirContents, entity.Bytes()...)

            case file.Mode() & os.ModeType == 0:
                /* Regular file -- find item type and creating listing */
                itemPath := path.Join(fd.Name(), file.Name())
                itemType := getItemType(itemPath)
                entity = newDirEntity(itemType, file.Name(), "/"+itemPath, *ServerHostname, *ServerPort)
                dirContents = append(dirContents, entity.Bytes()...)

            default:
                /* Ignore */
        }
    }

    return dirContents, nil
}
