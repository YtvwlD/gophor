# Gophor

A Gopher server written in GoLang as a means of learning about the Gopher
protocol, and more GoLang.

Linux only _for now_. Cross-compiled to way too many architectures.
Build-script now much improved, but still not pretty...

I'm unemployed and work on open-source projects like this and many others for
free. If you would like to help support my work that would be hugely
appreciated 💕 https://liberapay.com/grufwub/

WARNING: the development branch is filled with lava, fear and capitalism.

# Features

- Built with security, concurrency and efficiency in mind.

- ZERO external dependencies.

- LRU file caching with user-controlled cache size, max cached file size
  and cache refresh frequency.

- CGI-bin support with CGI 1.1 compliance.

- Executable gophermap support.

- Insert files with automated line reflowing, and inline shell script
  output within gophermaps.

- Support for all commonly accepted item type characters (beyond just
  RFC1436 support).

- Automatic replacement of `$hostname` or `$port` in gophermap lines with
  current host information.

- User supplied footer text appended to gophermaps and directory listings.

- Separate system and access logging with output and formatting options.

*Please note: previously, chrooting to server directory and dropping
privileges was supported by using Go C bindings. Unexpected circumstances
have not yet been witness... But as this is not officially supported due to
unexpected behaviour witnessed with `.Set{U,G}id()`, and there is a near 10 year 
ongoing tracked issue (https://github.com/golang/go/issues/1435), I decided to
drop this feature for now. As soon as this patch gets merged I'll add
support: https://go-review.googlesource.com/c/go/+/210639

As a result of removing this, request sanitization has been majorly improved
and checks are in place to prevent running Gophor as root.

If you run into issues binding to a lower port number due to insufficient
permissions then there are a few alternatives:

- set process capabilities using utility like capsh:
  https://linux.die.net/man/1/capsh

- use Docker (or some other solution) and configure port forwarding on the
  host

- start gopher in it's own namespace in a chroot*


# Usage

```
gophor [args]
       -root-dir            Change server root directory.
       -port                Change server listening port.
       -hostname            Change server hostname (FQDN, used to craft dir
                            lists).
       -bind-addr           Change server bind-address (used in creating
                            socket).
       -env                 New-line separated list of environment variables
                            to use when executing cgi-bin, gophermaps, shell
                            commands.

       -system-log          Path to gophor system log file.
       -access-log          Path to gophor access log file.
       -log-output          Change log output type (disable|stderr|file)
       -log-opts            Comma-separated list of lop opts (timestamp|ip)

       -cache-check         Change file-cache freshness check frequency.
       -cache-size          Change max no. files in file-cache.
       -cache-file-max      Change maximum allowed size of a cached file.
       -disable-cache       Disable file caching.

       -page-width          Change page width used when formatting output.
       -footer              Change gophermap footer text (Unix new-line
                            separated lines).
       -no-footer-separator Disable footer text line separator.
       -restrict-files      New-line separated list of regex statements
                            restricting files from showing in directory listing.

       -description         Change server description in generated caps.txt.
       -admin-email         Change admin email in generated caps.txt.
       -geoloc              Change geolocation in generated caps.txt.

       -version             Print version string.
```

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
 =   |     -    | [SERVER ONLY] Include or execute subgophermap, cgi-bin, regular
     |          |               file or shell command stdout here
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

- Character encoding support

- Fix file cache only updating if main gophermap changes (but not sub files)

Longterm:

- Rotating logs

- Add last-mod-time to directory listings

- TLS support

- Connection throttling + timeouts

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
