#!/bin/sh
set -e

if ! command -v go &> /dev/null; then
    echo "Go is not installed, please install Go first"
    exit 1
fi

if ! command -v golangci-lint &> /dev/null; then
    go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@latest
fi