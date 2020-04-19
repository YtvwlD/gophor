package main

const (
    /* Parsing */
    DOSLineEnd = "\r\n"
    UnixLineEnd = "\n"

    End = "."
    Tab = "\t"
    LastLine = End+DOSLineEnd

    MaxUserNameLen = 70  /* RFC 1436 standard */
    MaxSelectorLen = 255 /* RFC 1436 standard */

    NullHost = "null.host"
    NullPort = "1"

    SelectorErrorStr = "selector_length_error"
    GophermapRenderErrorStr = ""

    ReplaceStrHostname = "$hostname"

    /* Filesystem */
    GophermapFileStr = "gophermap"

    /* Misc */
    BytesInMegaByte = 1048576.0
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
    TypeDefault       = TypeBin

    /* Gophor specific types */
    TypeExec          = ItemType('$') /* Execute command and insert stdout here */
    TypeInfoNotStated = ItemType('z') /* INTERNAL USE. We use this in a switch case, a line never starts with this */
    TypeBanned        = ItemType('@') /* INTERNAL USE. We use this in a switch case, a line never starts with this */
    TypeUnknown       = ItemType('?') /* INTERNAL USE. We use this in a switch case, a line never starts with this */
)
