package main

import (
    "strings"
)

var FileExtMap = map[string]ItemType{
    ".out":       TypeBin,
    ".a":         TypeBin,
    ".o":         TypeBin,
    ".ko":        TypeBin, /* ... Though tbh, kernel extensions?!!! */
    ".msi":       TypeBin,
    ".exe":       TypeBin,

    ".lz":        TypeBinArchive,
    ".gz":        TypeBinArchive,
    ".bz2":       TypeBinArchive,
    ".7z":        TypeBinArchive,
    ".zip":       TypeBinArchive,

    ".gitignore": TypeFile,
    ".txt":       TypeFile,
    ".json":      TypeFile,
    ".yaml":      TypeFile,
    ".ocaml":     TypeFile,
    ".s":         TypeFile,
    ".c":         TypeFile,
    ".py":        TypeFile,
    ".h":         TypeFile,
    ".go":        TypeFile,
    ".fs":        TypeFile,
    ".odin":      TypeFile,
    ".vim":       TypeFile,
    ".nanorc":    TypeFile,

    ".md":        TypeMarkup,

    ".xml":       TypeXml,

    ".doc":       TypeDoc,
    ".docx":      TypeDoc,
    ".pdf":       TypeDoc,

    ".jpg":       TypeImage,
    ".jpeg":      TypeImage,
    ".png":       TypeImage,
    ".gif":       TypeImage,

    ".html":      TypeHtml,
    ".htm":       TypeHtml,

    ".ogg":       TypeAudio,
    ".mp3":       TypeAudio,
    ".wav":       TypeAudio,
    ".mod":       TypeAudio,
    ".it":        TypeAudio,
    ".xm":        TypeAudio,
    ".mid":       TypeAudio,
    ".vgm":       TypeAudio,
    ".opus":      TypeAudio,
    ".m4a":       TypeAudio,
    ".aac":       TypeAudio,

    ".mp4":       TypeVideo,
    ".mkv":       TypeVideo,
    ".webm":      TypeVideo,
}

func buildError(selector string) []byte {
    ret := string(TypeError)
    ret += selector + DOSLineEnd
    return []byte(ret)
}

/* Build gopher compliant line with supplied information */
func buildLine(t ItemType, name, selector, host string, port string) []byte {
    ret := string(t)

    /* Add name, truncate name if too long */    
    if len(name) > Config.PageWidth {
        ret += name[:Config.PageWidth-5]+"...\t"
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

    /* Add host + port */
    ret += host+"\t"+port+DOSLineEnd

    return []byte(ret)
}

/* Build gopher compliant info line */
func buildInfoLine(content string) []byte {
    return buildLine(TypeInfo, content, NullSelector, NullHost, NullPort)
}

/* Get item type for named file on disk */
func getItemType(name string) ItemType {
    /* Split, name MUST be lower */
    split := strings.Split(strings.ToLower(name), ".")

    /* First we look at how many '.' in name string */
    splitLen := len(split)
    switch splitLen {
        case 0:
            /* Always return TypeDefault. We can never tell */
            return TypeDefault

        default:
            /* Get index of str after last ".", look in FileExtMap */
            fileType, ok := FileExtMap["."+split[splitLen-1]]
            if ok {
                return fileType
            } else {
                return TypeDefault
            }
    }
}

/* Parse line type from contents */
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
