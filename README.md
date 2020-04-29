# Gophor

A Gopher server written in GoLang as a means of learning about the Gopher
protocol, and more GoLang.

Linux only _for now_. Cross-compiled to way too many architectures.
Build-script now much improved, but still not pretty...

I'm unemployed and work on open-source projects like this and many others for
free. If you would like to help support my work that would be hugely
appreciated 💕 https://liberapay.com/grufwub/

WARNING: the development branch is filled with lava, fear and capitalism.

# Usage

```
gophor [args]
       -root           Change server root directory.
       -port           Change server NON-TLS listening port.
       -hostname       Change server hostname (FQDN, used to craft dir lists).
       -bind-addr      Change server bind-address (used in creating socket).
       -user           Drop to supplied user's UID and GID permissions before execution.
       -system-log     Path to gophor system log file, else use stderr.
       -access-log     Path to gophor access log file, else use stderr.
       -cache-check    Change file-cache freshness check frequency.
       -cache-size     Change max no. files in file-cache.
       -cache-file-max Change maximum allowed size of a cached file.
       -page-width     Change page width used when formatting output.
       -restrict-files New-line separated list of regex statements restricting
                       files from showing in directory listing.
       -description    Change server description in auto generated caps.txt.
       -admin-email    Change admin email in auto generated caps.txt.
       -geoloc         Change geolocation in auto generated caps.txt.
       -version        Print version string.
```

# Features

- Built with concurrency and efficiency in mind.

- ZERO external dependencies.

- Security focused -- chroots into server direrctory and drops
  privileges. `maybe wait until stable release before use outside of hobby
  setups.`

- LRU file caching -- with user-controlled cache size, max cached file size
  and cache refresh frequency.

- Insert files within gophermaps, including automating reflowing of lines
  longer than (user definable) page width.

- Automatic replacement of `$hostname` or `$port` with the information of
  the host the client is connecting to.

- Item type characters beyond RFC 1436 standard (see below).

- Separate system and access logging with output to file if requested (or to
  disable both).

# Supported gophermap item types

All of the following item types are supported by Gophor, separated into
grouped standards. Most handling of item types is performed by the clients
connecting to Gophor, but when performing directory listings Gophor will
attempt to automatically classify files according to the below types.

Item types listed as `[SERVER ONLY]` means that these are item types
recognised ONLY by Gophor and to be used when crafting a gophermap. They
provide additional methods of formatting / functionality within a gophermap,
and the output of these item types is usually converted to informational
text lines before sending to connecting clients.

```
RFC 1436 Standard:
Type | Treat as | Meaning
 0   |   TEXT   | Regular file (text)
 1   |   MENU   | Directory (menu)
 2   | EXTERNAL | CCSO flat db; other db
 3   |  ERROR   | Error message
 4   |   TEXT   | Macintosh BinHex file
 5   |  BINARY  | Binary archive (zip, rar, 7zip, tar, gzip, etc)
 6   |   TEXT   | UUEncoded archive
 7   |   INDEX  | Query search engine or CGI script
 8   | EXTERNAL | Telnet to: VT100 series server
 9   |  BINARY  | Binary file (see also, 5)
 T   | EXTERNAL | Telnet to: tn3270 series server
 g   |  BINARY  | GIF format image file (just use I)
 I   |  BINARY  | Any format image file
 +   |     -    | Redundant (indicates mirror of previous item)

GopherII Standard:
Type | Treat as | Meaning
 c   |  BINARY  | Calendar file
 d   |  BINARY  | Word-processing document; PDF document
 h   |   TEXT   | HTML document
 i   |     -    | Informational text (not selectable)
 p   |   TEXT   | Page layout or markup document (plain text w/ ASCII tags)
 m   |  BINARY  | Email repository (MBOX)
 s   |  BINARY  | Audio recordings
 x   |   TEXT   | eXtensible Markup Language document
 ;   |  BINARY  | Video files

Commonly used:
Type | Treat as | Meaning
 !   |     -    | [SERVER ONLY] Menu title (set title ONCE per gophermap)
 #   |     -    | [SERVER ONLY] Comment, rest of line is ignored
 -   |     -    | [SERVER ONLY] Hide file/directory from directory listing
 .   |     -    | [SERVER ONLY] Last line -- stop processing gophermap default
 *   |     -    | [SERVER ONLY] Last line + directory listing -- stop processing
     |          |               gophermap and end on a directory listing
 =   |     -    | [SERVER ONLY] Include subgophermap / regular file here. Prints
     |          |               and formats file / gophermap in-place

Planned to be supported:
Type | Treat as | Meaning
 $   |     -    | [SERVER ONLY] Execute shell command and print stdout here
```

