package d64

import (
	"fmt"
	"math/bits"
	"math/rand"
	"testing"
)

const (
	base    = 10_000_000_000_000_000
	tenBase = 10 * base
	limit   = tenBase * tenBase
	limitHi = limit / (1 << 64)
	limitLo = limit % (1 << 64)
)

func TestU128_divrem_10_15(t *testing.T) {
	t.Parallel()
	testU128_divrem(t, u128_div_10_15, 1_000_000_000_000_000)
}

func TestU128_divrem_10_16(t *testing.T) {
	t.Parallel()
	testU128_divrem(t, u128_div_10_16, 10_000_000_000_000_000)
}

func testU128_divrem(t *testing.T, div func(hi, lo uint64) uint64, d uint64) {
	t.Helper()

	test := func(hi, lo uint64) func(t *testing.T) {
		return func(t *testing.T) {
			t.Helper()
			q := div(hi, lo)
			Q, _ := bits.Div64(hi, lo, d)
			if q != Q {
				t.Fatalf("U128_div_10_16(%016x_%016x) = %d (%016[3]x), want %d (%016[4]x)",
					hi, lo, q, Q)
			}
		}
	}
	t.Run("1", test(0, 0))
	t.Run("2", test(0, 1))
	t.Run("3", test(0, 10_000_000_000_000_000))
	t.Run("4", test(0, 20_000_000_000_000_000))
	t.Run("5", test(
		10_000_000_000_000_000_000_000/(1<<64),
		10_000_000_000_000_000_000_000%(1<<64),
	))
	t.Run("6", test(1, 20_000_000_000_000_000))

	r := rand.New(rand.NewSource(0))
	n := 1_000_000
	if testing.Short() {
		n = 100_000
	}
	for i := 0; i < n; i++ {
		// Ensure hi:lo < 10^32.
		hi := r.Uint64() % limitHi
		lo := r.Uint64() % limitLo
		if hi == limitHi {
			lo %= limitLo
		}
		t.Run(fmt.Sprintf("rand[%d]", i), test(hi, lo))
	}
}

var globalUint64 uint64

func BenchmarkDiv64_10_16(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		hi := uint64(i * 5421010862428)
		lo := uint64(i)
		globalUint64 = u128_div_10_16(hi, lo)
	}
}

func BenchmarkBitsDiv64_10_16(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		hi := uint64(i)
		lo := uint64(i)
		globalUint64, _ = bits.Div64(hi, lo, 10_000_000_000_000_000)
	}
}
