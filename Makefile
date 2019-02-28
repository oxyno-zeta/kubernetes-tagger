# All possible values : darwin/amd64 linux/amd64 windows/amd64 linux/386 linux/ppc64le linux/s390x linux/arm linux/arm64
TARGETS           ?= linux/amd64
PROJECT_NAME	  := kubernetes-tagger
PKG				  := github.com/oxyno-zeta/$(PROJECT_NAME)
PKG_LIST		  := $(shell go list ${PKG}/... | grep -v /vendor/)

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
	golint -set_exit_status ${PKG_LIST}

.PHONY: build
build: clean dep
	GOBIN=$(BINDIR) $(GO) install $(GOFLAGS) -tags '$(TAGS)' -ldflags '$(LDFLAGS)' $(PKG)/cmd/${PROJECT_NAME}

.PHONY: build-cross
build-cross: LDFLAGS += -extldflags "-static"
build-cross: clean dep
	CGO_ENABLED=0 gox -output="$(DISTDIR)/bin/$(BINARY_VERSION)/{{.OS}}-{{.Arch}}/{{.Dir}}" -osarch='$(TARGETS)' $(if $(TAGS),-tags '$(TAGS)',) -ldflags '$(LDFLAGS)' ${PKG}/cmd/${PROJECT_NAME}

.PHONY: release
release: build-cross

test: dep ## Run unittests
	$(GO) test -short ${PKG_LIST}

race: dep ## Run data race detector
	$(GO) test -race -short ${PKG_LIST}

.PHONY: clean
clean:
	@rm -rf $(BINDIR) $(DISTDIR)

#############
# Bootstrap #
#############

HAS_GIT := $(shell command -v git;)
HAS_GOX := $(shell command -v gox;)
HAS_GOLINT := $(shell command -v golint;)
HAS_GODEP := $(shell command -v dep;)

.PHONY: dep
dep:
ifndef HAS_GOX
	go get -u github.com/mitchellh/gox
endif
ifndef HAS_GOLINT
	go get -u golang.org/x/lint/golint
endif
ifndef HAS_GIT
	$(error You must install Git)
endif
ifndef HAS_GODEP
	go get -u github.com/golang/dep/cmd/dep
endif
	dep ensure