# Compliance

## Item types

Supported item types are listed above.

Informational lines are sent as `i<text here>\t/\tnull.host\t0`.

Titles are sent as `i<title text>\tTITLE\tnull.host\t0`.

Web address links are sent as `h<text here>\tURL:<address>\thostname\tport`.
An HTML redirect is sent in response to any requests beginning with `URL:`.

## Policy files

Upon request, `caps.txt` can be provided from the server root directory
containing server capabiities. This can either be user or server generated.

Upon request, `robots.txt` can be provided from the server root directory
containing robot access restriction policies. This can either be user or
server generated.

## Errors

Errors are sent according to GopherII standards, terminating with a last
line:
`3<error text>CR-LF`

Possible Gophor errors:
```
     Text                 |      Meaning
400 Bad Request           | Request not understood by server due to malformed
                          | syntax
401 Unauthorised          | Request requires authentication
403 Forbidden             | Request received but not fulfilled
404 Not Found             | Server could not find anything matching requested
                          | URL
408 Request Time-out      | Client did not produce request within server wait
                          | time
410 Gone                  | Requested resource no longer available with no
                          | forwarding address
500 Internal Server Error | Server encountered an unexpected condition which
                          | prevented request being fulfilled
501 Not Implemented       | Server does not support the functionality
                          | required to fulfil the request
503 Service Unavailable   | Server currently unable to handle the request
                          | due to temporary overload / maintenance
```

## Terminating full stop

Gophor will send a terminating full-stop for menus, but not for served
files.

## Placeholder text

All of the following are used as placeholder text in responses...

Null selector: `-`

Null host: `null.host`

Null port: `0`

# Todos

Shortterm:

- Set default charset -- need to think about implementation here...

- Fix file cache only updating if main gophermap changes (but not sub files)
  -- need to either rethink how we keep track of files, or rethink how
  gophermaps are stored in memory.

- Add more files to file extension map

Longterm:

- Finish inline shell scripting support -- current thinking is to either
  perform a C fork very early on, or create a separate modules binary, and
  either way the 2 processes interact via some IPC method. Could allow for
  other modules too.

- Rotating logs -- have a check on start for a file-size, rotate out if the
  file is too large. Possibly checks during run-time too?

- Add last-mod-time to directory listings -- have global time parser
  object, maybe separate out separate global instances of objects (e.g.
  worker related, cache related, config related?)

- TLS support -- ~~requires a rethink of how we're passing port functions
  generating gopher directory entries, also there is no definitive standard
  for this yet~~ implemented these changes! figuring out gopher + TLS itself
  though? no luck yet.

- Connection throttling + timeouts -- thread to keep track of list of
  recently connected IPs. Keep incremementing connection count and only
  remove from list when `lastIncremented` time is greater than timeout

- Header + footer text -- read in file / input string and format, hold in
  memory than append to end of gophermaps / dir listings

- More closely follow GoLang built-in net/http code style for worker -- just
  a neatness thing, maybe bring some performance improvements too and a
  generally different way of approaching some of the solutions to problems we
  have

# Please note

During the initial writing phase the quality of git commit messages may be
low and many changes are likely to be bundled together at a time, just
because the pace of development right now is rather break-neck.

As soon as we reach a stable point in development, or if other people start
contributing issues or PRs, whichever comes first, this will be changed
right away.

# Resources used

Gopher-II (The Next Generation Gopher WWIS):
https://tools.ietf.org/html/draft-matavka-gopher-ii-00

Gophernicus supported item types:
https://github.com/gophernicus/gophernicus/blob/master/README.gophermap

All of the below can be viewed from your standard web browser using
floodgap's Gopher proxy:
https://gopher.floodgap.com/gopher/gw

RFC 1436 (The Internet Gopher Protocol:
gopher://gopher.floodgap.com:70/0/gopher/tech/rfc1436.txt

Gopher+ (upward compatible enhancements):
gopher://gopher.floodgap.com:70/0/gopher/tech/gopherplus.txt
