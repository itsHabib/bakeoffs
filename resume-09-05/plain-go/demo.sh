#!/bin/sh
set -eu
cd "$(dirname "$0")"
mkdir -p .bin
go build -o .bin/resume .
exec .bin/resume demo --dir artifacts
