.PHONY: build build-alpine clean test help default

BIN_NAME=gosafex

VERSION := $(shell grep "const Version " version/version.go | sed -E 's/.*"(.+)"$$/\1/')
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_DIRTY=$(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
BUILD_DATE=$(shell date '+%Y-%m-%d-%H:%M:%S')
IMAGE_NAME := "atanmarko/gosafex"

default: test

help:
	@echo 'Management commands for gosafex:'
	@echo
	@echo 'Usage:'
	@echo '    make build           Compile the project.'
	@echo '    make clean           Clean the directory tree.'
	@echo

build:
	@echo "building ${BIN_NAME} ${VERSION}"
	@echo "GOPATH=${GOPATH}"
	@echo "commit=${GIT_COMIT}${GIT_DIRTY}"
	@echo "build date=${BUILD_DATE}"
	go build -ldflags "-X github.com/safex/gosafex/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/safex/gosafex/version.BuildDate=${BUILD_DATE}" -o bin/${BIN_NAME}

clean:
	@test ! -e bin/${BIN_NAME} || rm bin/${BIN_NAME}

test:
	go test ./...

