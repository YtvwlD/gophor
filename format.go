package main

import (
    "strings"
)

const (
    /* Just naming some constants */
    DOSLineEnd  = "\r\n"
    UnixLineEnd = "\n"
    End         = "."
    Tab         = "\t"
    LastLine    = End+DOSLineEnd

    /* Gopher line formatting */
    MaxUserNameLen          = 70  /* RFC 1436 standard, though we use user-supplied page-width */
    MaxSelectorLen          = 255 /* RFC 1436 standard */
    SelectorErrorStr        = "/max_selector_length_reached"
    GophermapRenderErrorStr = ""
    GophermapReadErrorStr   = "Error reading subgophermap: "
    GophermapExecErrorStr   = "Error executing gophermap: "

    /* Default null values */
    NullSelector = "-"
    NullHost     = "null.host"
    NullPort     = "0"

    /* Replacement strings */
    ReplaceStrHostname = "$hostname"
    ReplaceStrPort     = "$port"
)

/*
 * Item type characters:
 * Collected from RFC 1436 standard, Wikipedia, Go-gopher project
 * and Gophernicus project. Those with ALL-CAPS descriptions in
 * [square brackets] defined and used by Gophernicus, a popular
 * Gopher server.
 */
type ItemType byte
const (
    /* RFC 1436 Standard */
    TypeFile          = ItemType('0') /* Regular file (text) */
    TypeDirectory     = ItemType('1') /* Directory (menu) */
    TypeDatabase      = ItemType('2') /* CCSO flat db; other db */
    TypeError         = ItemType('3') /* Error message */
    TypeMacBinHex     = ItemType('4') /* Macintosh BinHex file */
    TypeBinArchive    = ItemType('5') /* Binary archive (zip, rar, 7zip, tar, gzip, etc), CLIENT MUST READ UNTIL TCP CLOSE */
    TypeUUEncoded     = ItemType('6') /* UUEncoded archive */
    TypeSearch        = ItemType('7') /* Query search engine or CGI script */
    TypeTelnet        = ItemType('8') /* Telnet to: VT100 series server */
    TypeBin           = ItemType('9') /* Binary file (see also, 5), CLIENT MUST READ UNTIL TCP CLOSE */
    TypeTn3270        = ItemType('T') /* Telnet to: tn3270 series server */
    TypeGif           = ItemType('g') /* GIF format image file (just use I) */
    TypeImage         = ItemType('I') /* Any format image file */
    TypeRedundant     = ItemType('+') /* Redundant (indicates mirror of previous item) */

    /* GopherII Standard */
    TypeCalendar      = ItemType('c') /* Calendar file */
    TypeDoc           = ItemType('d') /* Word-processing document; PDF document */
    TypeHtml          = ItemType('h') /* HTML document */
    TypeInfo          = ItemType('i') /* Informational text (not selectable) */
    TypeMarkup        = ItemType('p') /* Page layout or markup document (plain text w/ ASCII tags) */
    TypeMail          = ItemType('M') /* Email repository (MBOX) */
    TypeAudio         = ItemType('s') /* Audio recordings */
    TypeXml           = ItemType('x') /* eXtensible Markup Language document */
    TypeVideo         = ItemType(';') /* Video files */

    /* Commonly Used */
    TypeTitle         = ItemType('!') /* [SERVER ONLY] Menu title (set title ONCE per gophermap) */
    TypeComment       = ItemType('#') /* [SERVER ONLY] Comment, rest of line is ignored */
    TypeHiddenFile    = ItemType('-') /* [SERVER ONLY] Hide file/directory from directory listing */
    TypeEnd           = ItemType('.') /* [SERVER ONLY] Last line -- stop processing gophermap default */
    TypeSubGophermap  = ItemType('=') /* [SERVER ONLY] Include subgophermap / regular file here. */
    TypeEndBeginList  = ItemType('*') /* [SERVER ONLY] Last line + directory listing -- stop processing gophermap and end on directory listing */

    /* Default type */
    TypeDefault       = TypeBin

    /* Gophor specific types */
    TypeInfoNotStated = ItemType('I') /* [INTERNAL USE] */
    TypeUnknown       = ItemType('?') /* [INTERNAL USE] */
)


