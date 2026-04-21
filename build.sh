#!/usr/bin/env bash
set -euo pipefail

NAME="claude-fish"
VERSION="v2.1.116"
DIST="dist"

# Clean previous builds
rm -rf "$DIST"
mkdir -p "$DIST"

build() {
	local os=$1 arch=$2 ext=$3
	local dir="${NAME}-${VERSION}-${os}-${arch}"
	local outdir="$DIST/$dir"
	local bin="$outdir/$NAME$ext"

	echo "==> Building $os/$arch..."
	mkdir -p "$outdir"
	CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -ldflags="-s -w" -o "$bin" .

	# Copy novel directory (exclude .DS_Store)
	mkdir "$outdir/novel"
	find novel -type f ! -name '.DS_Store' -exec cp {} "$outdir/novel/" \;

	# Windows: add a launcher bat
	if [ "$os" = "windows" ]; then
		cat > "$outdir/start.bat" <<'BAT'
@echo off
chcp 65001 >nul 2>&1
"%~dp0claude-fish.exe" %*
pause
BAT
	fi

	# Package
	if [ "$os" = "windows" ]; then
		(cd "$DIST" && zip -r -q "${dir}.zip" "$dir")
		echo "    -> $DIST/${dir}.zip"
	else
		tar -czf "$DIST/${dir}.tar.gz" -C "$DIST" "$dir"
		echo "    -> $DIST/${dir}.tar.gz"
	fi

	rm -rf "$outdir"
}

# macOS
build darwin  arm64 ""
build darwin  amd64 ""

# Linux
build linux   amd64 ""
build linux   arm64 ""

# Windows
build windows amd64 ".exe"
build windows arm64 ".exe"

echo ""
echo "Done. Build artifacts in $DIST/:"
ls -lh "$DIST"/*
