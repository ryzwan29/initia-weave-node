#!/usr/bin/make -f

GO_VERSION=1.22
GO_SYSTEM_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f1-2)
REQUIRE_GO_VERSION = $(GO_VERSION)

BUILDDIR ?= $(CURDIR)/build
BUILD_TARGETS = build

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
	go $@ -mod=readonly .
endif

$(BUILDDIR)/:
	mkdir -p $(BUILDDIR)/

test: build
	clear
	./weave
