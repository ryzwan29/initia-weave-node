#!/usr/bin/make -f

GO_VERSION=1.22
GO_SYSTEM_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f1-2)
REQUIRE_GO_VERSION = $(GO_VERSION)

WEAVE_VERSION=v0.0.1

BUILDDIR ?= $(CURDIR)/build
BUILD_TARGETS = build

release=0.0.15

check_version:
ifneq ($(GO_SYSTEM_VERSION), $(REQUIRE_GO_VERSION))
	@echo "ERROR: Go version ${REQUIRE_GO_VERSION} is required for Weave."
	exit 1
endif


build: BUILD_ARGS=-o $(BUILDDIR)/

$(BUILD_TARGETS): check_version go.sum $(BUILDDIR)/
ifeq ($(OS),Windows_NT)
	exit 1
else
	go $@ -mod=readonly -ldflags "-X github.com/initia-labs/weave/cmd.Version=$(WEAVE_VERSION)" .
endif

$(BUILDDIR)/:
	mkdir -p $(BUILDDIR)/

install:
	go install -ldflags "-X github.com/initia-labs/weave/cmd.Version=$(WEAVE_VERSION)" .

test: build
	clear
	./weave

release:
	git tag -a v$(release) -m "Release version $(release)"
	git push origin v$(release)
	gh release create v$(release) --title "Release v$(release)" --notes "Release notes for version $(release)"
