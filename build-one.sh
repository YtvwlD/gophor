#!/bin/sh

echo "Building for current platform..."
CGO_ENABLED=1 go build -trimpath -o 'gophor' -buildmode 'pie' -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
upx --best --color 'gophor'
echo ""
