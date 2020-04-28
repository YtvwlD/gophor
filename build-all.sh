#!/bin/sh

set -e

PROJECT='gophor'
OUTDIR='build'
VERSION="$(cat 'constants.go' | grep -E '^\s*GophorVersion' | sed -e 's|\s*GophorVersion = \"||' -e 's|\"\s*$||')"
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

echo "PLEASE BE WARNED THIS SCRIPT IS WRITTEN FOR A VOID LINUX (MUSL) BUILD ENVIRONMENT"
echo "YOUR CC TOOLCHAIN LOCATIONS MAY DIFFER"
echo "IF THE SCRIPT FAILS, CHECK THE OUTPUT OF: ${LOGFILE}"
echo ""

# Clean and recreate directory
rm -rf "$OUTDIR"
mkdir -p "$OUTDIR"

# Build time :)
build_for '386'      'i686-linux-musl-gcc'         'linux' '386'     -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'

build_for 'amd64'    'x86_64-linux-musl-gcc'       'linux' 'amd64'   -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'

build_for 'armv5'    'arm-linux-musleabi-gcc'      'linux' 'arm' '5' -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'

build_for 'armv5hf'  'arm-linux-musleabihf-gcc'    'linux' 'arm' '5' -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'

build_for 'armv6'    'arm-linux-musleabi-gcc'      'linux' 'arm' '6' -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'

build_for 'armv6hf'  'arm-linux-musleabihf-gcc'    'linux' 'arm' '6' -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'

build_for 'armv7lhf' 'armv7l-linux-musleabihf-gcc' 'linux' 'arm' '7' -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'

build_for 'arm64'    'aarch64-linux-musl-gcc'      'linux' 'arm64'   -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'

build_for 'mips'     'mips-linux-musl-gcc'         'linux' 'mips'    -buildmode 'default' -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'

build_for 'mipshf'   'mips-linux-muslhf-gcc'       'linux' 'mips'    -buildmode 'default' -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'

build_for 'mipsle'   'mipsel-linux-musl-gcc'       'linux' 'mipsle'  -buildmode 'default' -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'

build_for 'mipslehf' 'mipsel-linux-muslhf-gcc'     'linux' 'mipsle'  -buildmode 'default' -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'

build_for 'ppc64le'  'powerpc64le-linux-musl-gcc'  'linux' 'ppc64le' -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
