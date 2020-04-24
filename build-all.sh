#!/bin/sh

set -e

PROJECT='gophor'
OUTDIR='build'
VERSION="$(cat 'constants.go' | grep -E '^\s*GophorVersion' | sed -e 's|\s*GophorVersion = \"||' -e 's|\"\s*$||')"
echo "VERSION: $VERSION"
GOVERSION="$(go version | sed -e 's|^go version go||' -e 's|\s.*$||')"

echo "PLEASE BE WARNED THIS SCRIPT IS WRITTEN FOR MY VOID LINUX BUILD ENVIRONMENT"
echo "YOUR CC CROSS-COMPILER LOCATIONS MAY DIFFER ON YOUR BUILD SYSTEM"
echo ""

# Clean and recreate directory
rm -rf "$OUTDIR"
mkdir -p "$OUTDIR"

# Build time :)
echo "Building for linux 386..."
filename="${OUTDIR}/${PROJECT}.${VERSION}_linux.386_${GOVERSION}"
CGO_ENABLED=1 CC='i686-linux-musl-gcc'         GOOS='linux' GOARCH='386'          go build -trimpath -o "$filename"      -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
upx --best "$filename"
upx -t "$filename"
echo ""

echo "Building for linux amd64..."
filename="${OUTDIR}/${PROJECT}.${VERSION}_linux.amd64_${GOVERSION}"
CGO_ENABLED=1 CC='x86_64-linux-musl-gcc'       GOOS='linux' GOARCH='amd64'        go build -trimpath -o "$filename"    -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
upx --best "$filename"
upx -t "$filename"
echo ""

echo "Building for linux armv5..."
filename="${OUTDIR}/${PROJECT}.${VERSION}_linux.armv5_${GOVERSION}"
CGO_ENABLED=1 CC='arm-linux-musleabi-gcc'      GOOS='linux' GOARCH='arm'  GOARM=5 go build -trimpath -o "$filename"    -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
upx --best "$filename"
upx -t "$filename"
echo ""

echo "Building for linux armv5hf..."
filename="${OUTDIR}/${PROJECT}.${VERSION}_linux.armv5hf_${GOVERSION}"
CGO_ENABLED=1 CC='arm-linux-musleabihf-gcc'    GOOS='linux' GOARCH='arm'  GOARM=5 go build -trimpath -o "$filename"  -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
upx --best "$filename"
upx -t "$filename"
echo ""

echo "Building for linux armv6..."
filename="${OUTDIR}/${PROJECT}.${VERSION}_linux.armv6_${GOVERSION}"
CGO_ENABLED=1 CC='arm-linux-musleabi-gcc'      GOOS='linux' GOARCH='arm'  GOARM=6 go build -trimpath -o "$filename"    -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
upx --best "$filename"
upx -t "$filename"
echo ""

echo "Building for linux armv6hf..."
filename="${OUTDIR}/${PROJECT}.${VERSION}_linux.armv6hf_${GOVERSION}"
CGO_ENABLED=1 CC='arm-linux-musleabihf-gcc'    GOOS='linux' GOARCH='arm'  GOARM=6 go build -trimpath -o "$filename"  -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
upx --best "$filename"
upx -t "$filename"
echo ""

echo "Building for linux armv7hf..."
filename="${OUTDIR}/${PROJECT}.${VERSION}_linux.armv7hf_${GOVERSION}"
CGO_ENABLED=1 CC='armv7l-linux-musleabihf-gcc' GOOS='linux' GOARCH='arm'  GOARM=7 go build -trimpath -o "$filename"  -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
upx --best "$filename"
upx -t "$filename"
echo ""

echo "Building for linux arm64..."
filename="${OUTDIR}/${PROJECT}.${VERSION}_linux.arm64_${GOVERSION}"
CGO_ENABLED=1 CC='aarch64-linux-musl-gcc'      GOOS='linux' GOARCH='arm64'        go build -trimpath -o "$filename"    -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
upx --best "$filename"
upx -t "$filename"
echo ""

echo "Building for linux mips..."
filename="${OUTDIR}/${PROJECT}.${VERSION}_linux.mips_${GOVERSION}"
CGO_ENABLED=1 CC='mips-linux-musl-gcc'         GOOS='linux' GOARCH='mips'         go build -trimpath -o "$filename"     -buildmode 'default' -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
upx --best "$filename"
upx -t "$filename"
echo ""

echo "Building for linux mipshf..."
filename="${OUTDIR}/${PROJECT}.${VERSION}_linux.mipshf_${GOVERSION}"
CGO_ENABLED=1 CC='mips-linux-muslhf-gcc'       GOOS='linux' GOARCH='mips'         go build -trimpath -o "$filename"   -buildmode 'default' -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
upx --best "$filename"
upx -t "$filename"
echo ""

echo "Building for linux mipsle..."
filename="${OUTDIR}/${PROJECT}.${VERSION}_linux.mipsle_${GOVERSION}"
CGO_ENABLED=1 CC='mipsel-linux-musl-gcc'       GOOS='linux' GOARCH='mipsle'       go build -trimpath -o "$filename"   -buildmode 'default' -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
upx --best "$filename"
upx -t "$filename"
echo ""

echo "Building for linux mipslehf..."
filename="${OUTDIR}/${PROJECT}.${VERSION}_linux.mipslehf_${GOVERSION}"
CGO_ENABLED=1 CC='mipsel-linux-muslhf-gcc'     GOOS='linux' GOARCH='mipsle'       go build -trimpath -o "$filename" -buildmode 'default' -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
upx --best "$filename"
upx -t "$filename"
echo ""

echo "Building for linux ppc64le..."
filename="${OUTDIR}/${PROJECT}.${VERSION}_linux.ppc64le_${GOVERSION}"
CGO_ENABLED=1 CC='powerpc64le-linux-musl-gcc'  GOOS='linux' GOARCH='ppc64le'      go build -trimpath -o "$filename"  -buildmode 'pie'     -a -tags 'netgo' -ldflags '-s -w -extldflags "-static"'
upx --best "$filename"
upx -t "$filename
echo ""

echo "PLEASE DON'T JUDGE THIS SCRIPT, IT IS TRULY SO AWFUL. TO BE IMPROVED..."
