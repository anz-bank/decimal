.PHONY: all
all: test-all build-linux lint


.PHONY: test-all
test-all: test test-32

.PHONY: test
test:
	go test $(GOTESTFLAGS)
	go test $(GOTESTFLAGS) -tags=decimal_debug

.PHONY: test-32
test-32:
	$(DOCKERRUN) -e GOARCH=arm golang:1.23.0 go test $(GOTESTFLAGS)

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
	GOARCH=$(GOARCH.$*) go test -c -o $@ $(GOTESTFLAGS) .

.INTERMEDIATE: decimal.32.debug.test decimal.64.debug.test
decimal.%.debug.test:
	GOARCH=$(GOARCH.$*) go test -c -o $@ -tags=decimal_debug $(GOTESTFLAGS) .

DOCKERRUN = docker run --rm \
	-w /app \
	-v $(PWD):/app \
	-v `go env GOCACHE`:/root/.cache/go-build \
	-v `go env GOMODCACHE`:/go/pkg/mod

# Dependency on build-linux primes Go caches.
.PHONY: lint
lint: build-linux
	$(DOCKERRUN) golangci/golangci-lint:v1.60.1-alpine golangci-lint run

%.pprof: %.prof
	go tool pprof -http=:8080 $<

.INTERMEDIATE: %.prof
%.prof: $(wildcard *.go)
	go test -$*profile $@ -count=10 $(GOPROFILEFLAGS)

.PHONY: bench
bench: bench.txt
	cat $<

bench-stat: bench.stat
	cat $<

bench.stat: bench.txt
	[ -f bench.old ] || git show @:$< > bench.old || (rm -f $@; false)
	benchstat bench.old $< > $@ || (rm -f $@; false)

bench.txt: test
	go test -run=^$$ -bench=. -benchmem $(GOBENCHFLAGS) > $@ || (rm -f $@; false)
