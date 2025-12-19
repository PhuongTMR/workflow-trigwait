#!/bin/bash
set -e

cd "$(dirname "$0")/.."

mkdir -p dist

echo "══════════════════════════════════════════════════════════════"
echo "  Building workflow-trigwait binaries"
echo "══════════════════════════════════════════════════════════════"
echo ""

build() {
    local os=$1
    local arch=$2
    local ext=$3
    local output="dist/workflow-trigwait-${os}-${arch}${ext}"

    printf "  %-20s" "${os}/${arch}..."
    CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build \
        -trimpath \
        -ldflags="-s -w -buildid=" \
        -o "$output" \
        ./cmd/main.go

    local size=$(ls -lh "$output" | awk '{print $5}')
    echo "✓ ${size}"
}

echo "▸ Compiling binaries..."
echo ""

build linux amd64 ""
build linux arm64 ""
build darwin amd64 ""
build darwin arm64 ""
build windows amd64 ".exe"

echo ""
echo "══════════════════════════════════════════════════════════════"

# Compress with UPX if available
if command -v upx &> /dev/null; then
    echo "▸ Compressing with UPX (LZMA)..."
    echo ""
    for bin in dist/workflow-trigwait-linux-* dist/workflow-trigwait-windows-*; do
        if [[ -f "$bin" ]]; then
            printf "  %-40s" "$(basename $bin)..."
            original_size=$(ls -lh "$bin" | awk '{print $5}')
            upx -q --best --lzma "$bin" > /dev/null 2>&1 || true
            compressed_size=$(ls -lh "$bin" | awk '{print $5}')
            echo "✓ ${original_size} → ${compressed_size}"
        fi
    done
    echo ""
    echo "══════════════════════════════════════════════════════════════"
fi

echo "▸ Build complete!"
echo ""
ls -lh dist/
echo ""
