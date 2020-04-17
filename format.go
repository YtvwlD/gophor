package main

import (
    "strconv"
    "strings"
    "path/filepath"
)

var FileExtensions = map[string]ItemType{
    ".out":   TypeBin,
    ".a":     TypeBin,
    ".o":     TypeBin,
    ".ko":    TypeBin, /* ... Though tbh, kernel extensions?!!! */
    ".msi":   TypeBin,
    ".exe":   TypeBin,

    ".txt":   TypeFile,
    ".md":    TypeFile,
    ".json":  TypeFile,
    ".xml":   TypeFile,
    ".yaml":  TypeFile,
    ".ocaml": TypeFile,
    ".s":     TypeFile,
    ".c":     TypeFile,
    ".py":    TypeFile,
    ".h":     TypeFile,
    ".go":    TypeFile,
    ".fs":    TypeFile,

    ".doc":   TypeDoc,
    ".docx":  TypeDoc,

    ".gif":   TypeGif,

    ".jpg":   TypeImage,
    ".jpeg":  TypeImage,
    ".png":   TypeImage,

    ".html":  TypeHtml,

    ".ogg":   TypeAudio,
    ".mp3":   TypeAudio,
    ".wav":   TypeAudio,
    ".mod":   TypeAudio,
    ".it":    TypeAudio,
    ".xm":    TypeAudio,
    ".mid":   TypeAudio,
    ".vgm":   TypeAudio,

    ".mp4":   TypeVideo,
    ".mkv":   TypeVideo,
}

func buildLine(t ItemType, name, selector, host string, port int) []byte {
    ret := string(t)

    /* Add name, truncate name if too long */    
    if len(name) > *PageWidth {
        ret += name[:*PageWidth-4]+"...\t"
    } else {
        ret += name+"\t"
    }

    /* Add selector. If too long use err, skip if empty */
    selectorLen := len(selector)
    if selectorLen > MaxSelectorLen {
        ret += SelectorErrorStr+"\t"
    } else if selectorLen > 0 {
        ret += selector+"\t"
    }

    /* Add host, set to nullhost if empty */
    if host == "" {
        ret += NullHost+"\t"
    }

    /* Add port, set to nullport if 0 */
    if port == 0 {
        ret += NullPort+CrLf
    } else {
        ret += strconv.Itoa(port)+CrLf
    }

    return []byte(ret)
}

func buildInfoLine(content string) []byte {
    return buildLine(TypeInfo, content, "", "", 0)
}

func getItemType(name string) ItemType {
    extension := strings.ToLower(filepath.Ext(name))
    fileType, ok := FileExtensions[extension]
    if !ok {
        return TypeDefault
    }
    return fileType
}

func parseLineType(line string) ItemType {
    lineLen := len(line)

    if lineLen == 0 {
        return TypeInfoNotStated
    } else if lineLen == 1 {
        /* The only accepted types for a length 1 line */
        switch ItemType(line[0]) {
            case TypeEnd:
                return TypeEnd
            case TypeEndBeginList:
                return TypeEndBeginList
            case TypeComment:
                return TypeComment
            case TypeInfo:
                return TypeInfo
            case TypeTitle:
                return TypeTitle
            default:
                return TypeUnknown
        }
    } else if !strings.Contains(line, string(Tab)) {
        /* The only accepted types for a line with no tabs */
        switch ItemType(line[0]) {
            case TypeComment:
                return TypeComment
            case TypeTitle:
                return TypeTitle
            case TypeInfo:
                return TypeInfo
            case TypeHiddenFile:
                return TypeHiddenFile
            case TypeSubGophermap:
                return TypeSubGophermap
            case TypeExec:
                return TypeExec
            default:
                return TypeInfoNotStated
        }
    }

    return ItemType(line[0])
}

