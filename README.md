# Gophor

A Gopher server written in GoLang as a means of learning about the Gopher
protocol, and more GoLang.

Possibly beta quality soon? More likely alpha right now.

`go build` and you're set to go! (...phor ha)

Linux only FOR NOW.

# Features

- ZERO external dependencies

- Item type characters beyond RFC 1436 standard (see below)

- Security focused -- `Until we reach beta/stable, maybe don't trust this...`

- File caching -- Until LRU is implemented, this means cache size is infinite,
  which for a gopher server of very limited size, this is fine. Probably.

- Separate system and access logging with output to file if requested (or
  disable both!)

# Supported gophermap item types

```
0 -- regular file (text)
1 -- directory (menu)
2 -- CSO phone-book server... should you be using this in 2020 lmao
3 -- Error
4 -- Binhexed macintosh file
5 -- DOS bin archive
6 -- Unix uuencoded file
7 -- Index-search server
8 -- Text-based telnet session
9 -- Binary file
T -- Text-based tn3270 session... in 2020???
g -- Gif format graphic
I -- Image file of some kind

+ -- Redundant server

. -- Lastline if online this followed by CrLf

i -- Info message
h -- HTML document
s -- Audio file
p -- PNG image
d -- Document

M -- MIME type file
; -- Video file
c -- Calendar file
! -- Title
# -- Comment (not displayed)
- -- Hide file from directory listing
= -- Include subgophermap (prints file output here)
* -- Act as-if lastline and print directory listing below

Unavailable for now due to issues with accessing path within chroot:
~~$ -- Execute shell command and print stdout here~~
```

# Todos

- TLS support

- Connection throttling + timeouts

- Header + footer text

- ~~Support proposed protocol extensions (Gopher+ etc)~~

- ~~Toggleable logging~~

- Rotating logs

- Set default page width (modifies max UserName / Selector fields)

- Set default charset

- Autogenerated caps.txt

- Server status page (?)

- Proxy over HTTP support

- Gopher servemux (?)

- ~~Treat plain line of text (no tabs) as info line~~

- Finish inline shell scripting support

- Implement LRU in file cache

- Allow setting UID+GID via username string

- Neaten-up newly added file caching code

# Please note

During the initial writing phase the quality of git commit messages may be
low and many changes are likely to be bundled together at a time, just
because the pace of development right now is rather break-neck.

As soon as we reach a stable point in development, or if other people start
contributing issues or PRs, whichever comes first, this will be changed
right away.

# Standards followed

Gopher-II (The Next Generation Gopher WWIS):
https://tools.ietf.org/html/draft-matavka-gopher-ii-00

All of the below can be viewed from your standard web browser using
floodgap's Gopher proxy:
https://gopher.floodgap.com/gopher/gw

RFC 1436 (The Internet Gopher Protocol:
gopher://gopher.floodgap.com:70/0/gopher/tech/rfc1436.txt

Gopher+ (upward compatible enhancements):
gopher://gopher.floodgap.com:70/0/gopher/tech/gopherplus.txt
