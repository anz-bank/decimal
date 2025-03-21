package d64

import (
	"math/bits"
	"math/rand"
	"testing"
)

func TestUint128Shl(t *testing.T) {
	t.Parallel()

	test := func(expected, original uint128T, shift uint) {
		t.Helper()
		original.shl(&original, shift)
		equal(t, expected, original)
	}
	test(uint128T{}, uint128T{}, 1)
	test(uint128T{2, 0}, uint128T{1, 0}, 1)
	test(uint128T{4, 0}, uint128T{2, 0}, 1)
	test(uint128T{4, 0}, uint128T{1, 0}, 2)
	test(uint128T{0, 1}, uint128T{1, 42}, 64)
	test(uint128T{0, 3}, uint128T{3, 42}, 64)
}

func TestNumDecimalDigits(t *testing.T) {
	t.Parallel()

	for i, num := range tenToThe {
		for j := uint64(1); j < 10 && i < 19; j++ {
			equal(t, i+1, int(numDecimalDigitsU64(num*j)))
		}
	}
}

func TestSqrtu64(t *testing.T) {
	t.Parallel()

	test := func(n uint64) {
		t.Helper()
		replayOnFail(t, func() {
			t.Helper()
			s := sqrtu64(n)
			sq := s * s
			s1q := (s + 1) * (s + 1)
			// s1q < sq handles overflow cases.
			check(t, sq <= n && (s1q < sq || s1q >= n)).Or(t.FailNow)
		})
	}

	// Fibonacci numbers, just because
	var a, b uint64 = 1, 1
	for a < b {
		test(a)
		a, b = b, a+b
	}

	hi := uint64(1<<64 - 1)

	for i := 0; i < 64; i++ {
		n := uint64(1) << i
		test(n)
		test(hi - n)
	}

	for i := uint64(0); i < 100_000; i++ {
		test(i)
		test(hi - i)
	}

	// (2**64)**1e-6 ~ 1.00004
	for i := uint64(100_000); i < (1<<64)*99_999/100_004; {
		test(i)
		test(hi - i)
		a, b := bits.Mul64(i, 100_004)
		i, _ = bits.Div64(a, b, 100_000)
	}

	s := rand.NewSource(0).(rand.Source64)
	for i := 0; i < 1_000_000; i++ {
		test(s.Uint64())
	}
}

func TestSqrtu16(t *testing.T) {
	t.Parallel()

	test := func(n uint16) {
		t.Helper()
		replayOnFail(t, func() {
			t.Helper()
			s := sqrtu16(n) >> 8
			sq := uint32(s) * uint32(s)
			s1q := uint32(s+1) * uint32(s+1)
			// s1q < sq handles overflow cases.
			check(t, sq <= uint32(n) && (s1q < sq || s1q >= uint32(n))).Or(t.FailNow)
		})
	}

	for i := uint16(0); i < 1<<16-1; i++ {
		test(i)
	}
	test(1<<16 - 1)
	// fmt.Printf("%#v\n", sqrtTable)
	// t.Fail()
}
