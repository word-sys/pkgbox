#!/usr/bin/env bash
set -euo pipefail

echo "Running PkgBox Tests"

echo "[1/2] Running Go Unit Tests"
go test -v ./...

echo "[2/2] Compiling PkgBox Binary"
mkdir -p bin
go build -o bin/pkgbox ./cmd/pkgbox

if [ -x bin/pkgbox ]; then
    echo "OK: bin/pkgbox built and executable."
    ls -lh bin/pkgbox
else
    echo "ERROR: bin/pkgbox build failed." >&2
    exit 1
fi

echo "All Tests Passed!"
