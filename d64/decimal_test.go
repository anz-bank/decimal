package d64

import (
	"math"
	"strings"
	"testing"
)

func TestNewFromInt64(t *testing.T) {
	t.Parallel()

	for i := int64(0); i <= 1000; i++ {
		replayOnFail(t, func() {
			d := NewFromInt64(i)
			j := d.Int64()
			equal(t, i, j)
		})
		replayOnFail(t, func() {
			d := NewFromInt64(-i)
			j := d.Int64()
			equal(t, -i, j)
		})
	}

	// Test the neighborhood of powers of two up to the high-significand
	// representation threshold.
	for e := 4; e < 54; e++ {
		base := int64(1) << uint(e)
		for i := base - 10; i <= base+10; i++ {
			replayOnFail(t, func() {
				d := NewFromInt64(i)
				j := d.Int64()
				equal(t, i, j)
			})
		}
	}
}

func TestNewFromInt64Big(t *testing.T) {
	t.Parallel()

	const limit = int64(decimalBase)
	const step = limit / 997
	for i := -int64(limit); i <= limit; i += step {
		replayOnFail(t, func() {
			d := NewFromInt64(i)
			j := d.Int64()
			equal(t, i, j)
		})
	}
}

func TestDecimalParse(t *testing.T) {
	t.Parallel()

	test := func(expected string, source string) {
		t.Helper()
		equal(t, strings.TrimSpace(expected), MustParse(source).String())
	}

	test("0", "0")
	test("1e-13", "0.0000000000001")
	test("1e-13", "1e-13")
	test("1", "1")
	test("100000", "100000")
	test("1e+6", "1000000")
}

func TestDecimalParseHalfEvenOdd(t *testing.T) {
	t.Parallel()

	ctx := Context{Rounding: HalfEven}
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

func TestDecimalParseHalfEvenEven(t *testing.T) {
	t.Parallel()

	ctx := Context{Rounding: HalfEven}
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

func TestDecimalParseHalfUp(t *testing.T) {
	t.Parallel()

	ctx := Context{Rounding: HalfUp}
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

func TestDecimalParseDown(t *testing.T) {
	t.Parallel()

	ctx := Context{Rounding: Down}
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

func TestDecimalFloat64(t *testing.T) {
	t.Parallel()

	equal(t, -1.0, NegOne.Float64())
	equal(t, 0.0, Zero.Float64())
	equal(t, 1.0, One.Float64())
	equal(t, 10.0, NewFromInt64(10).Float64())

	oneThird := One.Quo(NewFromInt64(3))
	one := oneThird.Add(oneThird).Add(oneThird)
	epsilon(t, oneThird.Float64(), 1.0/3.0)
	epsilon(t, 1.0, one.Float64())

	check(t, math.IsNaN(QNaN.Float64()))
	panics(t, func() { SNaN.Float64() })
	equal(t, math.Inf(1), Inf.Float64())
	equal(t, math.Inf(-1), NegInf.Float64())
}

func TestDecimalInt64(t *testing.T) {
	t.Parallel()

	equal(t, -1, NegOne.Int64())
	equal(t, 0, Zero.Int64())
	equal(t, -0, NegZero.Int64())
	equal(t, 1, One.Int64())
	equal(t, 10, NewFromInt64(10).Int64())

	equal(t, 0, QNaN.Int64())

	equal(t, int64(math.MaxInt64), Inf.Int64())
	equal(t, int64(math.MinInt64), NegInf.Int64())
	equal(t, int64(math.MaxInt64), MustParse("1e100").Int64())
	equal(t, int64(math.MaxInt64), MustParse("91234567890123456789e20").Int64())
}

func TestDecimal64IsInf(t *testing.T) {
	t.Parallel()

	check(t, Inf.IsInf())
	check(t, NegInf.IsInf())

	check(t, !Zero.IsInf())
	check(t, !NegZero.IsInf())
	check(t, !QNaN.IsInf())
	check(t, !SNaN.IsInf())
	check(t, !NewFromInt64(42).IsInf())
	check(t, !NewFromInt64(-42).IsInf())
}

func TestDecimalIsNaN(t *testing.T) {
	t.Parallel()

	check(t, QNaN.IsNaN())
	check(t, SNaN.IsNaN())

	check(t, QNaN.IsQNaN())
	check(t, !SNaN.IsQNaN())

	check(t, !QNaN.IsSNaN())
	check(t, SNaN.IsSNaN())

	notNaN := func(n Decimal) {
		check(t, !n.IsNaN())
		check(t, !n.IsQNaN())
		check(t, !n.IsSNaN())
	}
	notNaN(Inf)
	notNaN(NegInf)
	notNaN(Zero)
	notNaN(NegZero)
	notNaN(NewFromInt64(42))
	notNaN(NewFromInt64(-42))

}

func TestDecimalIsInt(t *testing.T) {
	t.Parallel()

	fortyTwo := NewFromInt64(42)

	check(t, Zero.IsInt())
	check(t, fortyTwo.IsInt())
	check(t, fortyTwo.Mul(fortyTwo).IsInt())
	check(t, fortyTwo.Quo(fortyTwo).IsInt())
	check(t, !One.Quo(fortyTwo).IsInt())

	check(t, !Inf.IsInt())
	check(t, !NegInf.IsInt())
	check(t, !QNaN.IsInt())
	check(t, !SNaN.IsInt())
}

func TestDecimalSign(t *testing.T) {
	t.Parallel()

	equal(t, 0, Zero.Sign())
	equal(t, 0, NegZero.Sign())
	equal(t, 1, One.Sign())
	equal(t, -1, NegOne.Sign())
}

func TestDecimalSignbit(t *testing.T) {
	t.Parallel()

	check(t, !Zero.Signbit())
	check(t, NegZero.Signbit())
	check(t, !One.Signbit())
	check(t, NegOne.Signbit())
}

func TestDecimalIsZero(t *testing.T) {
	t.Parallel()

	check(t, Zero.IsZero())
	check(t, NegZero.IsZero())
	check(t, !One.IsZero())
}

func TestIsSubnormal(t *testing.T) {
	t.Parallel()

	check(t, MustParse("0.1E-383").IsSubnormal())
	check(t, MustParse("-0.1E-383").IsSubnormal())
	check(t, !MustParse("NaN10").IsSubnormal())
	check(t, !NewFromInt64(42).IsSubnormal())
}
