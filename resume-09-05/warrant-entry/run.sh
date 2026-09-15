#!/bin/sh
set -eu
cd "$(dirname "$0")"
export GOPROXY=off GOTOOLCHAIN=local GOSUMDB=off
./verify-source.sh
mkdir -p .local
go build -o .local/resume .
case "${1:-demo}" in
  demo) shift "$(( $# > 0 ? 1 : 0 ))"; exec .local/resume demo "$@" ;;
  test) go test -count=1 ./...; go test -count=1 github.com/itsHabib/warrant ;;
  *) exec .local/resume "$@" ;;
esac
