#!/usr/bin/make -f

# Go version and build settings
GO_VERSION := 1.22
GO_SYSTEM_VERSION := $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f1-2)
REQUIRE_GO_VERSION := $(GO_VERSION)
ENABLED_FLAGS := minitia_launch

# Project version
WEAVE_VERSION := v0.0.1

# Build directory
BUILDDIR ?= $(CURDIR)/build

# Build targets
BUILD_TARGETS := build install test

release_version=$(filter-out $@,$(MAKECMDGOALS))

# Version check
check_version:
	@if [ "$(GO_SYSTEM_VERSION)" != "$(REQUIRE_GO_VERSION)" ]; then \
		echo "ERROR: Go version $(REQUIRE_GO_VERSION) is required for Weave."; \
		exit 1; \
	fi

# Build settings
LDFLAGS := -X github.com/initia-labs/weave/cmd.Version=$(WEAVE_VERSION) \
           -X github.com/initia-labs/weave/flags.EnabledFlags=$(ENABLED_FLAGS)

# Build targets
build: check_version $(BUILDDIR)
	go build -mod=readonly -ldflags "$(LDFLAGS)" -o $(BUILDDIR)/weave .

install: check_version
	go install -ldflags "$(LDFLAGS)" .

test: check_version
	go clean -testcache
	go test -v ./...

# Release process
release:
	@if [ -z "$(release_version)" ]; then \
		echo "ERROR: You must provide a release version. Example: make release v0.0.15"; \
		exit 1; \
	fi
	git tag -a $(release_version) -m "$(release_version)"
	git push origin $(release_version)
	gh release create $(release_version) --title "$(release_version)" --notes "Release notes for version $(release_version)"

# Catch-all target
%:
	@:
