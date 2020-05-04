#!/bin/sh

set -e

PROJECT='gophor'
OUTDIR='build'
VERSION="$(cat 'gophor.go' | grep -E '^\s*GophorVersion' | sed -e 's|\s*GophorVersion = \"||' -e 's|\"\s*$||')"
GOVERSION="$(go version | sed -e 's|^go version go||' -e 's|\s.*$||')"
LOGFILE='build.log'

silent() {
    "$@" > "$LOGFILE" 2>&1
}

build_for() {
    local archname="$1" toolchain="$2" os="$3" arch="$4"
    shift 4
    if [ "$arch" = 'arm' ]; then
        local armversion="$1"
        shift 1
    fi

    echo "Building for ${os} ${archname}..."
    local filename="${OUTDIR}/${PROJECT}_${os}_${archname}"
    CGO_ENABLED=1 CC="$toolchain" GOOS="$os" GOARCH="$arch" GOARM="$armversion" silent go build -trimpath -o "$filename" "$@"
    if [ "$?" -ne 0 ]; then
        echo "Failed!"
        return 1
    fi

    echo "Compressing ${filename}..."
    silent upx --best "$filename"
    silent upx -t "$filename"
    echo ""
}

# Build time :)
build_for 'amd64' 'x86_64-linux-musl-gcc' 'linux' 'amd64' -buildmode 'pie' -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
