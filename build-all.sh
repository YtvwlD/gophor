#!/bin/sh

PROJECT='gophor'
OUTDIR='build'

echo "PLEASE BE WARNED THIS SCRIPT IS WRITTEN FOR MY VOID LINUX BUILD ENVIRONMENT"
echo "YOUR CC CROSS-COMPILER LOCATIONS MAY DIFFER ON YOUR BUILD SYSTEM"
echo ""

# Clean and recreate directory
rm -rf "$OUTDIR"
mkdir -p "$OUTDIR"

# Build time :)
echo "Building for linux 386..."
CGO_ENABLED=1 CC='i686-linux-musl-gcc'        GOOS='linux' GOARCH='386'     go build -trimpath -o "$OUTDIR/$PROJECT.linux.386"     -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
echo ""

echo "Building for linux amd64..."
CGO_ENABLED=1 CC='x86_64-linux-musl-gcc'      GOOS='linux' GOARCH='amd64'   go build -trimpath -o "$OUTDIR/$PROJECT.linux.amd64"   -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
echo ""

echo "Building for linux arm..."
CGO_ENABLED=1 CC='arm-linux-musleabi-gcc'     GOOS='linux' GOARCH='arm'     go build -trimpath -o "$OUTDIR/$PROJECT.linux.arm"     -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
echo ""

echo "Building for linux arm64..."
CGO_ENABLED=1 CC='aarch64-linux-musl-gcc'     GOOS='linux' GOARCH='arm64'   go build -trimpath -o "$OUTDIR/$PROJECT.linux.arm64"   -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
echo ""

echo "Building for linux mips..."
CGO_ENABLED=1 CC='mips-linux-musl-gcc'        GOOS='linux' GOARCH='mips'    go build -trimpath -o "$OUTDIR/$PROJECT.linux.mips"    -buildmode 'default' -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
echo ""

echo "Building for linux ppc64le..."
CGO_ENABLED=1 CC='powerpc64le-linux-musl-gcc' GOOS='linux' GOARCH='ppc64le' go build -trimpath -o "$OUTDIR/$PROJECT.linux.ppc64le" -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
echo ""

echo "PLEASE DON'T JUDGE THIS SCRIPT, IT IS TRULY SO AWFUL. TO BE IMPROVED..."
