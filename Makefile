all: zotools

.PHONY: zotools tests

zotools:
	go build ./cmd/$@

tests:
	go test ./...

clean:
	$(RM) zotools
