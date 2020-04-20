#!/bin/sh

PROJECT='gophor'
OUTDIR='build'

build() {
    echo "Building $PROJECT for $2_$3..."
    CC="$1" CGO_ENABLED=1 GOOS="$2" GOARCH="$3" go build -trimpath -o="$OUTDIR/$PROJECT.$2.$3" -buildmode="$4" -a -tags "$5" -ldflags "$6"
    echo ''
}

echo "PLEASE BE WARNED THIS SCRIPT IS WRITTEN FOR MY VOID LINUX BUILD ENVIRONMENT"
echo "YOUR CC CROSS-COMPILER LOCATIONS MAY DIFFER ON YOUR BUILD SYSTEM"
echo ""

# Clean and recreate directory
rm -rf "$OUTDIR"
mkdir -p "$OUTDIR"

# Build time :)
build 'i686-linux-musl-gcc'        'linux' '386'     'pie'     'netgo' '-s -w -extldflags "-static"'
build 'x86_64-linux-musl-gcc'      'linux' 'amd64'   'pie'     'netgo' '-s -w -extldflags "-static"'
build 'arm-linux-musleabi-gcc'     'linux' 'arm'     'pie'     'netgo' '-s -w -extldflags "-static"'
build 'aarch64-linux-musl-gcc'     'linux' 'arm64'   'pie'     'netgo' '-s -w -extldflags "-static"'
build 'mips-linux-musl-gcc'        'linux' 'mips'    'default' 'netgo' '-s -w -extldflags "-static"'
#build 'powerpc64-linux-musl-gcc'   'linux' 'ppc64'   'default' 'netgo' '-s -w -extldflags "-static"'
build 'powerpc64le-linux-musl-gcc' 'linux' 'ppc64le' 'pie'     'netgo' '-s -w -extldflags "-static"'

echo "PLEASE DON'T JUDGE THIS SCRIPT, IT IS TRULY SO AWFUL. TO BE IMPROVED..."