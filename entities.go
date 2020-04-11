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

    MaxUserNameLen = 70  /* RFC 1436 standard */
    MaxSelectorLen = 255 /* RFC 1436 standard */

    UserNameErr = "! Err: Max UserName len reached"
    SelectorErr = "err_max_selector_len_reached"
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
    TypeTitle         = ItemType('!') /* [TITLE] */

    /* Default type */
    TypeDefault       = TypeFile
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
    if len(name) > MaxUserNameLen {
        name = UserNameErr
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
