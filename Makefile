.PHONY: all
all: test test-debug

.PHONY: test
test:
	go test

.PHONY: test-debug
test-debug:
	go test -tags=decimal_debug
