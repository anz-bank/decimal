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
	GOOS=linux $(MAKE) build

# Prime Go caches for docker golangci-lint.
.PHONY: build
build: build-64-bit build-32-bit

.PHONY: build-32-bit build-64-bit
build-32-bit: decimal.32.release.test decimal.32.debug.test
build-64-bit: decimal.64.release.test decimal.64.debug.test

GOARCH.32=arm
GOARCH.64=

.INTERMEDIATE: decimal.32.release.test decimal.64.release.test
decimal.%.release.test:
	GOARCH=$(GOARCH.$*) go test -c -o $@ .

.INTERMEDIATE: decimal.32.debug.test decimal.64.debug.test
decimal.%.debug.test:
	GOARCH=$(GOARCH.$*) go test -c -o $@ -tags=decimal_debug .

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
profile: cpu.prof
	go tool pprof -http=:8080 $<

.INTERMEDIATE: cpu.prof
cpu.prof:
	go test -cpuprofile $@ -count=10

.PHONY: bench
bench: bench.txt
	cat $<

bench-stat: bench.stat
	cat $<

bench.stat: bench.txt
	[ -f bench.old ] || git show @:$< > bench.old || (rm -f $@; false)
	benchstat bench.old $< > $@ || (rm -f $@; false)

bench.txt:
	go test -run=^$$ -bench=. -benchmem -count=10 > $@ || (rm -f $@; false)
