package decimal

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew64FromInt64(t *testing.T) {
	for i := int64(-1000); i <= 1000; i++ {
		d := NewDecimal64FromInt64(i)
		j := d.Int64()
		require.EqualValues(t, i, j, "%d", i)
	}

	// Test the neighborhood of powers of two up to the high-significand
	// representation threshold.
	for e := 4; e < 54; e++ {
		base := int64(1) << uint(e)
		for i := base - 10; i <= base+10; i++ {
			d := NewDecimal64FromInt64(i)
			j := d.Int64()
			require.EqualValues(t, i, j, "1<<%d + %d", e, i)
		}
	}
}

func TestNew64FromInt64Big(t *testing.T) {
	const limit = decimal64Base
	const step = limit / 997
	for i := -int64(limit); i <= limit; i += step {
		d := NewDecimal64FromInt64(i)
		j := d.Int64()
		require.EqualValues(t, i, j, "%d", i)
	}
}

func TestDecimal64Float64(t *testing.T) {
	require := require.New(t)

	require.Equal(-1.0, NegOne64.Float64())
	require.Equal(0.0, Zero64.Float64())
	require.Equal(-0.0, NegZero64.Float64())
	require.Equal(1.0, One64.Float64())
	require.Equal(10.0, NewDecimal64FromInt64(10).Float64())

	oneThird := One64.Quo(NewDecimal64FromInt64(3))
	one := oneThird.Add(oneThird).Add(oneThird)
	require.InEpsilon(oneThird.Float64(), 1.0/3.0, 0.00000001)
	require.InEpsilon(1.0, one.Float64(), 0.00000001)

	require.True(math.IsNaN(QNaN64.Float64()))
	require.Panics(func() { SNaN64.Float64() })
	require.Equal(math.Inf(1), Infinity64.Float64())
	require.Equal(math.Inf(-1), NegInfinity64.Float64())
}

func TestDecimal64Int64(t *testing.T) {
	require := require.New(t)

	require.EqualValues(-1, NegOne64.Int64())
	require.EqualValues(0, Zero64.Int64())
	require.EqualValues(-0, NegZero64.Int64())
	require.EqualValues(1, One64.Int64())
	require.EqualValues(10, NewDecimal64FromInt64(10).Int64())

	require.EqualValues(0, QNaN64.Int64())

	require.EqualValues(math.MaxInt64, Infinity64.Int64())
	require.EqualValues(math.MinInt64, NegInfinity64.Int64())

	googol, err := ParseDecimal64("1e100")
	require.NoError(err)
	require.EqualValues(math.MaxInt64, googol.Int64())

	long, err := ParseDecimal64("91234567890123456789e20")
	require.NoError(err)
	require.EqualValues(math.MaxInt64, long.Int64())
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

	a := Infinity64.getParts()
	require.True(t, a.isInf())

	b := NegInfinity64.getParts()
	require.True(t, b.isInf())
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

func TestDecimal64IsInt(t *testing.T) {
	require := require.New(t)

	fortyTwo := NewDecimal64FromInt64(42)

	require.True(Zero64.IsInt())
	require.True(fortyTwo.IsInt())
	require.True(fortyTwo.Mul(fortyTwo).IsInt())
	require.True(fortyTwo.Quo(fortyTwo).IsInt())
	require.False(One64.Quo(fortyTwo).IsInt())

	require.False(Infinity64.IsInt())
	require.False(NegInfinity64.IsInt())
	require.False(QNaN64.IsInt())
	require.False(SNaN64.IsInt())
}

func TestDecimal64Sign(t *testing.T) {
	require := require.New(t)

	require.Equal(0, Zero64.Sign())
	require.Equal(0, NegZero64.Sign())
	require.Equal(1, One64.Sign())
	require.Equal(-1, NegOne64.Sign())
}

func TestDecimal64Signbit(t *testing.T) {
	require := require.New(t)

	require.Equal(false, Zero64.Signbit())
	require.Equal(true, NegZero64.Signbit())
	require.Equal(false, One64.Signbit())
	require.Equal(true, NegOne64.Signbit())
}

func TestDecimal64isZero(t *testing.T) {
	require := require.New(t)

	require.Equal(true, Zero64.IsZero())
	require.Equal(true, Decimal64{Zero64.bits | neg64}.IsZero())
	require.Equal(false, One64.IsZero())
}

func TestNumDecimalDigits(t *testing.T) {
	require := require.New(t)
	for i, num := range powersOf10 {
		for j := uint64(1); j < 10 && i < 19; j++ {
			require.Equal(i+1, numDecimalDigits(num*j))
		}
	}
}

func TestIsNaN(t *testing.T) {
	require := require.New(t)
	a := Zero64.getParts()
	require.Equal(false, a.isNaN())

	b := SNaN64.getParts()
	require.Equal(true, b.isSNaN())

	c := QNaN64.getParts()
	require.Equal(true, c.isQNaN())
}

func TestIsSubnormal(t *testing.T) {
	require := require.New(t)

	require.Equal(true, MustParseDecimal64("0.1E-383").IsSubnormal())
	require.Equal(true, MustParseDecimal64("-0.1E-383").IsSubnormal())
	require.Equal(false, MustParseDecimal64("NaN10").IsSubnormal())
	subnormal64Parts := MustParseDecimal64("0.1E-383").getParts()
	require.Equal(true, subnormal64Parts.isSubnormal())

	require.Equal(false, NewDecimal64FromInt64(42).IsSubnormal())
	fortyTwoParts := NewDecimal64FromInt64(42).getParts()
	require.Equal(false, fortyTwoParts.isSubnormal())
}