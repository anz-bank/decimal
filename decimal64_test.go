package decimal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew64FromInt64(t *testing.T) {
	for i := int64(-1000); i <= 1000; i++ {
		d := NewDecimal64FromInt64(i)
		j := d.Int64()
		require.EqualValues(t, i, j)
	}
}

func TestNew64FromInt64Big(t *testing.T) {
	const limit = decimal64Base
	const step = limit / 997
	for i := -int64(limit); i <= limit; i += step {
		d := NewDecimal64FromInt64(i)
		j := d.Int64()
		require.EqualValues(t, i, j)
	}
}

func TestDecimal64Float64(t *testing.T) {
	r := require.New(t)
	r.Equal(NegOne64.Float64(), -1.0)
	r.Equal(Zero64.Float64(), 0.0)
	r.Equal(NegZero64.Float64(), -0.0)
	r.Equal(One64.Float64(), 1.0)
	r.Equal(NewDecimal64FromInt64(10).Float64(), 10.0)

	oneThird := One64.Quo(NewDecimal64FromInt64(3))
	one := oneThird.Add(oneThird).Add(oneThird)
	r.InEpsilon(oneThird.Float64(), 1.0/3.0, 0.00000001)
	r.Equal(one.Float64(), 1.0)
}

func TestDecimal64SNaN(t *testing.T) {
	require.Panics(t, func() {
		SNaN64.Add(Zero64)
	})
}

func TestDecimal64IsInf(t *testing.T) {
	require.True(t, Infinity64.IsInf())
	require.True(t, NegInfinity64.IsInf())

	require.False(t, Zero64.IsInf())
	require.False(t, NegZero64.IsInf())
	require.False(t, QNaN64.IsInf())
	require.False(t, SNaN64.IsInf())
	require.False(t, NewDecimal64FromInt64(42).IsInf())
	require.False(t, NewDecimal64FromInt64(-42).IsInf())
}

func TestDecimal64IsNaN(t *testing.T) {
	require.True(t, QNaN64.IsNaN())
	require.True(t, SNaN64.IsNaN())

	require.True(t, QNaN64.IsQNaN())
	require.False(t, SNaN64.IsQNaN())

	require.False(t, QNaN64.IsSNaN())
	require.True(t, SNaN64.IsSNaN())

	for _, n := range []Decimal64{
		Infinity64,
		NegInfinity64,
		Zero64,
		NegZero64,
		NewDecimal64FromInt64(42),
		NewDecimal64FromInt64(-42),
	} {
		require.False(t, n.IsNaN(), "%v", n)
		require.False(t, n.IsQNaN(), "%v", n)
		require.False(t, n.IsSNaN(), "%v", n)
	}
}
