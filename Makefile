.PHONY: all
all: test test-debug build-32bit

.PHONY: test
test:
	go test

.PHONY: test-debug
test-debug:
	go test -tags=decimal_debug

.PHONY: build-32bit
build-32bit:
	GOOS=linux GOARCH=arm go test -c -o . && rm -f decimal.test
	GOOS=linux GOARCH=arm go test -c -o -tags=decimal_debug . && rm -f decimal.test
