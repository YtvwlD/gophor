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
)

type ItemType byte

/*
 * Item type characters
 */
const (
    /* RFC 1436 Standard */
    TypeFile          = ItemType(0)   /* Regular file */
    TypeDirectory     = ItemType(1)   /* Directory */
    TypePhonebook     = ItemType(2)   /* CSO phone-book server */
    TypeError         = ItemType(3)   /* Error */
    TypeMacBinHex     = ItemType(4)   /* Binhexed macintosh file */
    TypeDosBinArchive = ItemType(5)   /* DOS bin archive, CLIENT MUST READ UNTIL TCP CLOSE */
    TypeUnixFile      = ItemType(6)   /* Unix uuencoded file */
    TypeIndexSearch   = ItemType(7)   /* Index-search server */
    TypeTelnet        = ItemType(8)   /* Text-based telnet session */
    TypeBin           = ItemType(9)   /* Binary file, CLIENT MUST READ UNTIL TCP CLOSE */
    TypeTn3270        = ItemType('T') /* Text-based tn3270 session */
    TypeGif           = ItemType('g') /* Gif format graphics file */
    TypeImage         = ItemType('I') /* Some kind of image file (client decides how to display) */

    TypeRedundant     = ItemType('+') /* Redundant server */

    /* Non-standard */
    TypeInfo          = ItemType('i') /* Informational message */
    TypeHtml          = ItemType('h') /* HTML document */
    TypeAudio         = ItemType('s') /* Audio file */
    TypePng           = ItemType('p') /* PNG image */
    TypeDoc           = ItemType('d') /* Document */

    /* Default type */
    TypeDefault       = TypeBin
)

func (i ItemType) String() string {
    switch i {
        case TypeFile:
            return "TXT"
        case TypeDirectory:
            return "DIR"
        case TypePhonebook:
            return "PHO"
        case TypeError:
            return "ERR"
        case TypeMacBinHex:
            return "HEX"
        case TypeDosBinArchive:
            return "ARC"
        case TypeUnixFile:
            return "UUE"
        case TypeIndexSearch:
            return "QRY"
        case TypeTelnet:
            return "TEL"
        case TypeBin:
            return "BIN"
        case TypeTn3270:
            return "TN3"
        case TypeGif:
            return "GIF"
        case TypeImage:
            return "IMG"
        case TypeRedundant:
            return "DUP"
        case TypeInfo:
            return "NFO"
        case TypeHtml:
            return "HTM"
        case TypeAudio:
            return "SND"
        case TypePng:
            return "PNG"
        case TypeDoc:
            return "DOC"
        default:
            return "???"
    }
}

/*
 * Directory Entity data structure for easier handling
 */
type DirEntity struct {
    Type     ItemType
    UserName string
    Selector string
    Host     string
    Port     string
}

func newDirEntity(t ItemType, name, selector, host string, port int) *DirEntity {
    entity := new(DirEntity)
    entity.Type = t
    entity.UserName = name
    if len(selector) > 255 {
        selector = selector[:254]
    }
    entity.Selector = selector
    entity.Host = host
    entity.Port = strconv.Itoa(port)
    return entity
}

func (entity *DirEntity) Bytes() []byte {
    b := make([]byte, 0)
    b = append(b, byte(entity.Type))
    b = append(b, []byte(entity.UserName)...)
    b = append(b, Tab)
    b = append(b, []byte(entity.Selector)...)
    b = append(b, Tab)
    b = append(b, []byte(entity.Host)...)
    b = append(b, Tab)
    b = append(b, []byte(entity.Port)...)
    b = append(b, []byte(CrLf)...)
    return b
}

var FileExtensions = map[string]ItemType{
    ".txt":  TypeFile,
    ".gif":  TypeGif,
    ".jpg":  TypeImage,
    ".jpeg": TypeImage,
    ".png":  TypeImage,
    ".html": TypeHtml,
    ".ogg":  TypeAudio,
    ".mp3":  TypeAudio,
    ".wav":  TypeAudio,
    ".mod":  TypeAudio,
    ".it":   TypeAudio,
    ".xm":   TypeAudio,
    ".mid":  TypeAudio,
    ".vgm":  TypeAudio,
    ".s":    TypeFile,
    ".c":    TypeFile,
    ".py":   TypeFile,
    ".h":    TypeFile,
    ".md":   TypeFile,
    ".go":   TypeFile,
    ".fs":   TypeFile,
}

func getItemType(name string) ItemType {
    extension := strings.ToLower(filepath.Ext(name))
    fileType, ok := FileExtensions[extension]
    if !ok {
        return TypeDefault
    }
    return fileType
}
