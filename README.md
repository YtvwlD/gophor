# Gophor

A Gopher server written in GoLang as a means of learning about the Gopher
protocol, and more GoLang.

Possibly beta quality soon? More likely alpha right now.

The Client <--> ClientManager setup of handling newly spawned goroutines for
new connections is a little over-complicated for the setup we have now, but
allows for much growth and easier handling in the future.

Tries to adhere to RFC 1436 as much as possible but this is all a work in
progress. I also generally set out to use as few outside dependencies as
possible, which was achieved here with 0 libraries outside GoLang standard
used.

Linux only FOR NOW.

# Todos

- TLS support
