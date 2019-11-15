package decimal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUint128Shl(t *testing.T) {
	require.Equal(t, uint128T{}, uint128T{}.shl(1))
	require.Equal(t, uint128T{2, 0}, uint128T{1, 0}.shl(1))
	require.Equal(t, uint128T{4, 0}, uint128T{2, 0}.shl(1))
	require.Equal(t, uint128T{4, 0}, uint128T{1, 0}.shl(2))
	require.Equal(t, uint128T{0, 1}, uint128T{1, 42}.shl(64))
	require.Equal(t, uint128T{0, 3}, uint128T{3, 42}.shl(64))
}

func TestUint128Sqrt(t *testing.T) {
	require.EqualValues(t, 0, uint128T{}.sqrt())
	require.EqualValues(t, 1, uint128T{1, 0}.sqrt())
	require.EqualValues(t, 1, uint128T{2, 0}.sqrt())
	require.EqualValues(t, 2, uint128T{4, 0}.sqrt())
	require.EqualValues(t, 2, uint128T{8, 0}.sqrt())
	require.EqualValues(t, 3, uint128T{9, 0}.sqrt())
	require.EqualValues(t, int64(1<<32), uint128T{0, 1}.sqrt())
	require.EqualValues(t, int64(2<<32), uint128T{0, 4}.sqrt())
}

func TestUint128DivBy10(t *testing.T) {
	require.Equal(t, uint128T{0, 0}, uint128T{0, 0}.divBy10())
	require.Equal(t, uint128T{0, 0}, uint128T{1, 0}.divBy10())
	require.Equal(t, uint128T{0, 0}, uint128T{9, 0}.divBy10())
	require.Equal(t, uint128T{1, 0}, uint128T{10, 0}.divBy10())
	require.Equal(t, uint128T{1, 0}, uint128T{19, 0}.divBy10())
	require.Equal(t, uint128T{2, 0}, uint128T{20, 0}.divBy10())
	require.Equal(t, uint128T{9, 0}, uint128T{99, 0}.divBy10())
	require.Equal(t, uint128T{10, 0}, uint128T{100, 0}.divBy10())
	require.Equal(t, uint128T{9999, 0}, uint128T{99999, 0}.divBy10())
	require.Equal(t, uint128T{10000, 0}, uint128T{100000, 0}.divBy10())
	require.Equal(t, uint128T{(1 << 64) / 10, 0}, uint128T{0, 1}.divBy10())
	require.Equal(t, uint128T{(2 << 64) / 10, 0}, uint128T{0, 2}.divBy10())
	require.Equal(t,
		uint128T{((123 << 64) / 10) % (1 << 64), ((123 << 64) / 10) >> 64},
		uint128T{0, 123}.divBy10(),
	)
	require.Equal(t,
		uint128T{((123456789 << 64) / 10) % (1 << 64), ((123456789 << 64) / 10) >> 64},
		uint128T{0, 123456789}.divBy10(),
	)
	require.Equal(t,
		uint128T{((123 << 56 << 64) / 10) % (1 << 64), ((123 << 56 << 64) / 10) >> 64},
		uint128T{0, 123 << 56}.divBy10(),
	)
	require.Equal(t,
		uint128T{((1<<128 - 1) / 10) % (1 << 64), ((1<<128 - 1) / 10) >> 64},
		uint128T{1<<64 - 1, 1<<64 - 1}.divBy10(),
	)
}
