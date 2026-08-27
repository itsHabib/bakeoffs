#!/bin/sh
set -eu

cd "$(dirname "$0")"
test -z "$(gofmt -l engine.go demo.go cmd/obligation-demo/main.go engine_test.go demo_test.go)"
go vet ./...
go test ./...
go run ./cmd/obligation-demo
