package main

import (
    "strconv"
    "strings"
    "path/filepath"
)

const (
    CrLf = "\r\n"
    End  = "."
    LastLine = End+CrLf
    Tab  = byte('\t')

/*    MaxUserNameLen = 70  RFC 1436 standard */
    MaxSelectorLen = 255 /* RFC 1436 standard */

    UserNameErr = "! Err: Max UserName len reached"
    SelectorErr = "err_max_selector_len_reached"

    NullHost = "null.host"
    NullPort = "1"
)

type ItemType byte

/*
 * Item type characters:
 * Collected from RFC 1436 standard, Wikipedia, Go-gopher project
 * and Gophernicus project. Those with ALL-CAPS descriptions in
 * [square brackets] defined and used by Gophernicus, a popular
 * Gopher server.
 */
const (
    /* RFC 1436 Standard */
    TypeFile          = ItemType('0') /* Regular file [TEXT] */
    TypeDirectory     = ItemType('1') /* Directory [MENU] */
    TypePhonebook     = ItemType('2') /* CSO phone-book server */
    TypeError         = ItemType('3') /* Error [ERROR] */
    TypeMacBinHex     = ItemType('4') /* Binhexed macintosh file */
    TypeDosBinArchive = ItemType('5') /* DOS bin archive, CLIENT MUST READ UNTIL TCP CLOSE [GZIP] */
    TypeUnixFile      = ItemType('6') /* Unix uuencoded file */
    TypeIndexSearch   = ItemType('7') /* Index-search server [QUERY] */
    TypeTelnet        = ItemType('8') /* Text-based telnet session */
    TypeBin           = ItemType('9') /* Binary file, CLIENT MUST READ UNTIL TCP CLOSE [BINARY] */
    TypeTn3270        = ItemType('T') /* Text-based tn3270 session */
    TypeGif           = ItemType('g') /* Gif format graphics file [GIF] */
    TypeImage         = ItemType('I') /* Some kind of image file (client decides how to display) [IMAGE] */

    TypeRedundant     = ItemType('+') /* Redundant server */

    TypeEnd           = ItemType('.') /* Indicates LastLine if only this + CrLf */

    /* Non-standard - as used by https://github.com/prologic/go-gopher
     * (also seen on Wikipedia: https://en.wikipedia.org/wiki/Gopher_%28protocol%29#Item_types)
     */
    TypeInfo          = ItemType('i') /* Informational message [INFO] */
    TypeHtml          = ItemType('h') /* HTML document [HTML] */
    TypeAudio         = ItemType('s') /* Audio file */
    TypePng           = ItemType('p') /* PNG image */
    TypeDoc           = ItemType('d') /* Document [DOC] */

    /* Non-standard - as used by Gopernicus https://github.com/gophernicus/gophernicus */
    TypeMime          = ItemType('M') /* [MIME] */
    TypeVideo         = ItemType(';') /* [VIDEO] */
    TypeCalendar      = ItemType('c') /* [CALENDAR] */
    TypeTitle         = ItemType('!') /* [TITLE] */
    TypeComment       = ItemType('#') /* [COMMENT] */
    TypeHiddenFile    = ItemType('-') /* [HIDDEN] Hides file from directory listing */
    TypeSubGophermap  = ItemType('=') /* [EXECUTE] read this file in here */
    TypeEndBeginList  = ItemType('*') /* If only this + CrLf, indicates last line but then followed by directory list */

    /* Default type */
    TypeDefault       = TypeFile

    /* Gophor specific types */
    TypeExec          = ItemType('$') /* Execute command and insert stdout here */
    TypeInfoNotStated = ItemType('z') /* INTERNAL USE. We use this in a switch case, a line never starts with this */
    TypeUnknown       = ItemType('?') /* INTERNAL USE. We use this in a switch case, a line never starts with this */
)

/*
 * Directory Entity data structure for easier handling
 */
type DirEntity struct {
    /* RFC 1436 standard */
    Type     ItemType
    UserName string
    Selector string
    Host     string
    Port     string

    /* Non-standard, proposed Gopher+
     * gopher://gopher.floodgap.com:70/0/gopher/tech/gopherplus.txt
     */
    Extras   string
}

func newDirEntity(t ItemType, name, selector, host string, port int) *DirEntity {
    entity := new(DirEntity)
    entity.Type = t

    /* Truncate username if we hit MaxUserNameLen */
    if len(name) > *PageWidth {
        name = name[:*PageWidth-4] + "..."
    }
    entity.UserName = name

    /* Truncate selector if we hit MaxSelectorLen */
    if len(selector) > MaxSelectorLen {
        selector = SelectorErr
    }
    entity.Selector = selector

    entity.Host = host
    entity.Port = strconv.Itoa(port)
    return entity
}

func (entity *DirEntity) Bytes() []byte {
    b := []byte{}
    b = append(b, byte(entity.Type))
    b = append(b, []byte(entity.UserName)...)
    b = append(b, Tab)
    b = append(b, []byte(entity.Selector)...)
    b = append(b, Tab)
    b = append(b, []byte(entity.Host)...)
    b = append(b, Tab)
    b = append(b, []byte(entity.Port)...)
    if entity.Extras != "" {
        b = append(b, Tab)
        b = append(b, []byte(entity.Extras)...)
    }
    b = append(b, []byte(CrLf)...)
    return b
}

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

func createInfoLine(content string) []byte {
    return []byte(string(TypeInfo)+content+string(Tab)+NullHost+string(Tab)+NullPort+CrLf)
}
