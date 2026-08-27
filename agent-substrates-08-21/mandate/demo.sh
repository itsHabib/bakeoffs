#!/bin/sh
set -eu

repo_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
cd "$repo_dir"

test -z "$(gofmt -l cmd/mandate/*.go mandate/*.go)"
(cd fixtures && shasum -a 256 -c SHA256SUMS)
go vet ./...
go test -race -count=1 ./...
go run ./cmd/mandate demo
