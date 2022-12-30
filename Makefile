#!/usr/bin/env bash

PROJECTNAME=$(shell basename "$(PWD)")
COMPANY=Octopus Network
NAME=trie-go
ifndef VERSION
VERSION=latest
endif
FULLDOCKERNAME=$(COMPANY)/$(NAME):$(VERSION)

.PHONY: help lint test go.sum
all: help
help: Makefile
	@echo
	@echo " Choose a make command to run in "$(PROJECTNAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

.PHONY: lint
lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.48
	golangci-lint run


## test: Runs `go test` on project test files.
test:
	@echo "  >  \033[32mRunning tests...\033[0m "
	go test -coverprofile c.out ./... -timeout=30m

go.sum: go.mod
	echo "Ensure dependencies have not been modified ..." >&2
	go mod verify
	go mod tidy



