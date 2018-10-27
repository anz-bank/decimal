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
	require.EqualValues(t, 1<<32, uint128T{0, 1}.sqrt())
	require.EqualValues(t, 2<<32, uint128T{0, 4}.sqrt())
}
