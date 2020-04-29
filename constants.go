package main

const (
    /* Gophor */
    GophorVersion = "0.5-alpha-PR3"

    /* Parsing */
    DOSLineEnd = "\r\n"
    UnixLineEnd = "\n"

    End = "."
    Tab = "\t"
    LastLine = End+DOSLineEnd

    MaxUserNameLen = 70  /* RFC 1436 standard */
    MaxSelectorLen = 255 /* RFC 1436 standard */

    NullSelector = "-"
    NullHost = "null.host"
    NullPort = "0"

    SelectorErrorStr = "selector_length_error"
    GophermapRenderErrorStr = ""

    ReplaceStrHostname = "$hostname"
    ReplaceStrPort = "$port"

    /* Filesystem */
    GophermapFileStr = "gophermap"
    CapsTxtStr = "caps.txt"
    RobotsTxtStr = "robots.txt"

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
    TypeEndBeginList  = ItemType('*') /* [SERVER ONLY] Last line + directory listing -- stop processing gophermap and end on a directory listing */

    /* Planned To Be Supported */
    TypeExec          = ItemType('$') /* [SERVER ONLY] Execute shell command and print stdout here */

    /* Default type */
    TypeDefault       = TypeBin

    /* Gophor specific types */
    TypeInfoNotStated = ItemType('z') /* [INTERNAL USE] */
    TypeUnknown       = ItemType('?') /* [INTERNAL USE] */
)
