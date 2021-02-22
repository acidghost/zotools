all: zotools

.PHONY: zotools
zotools:
	go build ./cmd/zotools

.PHONY: tests
tests:
	go test ./...

.PHONY: install
install:
	go install ./cmd/zotools

.PHONY: clean
clean:
	go clean
	$(RM) zotools
