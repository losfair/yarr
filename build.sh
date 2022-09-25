#!/bin/bash

set -e

VERSION=2.3
GITHASH=$(git rev-parse --short=8 HEAD)
GO_LDFLAGS="-s -w -X 'main.Version=$VERSION' -X 'main.GitHash=$GITHASH'"
go build -tags "sqlite_foreign_keys release libsqlite3" -ldflags="$GO_LDFLAGS" -o yarr src/main.go

