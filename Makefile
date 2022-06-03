PACKAGE     =  github.com/saltosystems/winrt-go
PKG         ?= ./...
APP         ?= winrt-go-gen
BUILD_TAGS  ?= 

include .go-builder/Makefile

.PHONY: prepare
prepare: $(prepare_targets)

.PHONY: sanity-check
sanity-check: $(sanity_check_targets)

.PHONY: build
build: $(build_targets)

.PHONY: test
test: $(test_targets)

.PHONY: release
release: $(release_targets)

.PHONY: clean
clean: $(clean_targets)
