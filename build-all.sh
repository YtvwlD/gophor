#!/bin/sh

PROJECT='gophor'
OUTDIR='build'

build() {
    echo "Building $PROJECT for $1_$2..."
    GOOS="$1" GOARCH="$2" go build -trimpath -o="$OUTDIR/$PROJECT.$1.$2" -buildmode="$3" -a -tags 'netgo' -ldflags '-w -extldflags "-static"'
    echo ''
}

mkdir -p "$OUTDIR"
build 'linux' '386'   'pie'
build 'linux' 'amd64' 'pie'
build 'linux' 'arm'   'pie'
build 'linux' 'arm64' 'pie'
build 'linux' 'mips'  'default'
