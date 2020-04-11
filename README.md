# Gophor

A Gopher server written in GoLang as a means of learning about the Gopher
protocol, and more GoLang.

Possibly beta quality soon? More likely alpha right now.

Tries to adhere to RFC 1436 as much as possible but this is all a work in
progress. I also generally set out to use as few outside dependencies as
possible, which was achieved here with 0 libraries outside GoLang standard
used.

Separate system and access loggers exist, with stderr as default if no file
supplied for both. Though the formatted print of each log line is still open
to changes as right now it feels a bit too hard to visually parse.

Linux only FOR NOW.

# Todos

- TLS support
