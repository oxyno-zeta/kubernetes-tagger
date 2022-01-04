# All possible values : darwin/amd64 linux/amd64 windows/amd64 linux/386 linux/ppc64le linux/s390x linux/arm linux/arm64
TARGETS           ?= linux/amd64
PROJECT_NAME	  := kubernetes-tagger
PKG				  := github.com/oxyno-zeta/$(PROJECT_NAME)

# go option
GO        ?= go
TAGS      :=
TESTS     := .
TESTFLAGS :=
LDFLAGS   := -w -s
GOFLAGS   := -i
BINDIR    := $(CURDIR)/bin
DISTDIR   := _dist

# Required for globs to work correctly
SHELL=/usr/bin/env bash

#  Version

GIT_COMMIT = $(shell git rev-parse HEAD)
GIT_SHA    = $(shell git rev-parse --short HEAD)
GIT_TAG    = $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
DATE	   = $(shell date +%F_%T%Z)

BINARY_VERSION ?= ${GIT_TAG}
ifeq ($(BINARY_VERSION),)
	BINARY_VERSION = ${GIT_SHA}
endif
# Clear the "unreleased" string in Metadata
ifneq ($(GIT_TAG),)
	LDFLAGS += -X ${PKG}/pkg/${PROJECT_NAME}/version.Metadata=
endif
LDFLAGS += -X ${PKG}/pkg/${PROJECT_NAME}/version.Version=${BINARY_VERSION}
LDFLAGS += -X ${PKG}/pkg/${PROJECT_NAME}/version.GitCommit=${GIT_COMMIT}
LDFLAGS += -X ${PKG}/pkg/${PROJECT_NAME}/version.BuildDate=${DATE}

#############
#   Build   #
#############

.PHONY: all
all: lint test build

.PHONY: lint
lint: dep
	golangci-lint run --timeout=3600s ./...

.PHONY: build
build: clean dep
	GOBIN=$(BINDIR) $(GO) install $(GOFLAGS) -tags '$(TAGS)' -ldflags '$(LDFLAGS)' $(PKG)/cmd/${PROJECT_NAME}

.PHONY: build-cross
build-cross: LDFLAGS += -extldflags "-static"
build-cross: clean dep
	CGO_ENABLED=0 gox -output="$(DISTDIR)/bin/$(BINARY_VERSION)/{{.OS}}-{{.Arch}}/{{.Dir}}" -osarch='$(TARGETS)' $(if $(TAGS),-tags '$(TAGS)',) -ldflags '$(LDFLAGS)' ${PKG}/cmd/${PROJECT_NAME}

.PHONY: release
release: build-cross
	cp Dockerfile $(DISTDIR)/bin/$(BINARY_VERSION)/linux-amd64
	docker build -t oxynozeta/kubernetes-tagger:$(BINARY_VERSION) $(DISTDIR)/bin/$(BINARY_VERSION)/linux-amd64

.PHONY: docker-latest
docker-latest:
	docker tag oxynozeta/kubernetes-tagger:$(BINARY_VERSION) oxynozeta/kubernetes-tagger:latest

.PHONY: docker-publish
docker-publish:
	docker push oxynozeta/kubernetes-tagger:$(BINARY_VERSION)

.PHONY: docker-publish-latest
docker-publish-latest:
	docker push oxynozeta/kubernetes-tagger:latest

test: dep ## Run unittests
	$(GO) test -short -cover -coverprofile=c.out ./...

.PHONY: coverage-report
coverage-report:
	$(GO) tool cover -html=c.out -o coverage.html
	$(GO) tool cover -func c.out

.PHONY: clean
clean:
	@rm -rf $(BINDIR) $(DISTDIR)

#############
# Bootstrap #
#############

HAS_GIT := $(shell command -v git;)
HAS_GOX := $(shell command -v gox;)
HAS_GOLANGCI_LINT := $(shell command -v golangci-lint;)
HAS_CURL:=$(shell command -v curl;)

.PHONY: dep
dep:
ifndef HAS_GOX
	go get -u github.com/mitchellh/gox
endif
ifndef HAS_GOLANGCI_LINT
	@echo "=> Installing golangci-lint tool"
ifndef HAS_CURL
	$(error You must install curl)
endif
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.43.0
endif
ifndef HAS_GIT
	$(error You must install Git)
endif
	$(GO) mod download all
	$(GO) mod tidy
