#!/bin/bash -x

VERSION=?
GIT_COMMIT=$(git rev-list -1 HEAD)
SHA1=$(git rev-parse HEAD)
BRANCH=$(git rev-parse --abbrev-ref HEAD)
NOW=$(date +'%Y-%m-%d_%T')
LDFLAGS="-X github.com/etowett/requestl/build.Sha1Ver=${SHA1} -X github.com/etowett/requestl/build.Time=${NOW} -X github.com/etowett/requestl/build.GitCommit=${GIT_COMMIT} -X github.com/etowett/requestl/build.GitBranch=${BRANCH} -X github.com/etowett/requestl/build.Version=${VERSION}"

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "${LDFLAGS}" -a -o requestl cmd/requestl/main.go
