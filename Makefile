all: zotools

.PHONY: zotools

zotools:
	go build ./cmd/$@

clean:
	$(RM) zotools
