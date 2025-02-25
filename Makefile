.PHONY: all
all: test-all build-linux lint no-allocs

.PHONY: ci
ci: test-all no-allocs

.PHONY: test-all
test-all: test test-32

.PHONY: test
test: test-release
	go test $(GOTESTFLAGS) -tags=decimal_debug ./d64

.PHONY: test-release
test-release:
	go test $(GOTESTFLAGS) ./d64

.PHONY: test-32
test-32:
	if [ "$(shell go env GOOS)" = "linux" ]; then \
		GOARCH=386 go test $(subst -race,,$(GOTESTFLAGS)) ./d64; \
	else \
		$(DOCKERRUN) -e GOARCH=arm golang:1.23.0 go test $(GOTESTFLAGS) ./d64; \
	fi

.PHONY: build-linux
build-linux:
	GOOS=linux $(MAKE) build

# Prime Go caches for docker golangci-lint.
.PHONY: build
build: build-64-bit build-32-bit

.PHONY: build-32-bit build-64-bit
build-32-bit: d64/decimal.32.release.test d64/decimal.32.debug.test
build-64-bit: d64/decimal.64.release.test d64/decimal.64.debug.test

GOARCH.32=arm
GOARCH.64=

.INTERMEDIATE: d64/decimal.32.debug.test d64/decimal.64.debug.test
d64/decimal.%.debug.test:
	GOARCH=$(GOARCH.$*) go test -c -o $@ -tags=decimal_debug $(GOTESTFLAGS) ./d64

.INTERMEDIATE: d64/decimal.32.release.test d64/decimal.64.release.test
d64/decimal.%.release.test:
	GOARCH=$(GOARCH.$*) go test -c -o $@ $(GOTESTFLAGS) ./d64

.PHONY: clean
clean:
	rm -f decimal.*.release.test decimal.*.debug.test

DOCKERRUN = docker run --rm \
	-w /app \
	-v $(PWD):/app \
	-v `go env GOCACHE`:/root/.cache/go-build \
	-v `go env GOMODCACHE`:/go/pkg/mod

# Dependency on build-linux primes Go caches.
.PHONY: lint
lint: build-linux
	$(DOCKERRUN) golangci/golangci-lint:v1.64.5-alpine golangci-lint run

%.pprof: %.prof
	go tool pprof -http=:8080 $<

.INTERMEDIATE: %.prof
%.prof: $(wildcard *.go)
	go test -$*profile $@ $(GOPROFILEFLAGS)

.PHONY: bench
bench: d64/bench.txt
	cat $<

.PHONY: bench-stat
bench-stat: d64/bench.stat
	cat $<

d64/bench.stat: d64/bench.txt
	[ -f d64/bench.old ] || git show @:$< > d64/bench.old || (rm -f $@; false)
	benchstat d64/bench.old $< > $@ || (rm -f $@; false)

d64/bench.txt: test
	go test -run=^$$ -bench=. -benchmem $(GOBENCHFLAGS) ./d64 | tee $@ || (rm -f $@; false)

NOALLOC = \
	BenchmarkIODecimal64String2 \
	BenchmarkIODecimal64Append \
	BenchmarkDecimal64Abs \
	BenchmarkDecimal64Add \
	BenchmarkDecimal64Cmp \
	BenchmarkDecimal64Mul \
	BenchmarkFloat64Mul \
	BenchmarkDecimal64Quo \
	BenchmarkDecimal64Sqrt \
	BenchmarkDecimal64Sub

no-allocs:
	allocs=$$( \
		go test -run=^$$ -bench="^($$(echo $(NOALLOC) | sed 's/ /|/g'))$$" -benchmem $(GOBENCHFLAGS) ./d64 | \
			awk '/^Benchmark/ {if ($$7 != "0") print}' \
	); \
	if [ -n "$$allocs" ]; then \
		echo "** alloc regression **"; \
		echo "$$allocs"; \
		false; \
	fi
