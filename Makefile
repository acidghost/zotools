GO ?= go
BUILD_DIR := ./build
PREFIX := ~/.local
BIN_DIR := $(PREFIX)/bin/
NAME := zotools
VERSION := $(shell cat ./VERSION)
COMMIT := $(shell git rev-parse HEAD 2> /dev/null || true)

GO_SRC = $(shell find . -name '*.go')

all: check build

.PHONY: build
build: $(GO_SRC)
	$(GO) build \
		-buildmode=pie -o $(BUILD_DIR)/$(NAME) \
		-ldflags="-s -w -X main.version=${VERSION}-${COMMIT}" \
		./cmd/zotools

.PHONY: check
check: $(GO_SRC)
	golangci-lint run

.PHONY: tests
tests: $(GO_SRC)
	$(GO) test ./...

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
	$(RM) $(BUILD_DIR)
