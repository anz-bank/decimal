package decimal

import "testing"

func TestUint128Shl(t *testing.T) {
	t.Parallel()

	test := func(expected, original uint128T, shift uint) {
		t.Helper()
		actual := original.shl(shift)
		equal(t, expected, actual)
	}
	test(uint128T{}, uint128T{}, 1)
	test(uint128T{2, 0}, uint128T{1, 0}, 1)
	test(uint128T{4, 0}, uint128T{2, 0}, 1)
	test(uint128T{4, 0}, uint128T{1, 0}, 2)
	test(uint128T{0, 1}, uint128T{1, 42}, 64)
	test(uint128T{0, 3}, uint128T{3, 42}, 64)
}

func TestUint128Shr(t *testing.T) {
	t.Parallel()

	test := func(expected, original uint128T, shift uint) {
		t.Helper()
		actual := original.shr(shift)
		equal(t, expected, actual)
	}
	test(uint128T{}, uint128T{}, 1)
	test(uint128T{1, 0}, uint128T{2, 0}, 1)
	test(uint128T{2, 0}, uint128T{4, 0}, 1)
	test(uint128T{1, 0}, uint128T{4, 0}, 2)
	test(uint128T{1, 0}, uint128T{0, 1}, 64)
	test(uint128T{3, 0}, uint128T{0, 3}, 64)
	test(uint128T{0x80000000 << 32, 0}, uint128T{0, 1}, 1)
	test(uint128T{0xffff8000 << 32, 0x7fff}, uint128T{0xffff0000 << 32, 0xffff}, 1)
}

func TestUint128Sqrt(t *testing.T) {
	t.Parallel()

	test := func(expected uint64, original uint128T) {
		t.Helper()
		equal(t, expected, original.sqrt())
	}
	test(0, uint128T{})
	test(1, uint128T{1, 0})
	test(1, uint128T{2, 0})
	test(2, uint128T{4, 0})
	test(2, uint128T{8, 0})
	test(3, uint128T{9, 0})
	test(uint64(1<<32), uint128T{0, 1})
	test(uint64(2<<32), uint128T{0, 4})
}

func TestUint128MulBy10(t *testing.T) {
	t.Parallel()

	test := func(expected, original uint128T) {
		t.Helper()
		actual := original.mulBy10()
		equal(t, expected, actual)
	}
	test(uint128T{0, 0}, uint128T{0, 0})
	test(uint128T{10, 0}, uint128T{1, 0})
	test(uint128T{20, 0}, uint128T{2, 0})
	test(uint128T{90, 0}, uint128T{9, 0})
	test(uint128T{100, 0}, uint128T{10, 0})
	test(uint128T{99990, 0}, uint128T{9999, 0})
	test(uint128T{100000, 0}, uint128T{10000, 0})
	test(uint128T{0xfffffffffffffffa, 0}, uint128T{(1 << 64) / 10, 0})
	test(uint128T{0xfffffffffffffffc, 3}, uint128T{(4 << 64) / 10, 0})
}

func TestUint128DivBy10(t *testing.T) {
	t.Parallel()

	test := func(expected, original uint128T) {
		t.Helper()
		actual := original.divBy10()
		equal(t, expected, actual)
	}
	test(uint128T{0, 0}, uint128T{0, 0})
	test(uint128T{0, 0}, uint128T{1, 0})
	test(uint128T{0, 0}, uint128T{9, 0})
	test(uint128T{1, 0}, uint128T{10, 0})
	test(uint128T{1, 0}, uint128T{19, 0})
	test(uint128T{2, 0}, uint128T{20, 0})
	test(uint128T{9, 0}, uint128T{99, 0})
	test(uint128T{10, 0}, uint128T{100, 0})
	test(uint128T{9999, 0}, uint128T{99999, 0})
	test(uint128T{10000, 0}, uint128T{100000, 0})
	test(uint128T{(1 << 64) / 10, 0}, uint128T{0, 1})
	test(uint128T{(2 << 64) / 10, 0}, uint128T{0, 2})
	test(uint128T{((123 << 64) / 10) % (1 << 64), ((123 << 64) / 10) >> 64},
		uint128T{0, 123},
	)
	test(uint128T{((123456789 << 64) / 10) % (1 << 64), ((123456789 << 64) / 10) >> 64},
		uint128T{0, 123456789},
	)
	test(uint128T{((123 << 56 << 64) / 10) % (1 << 64), ((123 << 56 << 64) / 10) >> 64},
		uint128T{0, 123 << 56},
	)
	test(uint128T{((1<<128 - 1) / 10) % (1 << 64), ((1<<128 - 1) / 10) >> 64},
		uint128T{1<<64 - 1, 1<<64 - 1},
	)
}
