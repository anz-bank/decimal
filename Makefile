.PHONY: all
all: test test-debug build-linux lint

.PHONY: test
test:
	go test

.PHONY: test-debug
test-debug:
	go test -tags=decimal_debug

.PHONY: build-linux
build-linux:
	GOOS=linux $(MAKE) build build-32bit

.PHONY: build
build-64bit:
	go test -c -o decimal.$@.test . && rm -f decimal.$@.test
	go test -c -o decimal.$@.debug.test -tags=decimal_debug . && rm -f decimal.$@.debug.test

.PHONY: build-32bit
build-32bit:
	GOARCH=arm go test -c -o decimal.$@.test . && rm -f decimal.$@.test
	GOARCH=arm go test -c -o decimal.$@.debug.test -tags=decimal_debug . && rm -f decimal.$@.debug.test

# Dependency on build-linux primes Go caches.
.PHONY: lint
lint: build-linux
	docker run --rm \
		-w /app \
		-v $(PWD):/app \
		-v `go env GOCACHE`:/root/.cache/go-build \
		-v `go env GOMODCACHE`:/go/pkg/mod \
		golangci/golangci-lint:v1.60.1-alpine \
		golangci-lint run

.PHONY: profile
profile:
	go test -cpuprofile cpu.prof -count=10 && go tool pprof -http=:8080 cpu.prof
