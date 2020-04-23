#!/bin/sh

set -e

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
CGO_ENABLED=1 CC='i686-linux-musl-gcc'         GOOS='linux' GOARCH='386'          go build -trimpath -o "$OUTDIR/$PROJECT.linux.386"      -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
upx --best "$OUTDIR/$PROJECT.linux.386"
upx -t "$OUTDIR/$PROJECT.linux.386"
echo ""

echo "Building for linux amd64..."
CGO_ENABLED=1 CC='x86_64-linux-musl-gcc'       GOOS='linux' GOARCH='amd64'        go build -trimpath -o "$OUTDIR/$PROJECT.linux.amd64"    -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
upx --best "$OUTDIR/$PROJECT.linux.amd64"
upx -t "$OUTDIR/$PROJECT.linux.amd64"
echo ""

echo "Building for linux armv5..."
CGO_ENABLED=1 CC='arm-linux-musleabi-gcc'      GOOS='linux' GOARCH='arm'  GOARM=5 go build -trimpath -o "$OUTDIR/$PROJECT.linux.armv5"    -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
upx --best "$OUTDIR/$PROJECT.linux.armv5"
upx -t "$OUTDIR/$PROJECT.linux.armv5"
echo ""

echo "Building for linux armv5hf..."
CGO_ENABLED=1 CC='arm-linux-musleabihf-gcc'    GOOS='linux' GOARCH='arm'  GOARM=5 go build -trimpath -o "$OUTDIR/$PROJECT.linux.armv5hf"  -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
upx --best "$OUTDIR/$PROJECT.linux.armv5hf"
upx -t "$OUTDIR/$PROJECT.linux.armv5hf"
echo ""

echo "Building for linux armv6..."
CGO_ENABLED=1 CC='arm-linux-musleabi-gcc'      GOOS='linux' GOARCH='arm'  GOARM=6 go build -trimpath -o "$OUTDIR/$PROJECT.linux.armv6"    -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
upx --best "$OUTDIR/$PROJECT.linux.armv6"
upx -t "$OUTDIR/$PROJECT.linux.armv6"
echo ""

echo "Building for linux armv6hf..."
CGO_ENABLED=1 CC='arm-linux-musleabihf-gcc'    GOOS='linux' GOARCH='arm'  GOARM=6 go build -trimpath -o "$OUTDIR/$PROJECT.linux.armv6hf"  -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
upx --best "$OUTDIR/$PROJECT.linux.armv6hf"
upx -t "$OUTDIR/$PROJECT.linux.armv6hf"
echo ""

echo "Building for linux armv7hf..."
CGO_ENABLED=1 CC='armv7l-linux-musleabihf-gcc' GOOS='linux' GOARCH='arm'  GOARM=7 go build -trimpath -o "$OUTDIR/$PROJECT.linux.armv7hf"  -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
upx --best "$OUTDIR/$PROJECT.linux.armv7hf"
upx -t "$OUTDIR/$PROJECT.linux.armv7hf"
echo ""

echo "Building for linux arm64..."
CGO_ENABLED=1 CC='aarch64-linux-musl-gcc'      GOOS='linux' GOARCH='arm64'        go build -trimpath -o "$OUTDIR/$PROJECT.linux.arm64"    -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
upx --best "$OUTDIR/$PROJECT.linux.arm64"
upx -t "$OUTDIR/$PROJECT.linux.arm64"
echo ""

echo "Building for linux mips..."
CGO_ENABLED=1 CC='mips-linux-musl-gcc'         GOOS='linux' GOARCH='mips'         go build -trimpath -o "$OUTDIR/$PROJECT.linux.mips"     -buildmode 'default' -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
upx --best "$OUTDIR/$PROJECT.linux.mips"
upx -t "$OUTDIR/$PROJECT.linux.mips"
echo ""

echo "Building for linux mipshf..."
CGO_ENABLED=1 CC='mips-linux-muslhf-gcc'       GOOS='linux' GOARCH='mips'         go build -trimpath -o "$OUTDIR/$PROJECT.linux.mipshf"   -buildmode 'default' -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
upx --best "$OUTDIR/$PROJECT.linux.mipshf"
upx -t "$OUTDIR/$PROJECT.linux.mipshf"
echo ""

echo "Building for linux mipsle..."
CGO_ENABLED=1 CC='mipsel-linux-musl-gcc'       GOOS='linux' GOARCH='mipsle'       go build -trimpath -o "$OUTDIR/$PROJECT.linux.mipsle"   -buildmode 'default' -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
upx --best "$OUTDIR/$PROJECT.linux.mipsle"
upx -t "$OUTDIR/$PROJECT.linux.mipsle"
echo ""

echo "Building for linux mipslehf..."
CGO_ENABLED=1 CC='mipsel-linux-muslhf-gcc'     GOOS='linux' GOARCH='mipsle'       go build -trimpath -o "$OUTDIR/$PROJECT.linux.mipslehf" -buildmode 'default' -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
upx --best "$OUTDIR/$PROJECT.linux.mipslehf"
upx -t "$OUTDIR/$PROJECT.linux.mipslehf"
echo ""

echo "Building for linux ppc64le..."
CGO_ENABLED=1 CC='powerpc64le-linux-musl-gcc'  GOOS='linux' GOARCH='ppc64le'      go build -trimpath -o "$OUTDIR/$PROJECT.linux.ppc64le"  -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
upx --best "$OUTDIR/$PROJECT.linux.ppc64le"
upx -t "$OUTDIR/$PROJECT.linux.ppc64le"
echo ""

echo "PLEASE DON'T JUDGE THIS SCRIPT, IT IS TRULY SO AWFUL. TO BE IMPROVED..."
