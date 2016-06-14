#!/usr/bin/env bash

source "${BASH_SOURCE%/*}/common.sh"

GOPATH=$PWD/_vendor:$GOPATH go test -v ./...
