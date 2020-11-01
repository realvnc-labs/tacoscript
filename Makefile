# set default shell
SHELL = bash -e -o pipefail

# Variables
VERSION                  ?= $(shell cat ./VERSION)

default: build-win

help:
	@echo "Usage: make [<target>]"
	@echo "where available targets are:"
	@echo
	@echo "build             : Build Tacoscript binary for the current OS"
	@echo "build-win         : Build Tacoscript binary for Windows"
	@echo "build-nix         : Build Tacoscript binary for Linux"
	@echo "help              : Print this help"
	@echo "test              : Run unit tests, if any"
	@echo "sca               : Run SCA"
	@echo

build-win:
	mkdir -p bin
	GOOS=windows GOARCH=386 go build -o bin/tacoscript.exe main.go

build-nix:
	mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -o bin/tacoscript_linux main.go

build:
	mkdir -p bin
	go build -o bin/tacoscript_mac main.go

test:
	go test -race -v -p 1 ./...

sca:
	golangci-lint run
