package decimal

import (
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew64FromInt64(t *testing.T) {
	for i := int64(-1000); i <= 1000; i++ {
		d := New64FromInt64(i)
		j := d.Int64()
		require.EqualValues(t, i, j, "%d", i)
	}

	// Test the neighborhood of powers of two up to the high-significand
	// representation threshold.
	for e := 4; e < 54; e++ {
		base := int64(1) << uint(e)
		for i := base - 10; i <= base+10; i++ {
			d := New64FromInt64(i)
			j := d.Int64()
			require.EqualValues(t, i, j, "1<<%d + %d", e, i)
		}
	}
}

func TestNew64FromInt64Big(t *testing.T) {
	const limit = int64(decimal64Base)
	const step = limit / 997
	for i := -int64(limit); i <= limit; i += step {
		d := New64FromInt64(i)
		j := d.Int64()
		require.EqualValues(t, i, j, "%d", i)
	}
}

func TestDecimal64Parse(t *testing.T) {
	t.Parallel()

	test := func(expected string, source string) {
		t.Helper()
		assert.Equal(t, strings.TrimSpace(expected), MustParse64(source).String())
	}

	test("0", "0")
	test("1e-13", "0.0000000000001")
	test("1e-13", "1e-13")
	test("1", "1")
	test("100000", "100000")
	test("1e+6", "1000000")
}

func TestDecimal64ParseHalfEvenOdd(t *testing.T) {
	t.Parallel()

	ctx := Context64{Rounding: HalfEven}
	test := func(expected string, source string) {
		t.Helper()
		assert.Equal(t, strings.TrimSpace(expected), ctx.MustParse(source).String())
	}

	test("1.000000000000007    ", "1.000000000000007")
	test("1.000000000000007    ", "1.0000000000000074999999")
	test("1.000000000000008    ", "1.0000000000000075000000")
	test("1.000000000000008    ", "1.0000000000000075000001")

	test("1.000000000000007e+11", "100000000000.0007")
	test("1.000000000000007e+11", "100000000000.00074999999")
	test("1.000000000000008e+11", "100000000000.00075000000")
	test("1.000000000000008e+11", "100000000000.00075000001")
}

func TestDecimal64ParseHalfEvenEven(t *testing.T) {
	t.Parallel()

	ctx := Context64{Rounding: HalfEven}
	test := func(expected string, source string) {
		t.Helper()
		assert.Equal(t, strings.TrimSpace(expected), ctx.MustParse(source).String())
	}

	test("1.000000000000008    ", "1.000000000000008")
	test("1.000000000000008    ", "1.0000000000000084999999")
	test("1.000000000000008    ", "1.0000000000000085000000")
	test("1.000000000000009    ", "1.0000000000000085000001")

	test("1.000000000000008e+11", "100000000000.0008")
	test("1.000000000000008e+11", "100000000000.00084999999")
	test("1.000000000000008e+11", "100000000000.00085000000")
	test("1.000000000000009e+11", "100000000000.00085000001")
}

func TestDecimal64ParseHalfUp(t *testing.T) {
	t.Parallel()

	ctx := Context64{Rounding: HalfUp}
	test := func(expected string, source string) {
		t.Helper()
		assert.Equal(t, strings.TrimSpace(expected), ctx.MustParse(source).String())
	}

	test("0", "0")
	test("1e-13", "0.0000000000001")
	test("1e-13", "1e-13")
	test("1", "1")
	test("100000", "100000")
	test("1e+6", "1000000")

	test("1.49999999999999 ", "1.49999999999999")
	test("1.499999999999999", "1.499999999999999")
	test("1.499999999999999", "1.4999999999999994999999")
	test("1.5              ", "1.4999999999999995000000")
	test("1.5              ", "1.4999999999999995000001")

	test("1.99999999999949 ", "1.99999999999949")
	test("1.999999999999499", "1.999999999999499")
	test("1.999999999999499", "1.9999999999994994999999")
	test("1.9999999999995  ", "1.9999999999994995000000")
	test("1.9999999999995  ", "1.9999999999994995000001")

	test("1.99999999999994 ", "1.99999999999994")
	test("1.999999999999949", "1.999999999999949")
	test("1.999999999999949", "1.9999999999999494999999")
	test("1.99999999999995 ", "1.9999999999999495000000")
	test("1.99999999999995 ", "1.9999999999999495000001")

	test("10.4999999999999 ", "10.4999999999999")
	test("10.49999999999999", "10.49999999999999")
	test("10.49999999999999", "10.499999999999994999999")
	test("10.5             ", "10.499999999999995000000")
	test("10.5             ", "10.499999999999995000001")

	test("1.00000000000499e+11 ", "100000000000.499")
	test("1.000000000004999e+11", "100000000000.4999")
	test("1.000000000005e+11   ", "100000000000.49999")
}

func TestDecimal64ParseDown(t *testing.T) {
	t.Parallel()

	ctx := Context64{Rounding: Down}
	test := func(expected string, source string) {
		t.Helper()
		assert.Equal(t, strings.TrimSpace(expected), ctx.MustParse(source).String())
	}

	test("0", "0")
	test("1e-13", "0.0000000000001")
	test("1e-13", "1e-13")
	test("1", "1")
	test("100000", "100000")
	test("1e+6", "1000000")

	test("1.49999999999999 ", "1.49999999999999")
	test("1.499999999999999", "1.499999999999999")
	test("1.499999999999999", "1.4999999999999994999999")
	test("1.499999999999999", "1.4999999999999995000000")
	test("1.499999999999999", "1.4999999999999995000001")

	test("1.99999999999949 ", "1.99999999999949")
	test("1.999999999999499", "1.999999999999499")
	test("1.999999999999499", "1.9999999999994994999999")
	test("1.999999999999499", "1.9999999999994995000000")
	test("1.999999999999499", "1.9999999999994995000001")

	test("1.99999999999994 ", "1.99999999999994")
	test("1.999999999999949", "1.999999999999949")
	test("1.999999999999949", "1.9999999999999494999999")
	test("1.999999999999949", "1.9999999999999495000000")
	test("1.999999999999949", "1.9999999999999495000001")

	test("10.4999999999999 ", "10.4999999999999")
	test("10.49999999999999", "10.49999999999999")
	test("10.49999999999999", "10.499999999999994999999")
	test("10.49999999999999", "10.499999999999995000000")
	test("10.49999999999999", "10.499999999999995000001")

	test("1.00000000000499e+11 ", "100000000000.499")
	test("1.000000000004999e+11", "100000000000.4999")
	test("1.000000000004999e+11", "100000000000.49999")
}

func TestDecimal64Float64(t *testing.T) {
	require := require.New(t)

	require.Equal(-1.0, NegOne64.Float64())
	require.Equal(0.0, Zero64.Float64())
	require.Equal(1.0, One64.Float64())
	require.Equal(10.0, New64FromInt64(10).Float64())

	oneThird := One64.Quo(New64FromInt64(3))
	one := oneThird.Add(oneThird).Add(oneThird)
	require.InEpsilon(oneThird.Float64(), 1.0/3.0, 0.00000001)
	require.InEpsilon(1.0, one.Float64(), 0.00000001)

	require.True(math.IsNaN(QNaN64.Float64()))
	require.Panics(func() { SNaN64.Float64() })
	require.Equal(math.Inf(1), Infinity64.Float64())
	require.Equal(math.Inf(-1), NegInfinity64.Float64())
}

func TestDecimal64Int64(t *testing.T) {
	t.Parallel()

	assert.EqualValues(t, -1, NegOne64.Int64())
	assert.EqualValues(t, 0, Zero64.Int64())
	assert.EqualValues(t, -0, NegZero64.Int64())
	assert.EqualValues(t, 1, One64.Int64())
	assert.EqualValues(t, 10, New64FromInt64(10).Int64())

	assert.EqualValues(t, 0, QNaN64.Int64())

	assert.EqualValues(t, int64(math.MaxInt64), Infinity64.Int64())
	assert.EqualValues(t, int64(math.MinInt64), NegInfinity64.Int64())
	assert.EqualValues(t, int64(math.MaxInt64), MustParse64("1e100").Int64())
	assert.EqualValues(t, int64(math.MaxInt64), MustParse64("91234567890123456789e20").Int64())
}

func TestDecimal64IsInf(t *testing.T) {
	require.True(t, Infinity64.IsInf())
	require.True(t, NegInfinity64.IsInf())

	require.False(t, Zero64.IsInf())
	require.False(t, NegZero64.IsInf())
	require.False(t, QNaN64.IsInf())
	require.False(t, SNaN64.IsInf())
	require.False(t, New64FromInt64(42).IsInf())
	require.False(t, New64FromInt64(-42).IsInf())
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
		New64FromInt64(42),
		New64FromInt64(-42),
	} {
		require.False(t, n.IsNaN(), "%v", n)
		require.False(t, n.IsQNaN(), "%v", n)
		require.False(t, n.IsSNaN(), "%v", n)
	}
}

func TestDecimal64IsInt(t *testing.T) {
	require := require.New(t)

	fortyTwo := New64FromInt64(42)

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
	require.Equal(true, NegZero64.IsZero())
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

func TestIsSubnormal(t *testing.T) {
	require := require.New(t)

	require.Equal(true, MustParse64("0.1E-383").IsSubnormal())
	require.Equal(true, MustParse64("-0.1E-383").IsSubnormal())
	require.Equal(false, MustParse64("NaN10").IsSubnormal())
	require.Equal(false, New64FromInt64(42).IsSubnormal())

}
