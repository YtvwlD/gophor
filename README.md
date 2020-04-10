# Gophor

A Gopher server written in GoLang as a means of learning about the Gopher
protocol, and more GoLang.

Possibly beta quality soon? More likely alpha right now.

The Client <--> ClientManager setup of handling newly spawned goroutines for
new connections is a little over-complicated for the setup we have now, but
allows for much growth and easier handling in the future.

Tries to adhere to RFC 1436 as much as possible but this is all a work in
progress.

# Todos

- TLS support
