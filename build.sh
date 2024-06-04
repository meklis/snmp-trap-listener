#!/bin/bash
rm -Rf ./build
mkdir -p build
cp ./configs/trap-listener.yml ./internal/config
cp ./configs/trap-listener.yml ./build

go mod download
GOOS=linux GOARCH=amd64 go build -o ./build/traplistener ./cmd/listener/main.go
rm -f ./internal/config/trap-listener.yml


