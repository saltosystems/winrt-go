SHELL       ?= /bin/bash
NAME        =  $(shell echo $(PACKAGE) | rev | cut -d/ -f1 | rev)
PLATFORM    ?= linux darwin windows
DOCKER      = $(shell command -v docker 2>/dev/null)
REPORTS_DIR ?= .reports
PREFIX      ?= gcr.io/salto-containers
BUILD_TAGS  ?=
SANITY_CHECKING_CONFIG ?= .go-builder/.golangci.yml

# Overridable by CI

COMMIT_SHORT     ?= $(shell git rev-parse --verify --short HEAD)
VERSION          ?= v0.0.0-sha.$(COMMIT_SHORT)
VERSION_NOPREFIX ?= $(shell echo $(VERSION) | sed -e 's/^[[v]]*//')

#
# Common methodology based targets
#

prepare_targets      = mod-download
sanity_check_targets = check-sync golangci-lint mod-verify mod-check
build_targets        = go-build
test_targets         = 
release_targets      = 
clean_targets        = app-clean

# If there's a build container then also run the docker build and release
ifneq ("$(wildcard build/container/Dockerfile)","")
build_targets       += docker-build
release_targets     += docker-push
endif

#
# Global checks
#

ifndef PACKAGE
$(error You must define the mandatory PACKAGE variable)
endif

ifndef PKG
$(error You must define the mandatory PKG variable)
endif

ifndef APP
$(error You must define the mandatory APP variable)
endif

#
# Custom project related targets
#

.PHONY: mod-download
mod-download:
	@echo "Running mod download..."
	@go mod download

.PHONY: golangci-lint
golangci-lint:
	@echo "Running golangci-lint..."
	@golangci-lint run $(PKG) --timeout 30m -v --config=$(SANITY_CHECKING_CONFIG)


# Another modules verification apart from mod-verify step should be to ensure that go mod tidy has been run on dev box when this gets fixed:
# https://github.com/golang/go/issues/27005
.PHONY: mod-verify
mod-verify:
	@echo "Running mod verify..."
	@go mod verify

# Replace this temporary "hack" with `go mod check` feature once implemented:
# https://github.com/golang/go/issues/27005
.PHONY: mod-check
mod-check:
	@echo "Running mod check..."
	@go mod tidy && git diff --exit-code -- 'go.mod' 'go.sum' > /dev/null \
	|| (echo "mod-check failed, run \"go mod tidy\", commit the \"go.mod\" and/or \"go.sum\" files and try again"; exit 1)

.PHONY: go-build
go-build:
	@for app in $(APP) ; do \
		for os in $(PLATFORM) ; do \
			ext=""; \
			if [ "$$os" == "windows" ]; then \
				ext=".exe"; \
			fi; \
			GOOS=$$os GOARCH=amd64 CGO_ENABLED=0 \
			go build \
				-a -x -tags "$(BUILD_TAGS)" -installsuffix cgo -installsuffix netgo \
				-ldflags " \
					-X main.Version=$(VERSION_NOPREFIX) \
					-X main.GitRev=$(COMMIT_SHORT) \
				" \
				-o ./bin/$$app-$(VERSION_NOPREFIX)-$$os-amd64$$ext \
				./cmd/$$app; \
		done; \
	done

.PHONY: docker-build
docker-build:
	@echo "Building docker image..."
	@for app in $(APP) ; do \
		cp bin/$$app-$(VERSION_NOPREFIX)-linux-amd64 build/container/$$app-linux-amd64; \
		chmod 0755 build/container/$$app-linux-amd64; \
	done; \
	"$(DOCKER)" build \
		-f build/container/Dockerfile \
		-t $(PREFIX)/$(NAME):$(VERSION) \
	build/container/

.PHONY: docker-push
docker-push:
	@echo "Pushing docker image..."
	"$(DOCKER)" push $(PREFIX)/$(NAME):$(VERSION); \

.PHONY: app-clean
app-clean:
	rm -f ./bin/*
	@for app in $(APP) ; do \
		rm -f cmd/$$app/*-linux-amd64; \
	done

.PHONY: check-sync
check-sync:
	@echo "Checking if everything is in sync..."
	@.go-builder/scripts/check-sync.sh

.PHONY: sync-local
sync-local:
	@echo "Synchronizing..."
	@mv .go-builder .go-builder-tmp
	@if docker run -v $(shell pwd):/workspace --rm -it $(shell docker build -q -f Dockerfile.build . | head -n 1) cp -r /etc/go-builder /workspace/.go-builder; then \
		rm -r .go-builder-tmp; \
	else \
		mv .go-builder-tmp .go-builder; \
	fi

#
# Debug any makefile variable
# Usage: print-<VAR>
#
print-%  : ; @echo $* = $($*)
