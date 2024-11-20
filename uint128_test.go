package decimal

import "testing"

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
