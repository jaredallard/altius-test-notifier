# go option
GO             ?= go
SHELL          := /usr/bin/env bash
GOOS           := $(shell go env GOOS)
GOARCH         := $(shell go env GOARCH)
PKG            := $(GO) mod download -x
# TODO(jaredallard): infer from Git tag
APP_VERSION    := $(shell git describe --match 'v[0-9]*' --tags --abbrev=0 --always HEAD)
LDFLAGS        := -w -s
GOFLAGS        :=
GOPROXY        := https://proxy.golang.org
GO_EXTRA_FLAGS := -v
TAGS           :=
BINDIR         := $(CURDIR)/bin
PKGDIR         := github.com/jaredallard/altius-test-notifier
CGO_ENABLED    := 1
BENCH_FLAGS    := "-bench=Bench $(BENCH_FLAGS)"
TEST_TAGS      ?= tm_test
LOG            := "$(CURDIR)/scripts/make-log-wrapper.sh"

.PHONY: default
default: build

.PHONY: version
version:
	@echo $(APP_VERSION)

.PHONY: release
release:
	@$(LOG) info "Creating a release"
	./scripts/gobin.sh github.com/goreleaser/goreleaser

.PHONY: pre-commit
pre-commit: fmt

.PHONY: build
build: gogenerate gobuild

.PHONY: test
test:
	GOPROXY=$(GOPROXY) ./scripts/test.sh

.PHONY: docs
docs:
	./scripts/docs.sh

.PHONY: dep
dep:
	@$(LOG) info "Installing dependencies using '$(PKG)'"
	GOPROXY=$(GOPROXY) $(PKG)

.PHONY: gogenerate
gogenerate:
	@$(LOG) info "Running gogenerate"
	GOPROXY=$(GOPROXY) $(GO) generate ./...

.PHONY: gobuild
gobuild:
	@$(LOG) info "Building releases into ./bin"
	mkdir -p $(BINDIR)
	GOPROXY=$(GOPROXY) CGO_ENABLED=$(CGO_ENABLED) $(GO) build -o $(BINDIR)/ -ldflags "$(LDFLAGS)" $(GO_EXTRA_FLAGS) $(PKGDIR)/...

.PHONY: docker-build
docker-build:
	@$(LOG) info "Building docker image"
	DOCKER_BUILDKIT=1 docker build -t "jaredallard/localizer:latest" .

.PHONY: fmt
fmt:
	@$(LOG) info "Running goimports"
	find  . -path ./vendor -prune -o -type f -name '*.go' -print | xargs ./scripts/gobin.sh golang.org/x/tools/cmd/goimports -w
	@$(LOG) info "Running shfmt"
	./scripts/gobin.sh mvdan.cc/sh/v3/cmd/shfmt -l -w -s .
