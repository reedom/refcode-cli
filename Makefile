## meta
NAME := refcode-mapper
VERSION := $(shell git describe --tags --abbrev=0)
REVISION := $(shell git rev-parse --short HEAD)
LDFLAGS := -X 'github.com/reedom/refcode-cli/cmd.version=$(VERSION)' \
	-X 'github.com/reedom/refcode-cli/cmd.revision=$(REVISION)'
SRC := $(shell find . -name '*.go')

## setup
setup:
	go get github.com/Masterminds/glide
	go get github.com/golang/lint/golint
	go get golang.org/x/tools/cmd/goimports
	go get github.com/Songmu/make2help/cmd/make2help

## run tests
test: deps
	go test $$(glide novendor)

## install dependencies
deps: setup
	glide install

# update dependencies
update: setup
	glide update

## lint
lint: setup
	go vet $$(glide novendor)
	for pkg in $$(glide novendor -x); do
		golint -set_exit_status $$pkg || exit $$?;
	done

fmt: setup
	goimports -w $$(glide nv -x)

## build binaries
build: $(SRC)
	go build -ldflags "$(LDFLAGS)" -o bin/refcode main.go

## show help
help:
	@make2help $(MAKEFILE_LIST)

version:
	@echo $(SRC)

.PHONY: setup deps update test lint help version
