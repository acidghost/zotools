SHELL = /bin/bash
GO ?= go
BUILD_DIR := ./build
PREFIX := ~/.local
BIN_DIR := $(PREFIX)/bin/
NAME := zotools
VERSION := $(shell cat ./VERSION)
COMMIT := $(shell git rev-parse HEAD 2> /dev/null || true)

GO_SRC = $(shell find . -name '*.go')

COVERAGE_PATH ?= $(shell pwd)/.coverage
export COVERAGE_PATH
COVERAGE_FLAGS ?= -covermode=count -coverpkg=./...
$(shell mkdir -p ${COVERAGE_PATH})

all: check build

.PHONY: build
build: $(GO_SRC)
	$(GO) build \
		-buildmode=pie -o $(BUILD_DIR)/$(NAME) \
		-ldflags="-s -w -X main.version=$(VERSION)-$(COMMIT)" \
		./cmd/zotools

.PHONY: build.coverage
build.coverage: $(GO_SRC)
	$(GO) test \
		$(COVERAGE_FLAGS) \
		-tags coverage \
		-buildmode=pie -c -o $(BUILD_DIR)/$(NAME) \
		-ldflags="-s -w -X main.version=$(VERSION)-$(COMMIT)" \
		./cmd/zotools

.PHONY: check
check: $(GO_SRC)
	golangci-lint run

.PHONY: test
test: $(GO_SRC)
	$(GO) test ./...

.PHONY: test.coverage
test.coverage: $(GO_SRC)
	$(GO) test \
		$(COVERAGE_FLAGS) \
		-test.coverprofile=coverage.test.txt \
		-test.outputdir=$(COVERAGE_PATH) \
		./...

.PHONY: test-integration
test-integration:
	export PATH=./test/bats/bin:$$PATH; bats test/*.bats

.PHONY: test-integration.coverage
test-integration.coverage:
	export PATH=./test/bats/bin:$$PATH; export COVERAGE=1; bats test/*.bats

.PHONY: codecov
codecov:
	bash <(curl -s https://codecov.io/bash) -v -s $(COVERAGE_PATH) -f "coverage.*"

.PHONY: install
install:
	install -d -m 755 $(BIN_DIR)
	install -m 755 $(BUILD_DIR)/$(NAME) $(BIN_DIR)

.PHONY: uninstall
uninstall:
	rm $(BIN_DIR)/$(NAME)

.PHONY: banner.txt
banner.txt:
	echo -n 'text 0,0 "' > $@
	toilet -f larry3d $(NAME) | head -n -2 >> $@
	echo '"' >> $@

logo.png: banner.txt
	convert -size 360x100 xc:white -transparent white -font "FreeMono" \
		-pointsize 12 -fill green -draw @$< $@

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