var FileExtMap = map[string]ItemType{
    ".out":          TypeBin,
    ".a":            TypeBin,
    ".o":            TypeBin,
    ".ko":           TypeBin, /* ... Though tbh, kernel extensions?!!! */
    ".msi":          TypeBin,
    ".exe":          TypeBin,

    ".lz":           TypeBinArchive,
    ".gz":           TypeBinArchive,
    ".bz2":          TypeBinArchive,
    ".7z":           TypeBinArchive,
    ".zip":          TypeBinArchive,

    ".gitignore":    TypeFile,
    ".txt":          TypeFile,
    ".json":         TypeFile,
    ".yaml":         TypeFile,
    ".ocaml":        TypeFile,
    ".s":            TypeFile,
    ".c":            TypeFile,
    ".py":           TypeFile,
    ".h":            TypeFile,
    ".go":           TypeFile,
    ".fs":           TypeFile,
    ".odin":         TypeFile,
    ".nanorc":       TypeFile,
    ".bashrc":       TypeFile,
    ".mkshrc":       TypeFile,
    ".vimrc":        TypeFile,
    ".vim":          TypeFile,
    ".viminfo":      TypeFile,
    ".sh":           TypeFile,
    ".conf":         TypeFile,
    ".xinitrc":      TypeFile,
    ".jstarrc":      TypeFile,
    ".joerc":        TypeFile,
    ".jpicorc":      TypeFile,
    ".profile":      TypeFile,
    ".bash_profile": TypeFile,
    ".bash_logout":  TypeFile,
    ".log":          TypeFile,
    ".ovpn":         TypeFile,

    ".md":           TypeMarkup,

    ".xml":          TypeXml,

    ".doc":          TypeDoc,
    ".docx":         TypeDoc,
    ".pdf":          TypeDoc,

    ".jpg":          TypeImage,
    ".jpeg":         TypeImage,
    ".png":          TypeImage,
    ".gif":          TypeImage,

    ".html":         TypeHtml,
    ".htm":          TypeHtml,

    ".ogg":          TypeAudio,
    ".mp3":          TypeAudio,
    ".wav":          TypeAudio,
    ".mod":          TypeAudio,
    ".it":           TypeAudio,
    ".xm":           TypeAudio,
    ".mid":          TypeAudio,
    ".vgm":          TypeAudio,
    ".opus":         TypeAudio,
    ".m4a":          TypeAudio,
    ".aac":          TypeAudio,

    ".mp4":          TypeVideo,
    ".mkv":          TypeVideo,
    ".webm":         TypeVideo,
}

func buildError(selector string) []byte {
    ret := string(TypeError)
    ret += selector + DOSLineEnd
    ret += LastLine
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

/* Build a line separator of supplied width */
func buildLineSeparator(count int) string {
    ret := ""
    for i := 0; i < count; i += 1 {
        ret += "_"
    }
    return ret
}

/* Formats an info-text footer from string. Add last line as we use the footer to contain last line (regardless if empty) */
func formatGophermapFooter(text string, useSeparator bool) []byte {
    ret := make([]byte, 0)
    if text != "" {
        ret = append(ret, buildInfoLine("")...)
        if useSeparator {
            ret = append(ret, buildInfoLine(buildLineSeparator(Config.PageWidth))...)
        }
        for _, line := range strings.Split(text, "\n") {
            ret = append(ret, buildInfoLine(line)...)
        }
    }
    ret = append(ret, []byte(LastLine)...)
    return ret
}

/* Replace standard replacement strings */
func replaceStrings(str string, connHost *ConnHost) []byte {
    str = strings.Replace(str, ReplaceStrHostname, connHost.Name, -1)
    str = strings.Replace(str, ReplaceStrPort, connHost.Port, -1)
    return []byte(str)
}

/* Replace new-line characters */
func replaceNewLines(str string) string {
    return strings.Replace(str, "\n", "", -1)
}
