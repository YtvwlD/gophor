package main

import (
    "strconv"
    "strings"
)

var SingleFileExtMap = map[string]ItemType{
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

var DoubleFileExtMap = map[string]ItemType{
    ".tar.gz": TypeBin,
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

    /* Add host + port */
    ret += host+"\t"+strconv.Itoa(port)+DOSLineEnd

    return []byte(ret)
}

func buildInfoLine(content string) []byte {
    return buildLine(TypeInfo, content, NullSelector, NullHost, NullPort)
}

/* getItemType(name string) ItemType:
 * Here we use an empty function pointer, and set the correct
 * function to be used during the restricted files regex parsing.
 * This negates need to check if RestrictedFilesRegex is nil every
 * single call.
 */
var getItemType func(name string) ItemType
func _getItemType(name string) ItemType {
    return _checkItemType(strings.ToLower(name))
}
func _getItemTypeMatchRestricted(name string) ItemType {
    nameLower := strings.ToLower(name)

    /* If regex compiled, check if matches user restricted files */
    if RestrictedFilesRegex != nil &&
       RestrictedFilesRegex.MatchString(nameLower) {
        return TypeBanned
    }

    return _checkItemType(nameLower)
}
func _checkItemType(nameLower string) ItemType {
    /* Name MUST be lower before passed to this function */

    /* First we look at how many '.' in name string */
    switch strings.Count(nameLower, ".") {
        case 0:
            /* Always return TypeDefault. We can never tell */
            return TypeDefault

        case 1:
            /* Get index of ".", try look in SingleFileExtMap */
            i := strings.IndexByte(nameLower, '.')
            fileType, ok := SingleFileExtMap[nameLower[i:]]
            if ok {
                return fileType
            } else {
                return TypeDefault
            }

        default:
            /* Get index of penultimate ".", try look in DoubleFileExtMap */
            i, j := len(nameLower)-1, 0
            for i >= 0 {
                if nameLower[i] == '.' {
                    if j == 1 {
                        break
                    } else {
                        j += 1
                    }
                }
                i -= 1
            }
            fileType, ok := DoubleFileExtMap[nameLower[i:]]
            if ok {
                return fileType
            } else {
                return TypeDefault
            }
    }
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

