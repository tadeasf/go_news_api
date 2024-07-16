#!/bin/bash
export GIN_MODE=release
go mod tidy
go build -o ./dist/go_news_api -ldflags='-w -s' .
