package decimal

import (
	"math"
	"strings"
	"testing"
)

func TestNew64FromInt64(t *testing.T) {
	t.Parallel()

	for i := int64(-1000); i <= 1000; i++ {
		d := New64FromInt64(i)
		j := d.Int64()
		equal(t, i, j)
	}

	// Test the neighborhood of powers of two up to the high-significand
	// representation threshold.
	for e := 4; e < 54; e++ {
		base := int64(1) << uint(e)
		for i := base - 10; i <= base+10; i++ {
			d := New64FromInt64(i)
			j := d.Int64()
			equal(t, i, j)
		}
	}
}

func TestNew64FromInt64Big(t *testing.T) {
	t.Parallel()

	const limit = int64(decimal64Base)
	const step = limit / 997
	for i := -int64(limit); i <= limit; i += step {
		d := New64FromInt64(i)
		j := d.Int64()
		equal(t, i, j)
	}
}

func TestDecimal64Parse(t *testing.T) {
	t.Parallel()

	test := func(expected string, source string) {
		t.Helper()
		equal(t, strings.TrimSpace(expected), MustParse64(source).String())
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
		equal(t, strings.TrimSpace(expected), ctx.MustParse(source).String())
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
		equal(t, strings.TrimSpace(expected), ctx.MustParse(source).String())
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
		equal(t, strings.TrimSpace(expected), ctx.MustParse(source).String())
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
		equal(t, strings.TrimSpace(expected), ctx.MustParse(source).String())
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
	t.Parallel()

	equal(t, -1.0, NegOne64.Float64())
	equal(t, 0.0, Zero64.Float64())
	equal(t, 1.0, One64.Float64())
	equal(t, 10.0, New64FromInt64(10).Float64())

	oneThird := One64.Quo(New64FromInt64(3))
	one := oneThird.Add(oneThird).Add(oneThird)
	epsilon(t, oneThird.Float64(), 1.0/3.0)
	epsilon(t, 1.0, one.Float64())

	check(t, math.IsNaN(QNaN64.Float64()))
	panics(t, func() { SNaN64.Float64() })
	equal(t, math.Inf(1), Infinity64.Float64())
	equal(t, math.Inf(-1), NegInfinity64.Float64())
}

func TestDecimal64Int64(t *testing.T) {
	t.Parallel()

	equal(t, -1, NegOne64.Int64())
	equal(t, 0, Zero64.Int64())
	equal(t, -0, NegZero64.Int64())
	equal(t, 1, One64.Int64())
	equal(t, 10, New64FromInt64(10).Int64())

	equal(t, 0, QNaN64.Int64())

	equal(t, int64(math.MaxInt64), Infinity64.Int64())
	equal(t, int64(math.MinInt64), NegInfinity64.Int64())
	equal(t, int64(math.MaxInt64), MustParse64("1e100").Int64())
	equal(t, int64(math.MaxInt64), MustParse64("91234567890123456789e20").Int64())
}

func TestDecimal64IsInf(t *testing.T) {
	t.Parallel()

	check(t, Infinity64.IsInf())
	check(t, NegInfinity64.IsInf())

	check(t, !Zero64.IsInf())
	check(t, !NegZero64.IsInf())
	check(t, !QNaN64.IsInf())
	check(t, !SNaN64.IsInf())
	check(t, !New64FromInt64(42).IsInf())
	check(t, !New64FromInt64(-42).IsInf())
}

func TestDecimal64IsNaN(t *testing.T) {
	t.Parallel()

	check(t, QNaN64.IsNaN())
	check(t, SNaN64.IsNaN())

	check(t, QNaN64.IsQNaN())
	check(t, !SNaN64.IsQNaN())

	check(t, !QNaN64.IsSNaN())
	check(t, SNaN64.IsSNaN())

	notNaN := func(n Decimal64) {
		check(t, !n.IsNaN())
		check(t, !n.IsQNaN())
		check(t, !n.IsSNaN())
	}
	notNaN(Infinity64)
	notNaN(NegInfinity64)
	notNaN(Zero64)
	notNaN(NegZero64)
	notNaN(New64FromInt64(42))
	notNaN(New64FromInt64(-42))

}

func TestDecimal64IsInt(t *testing.T) {
	t.Parallel()

	fortyTwo := New64FromInt64(42)

	check(t, Zero64.IsInt())
	check(t, fortyTwo.IsInt())
	check(t, fortyTwo.Mul(fortyTwo).IsInt())
	check(t, fortyTwo.Quo(fortyTwo).IsInt())
	check(t, !One64.Quo(fortyTwo).IsInt())

	check(t, !Infinity64.IsInt())
	check(t, !NegInfinity64.IsInt())
	check(t, !QNaN64.IsInt())
	check(t, !SNaN64.IsInt())
}

func TestDecimal64Sign(t *testing.T) {
	t.Parallel()

	equal(t, 0, Zero64.Sign())
	equal(t, 0, NegZero64.Sign())
	equal(t, 1, One64.Sign())
	equal(t, -1, NegOne64.Sign())
}

func TestDecimal64Signbit(t *testing.T) {
	t.Parallel()

	check(t, !Zero64.Signbit())
	check(t, NegZero64.Signbit())
	check(t, !One64.Signbit())
	check(t, NegOne64.Signbit())
}

func TestDecimal64isZero(t *testing.T) {
	t.Parallel()

	check(t, Zero64.IsZero())
	check(t, NegZero64.IsZero())
	check(t, !One64.IsZero())
}

func TestNumDecimalDigits(t *testing.T) {
	t.Parallel()

	for i, num := range tenToThe {
		for j := uint64(1); j < 10 && i < 19; j++ {
			equal(t, i+1, numDecimalDigits(num*j))
		}
	}
}

func TestIsSubnormal(t *testing.T) {
	t.Parallel()

	check(t, MustParse64("0.1E-383").IsSubnormal())
	check(t, MustParse64("-0.1E-383").IsSubnormal())
	check(t, !MustParse64("NaN10").IsSubnormal())
	check(t, !New64FromInt64(42).IsSubnormal())
}
