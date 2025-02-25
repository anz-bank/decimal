package d64

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestPrecScal(t *testing.T) {
	t.Parallel()

	test := func(s string, prec int) {
		t.Helper()
		expected := MustParse(s)
		actual := newDec(precScale(prec).bits)
		equal(t, expected, actual)
	}

	test("1", 0)
	test("0.1", 1)
	test("0.01", 2)
	test("0.001", 3)
	test("0.0001", 4)
}

func TestDecimalString(t *testing.T) {
	t.Parallel()

	equal(t, strconv.Itoa(0), NewFromInt64(0).String())
	for i := int64(-1000); i <= 1000; i++ {
		equal(t, strconv.Itoa(int(i)), NewFromInt64(i).String())
	}

	for f := 1; f < 1000; f += 11 {
		fdigits := strings.TrimRight(fmt.Sprintf("%03d", f), "0")
		fraction := NewFromInt64(int64(f)).Quo(NewFromInt64(1000))
		for i := int64(0); i <= 100; i++ {
			nopanic(t, func() {
				equal(t,
					strconv.Itoa(int(i))+"."+fdigits,
					NewFromInt64(i).Add(fraction).String(),
				)
			})
		}
		for i := int64(-100); i < 0; i++ {
			nopanic(t, func() {
				equal(t,
					strconv.Itoa(int(i))+"."+fdigits,
					NewFromInt64(i).Sub(fraction).String(),
				)
			})
		}
	}
}

func TestDecimalStringEdgeCases(t *testing.T) {
	t.Parallel()

	test := func(expected, source string) {
		t.Helper()
		equal(t, strings.TrimSpace(expected), MustParse(strings.TrimSpace(source)).String())
	}
	test(" 123456", "123456")
	test("-123456", "-123456")
	test(" 1.234567e+6", "1234567")
	test("-1.234567e+6", "-1234567")
	test(" 0.0001", "0.0001")
	test("-0.0001", "-0.0001")
	test(" 1e-5", "0.00001")
	test("-1e-5", "-0.00001")
	test(" 9.999999999999999e+384", "9.999999999999999e+384")
	test("-9.999999999999999e+384", "-9.999999999999999e+384")
	test(" 1e-398", " 1e-398")
	test("-1e-398", "-1e-398")

	// regression for prefix-zeros-after-dot bug
	test("  1.666666666666667", "  1.666666666666667")
	test("0.01666666666666667", "0.01666666666666667")
}

// Non-representative sample, but retained for comparison purposes.
func BenchmarkIODecimalString(b *testing.B) {
	d := NewFromInt64(123456789)
	for i := 0; i <= b.N; i++ {
		_ = d.String()
	}
}

func BenchmarkIODecimalString2(b *testing.B) {
	dd := []Decimal{
		Zero,
		Pi,
		NewFromInt64(123456789),
		MustParse("-12345678901234E-380"),
		MustParse("+12345678901234E+380"),
		QNaN,
		Inf,
	}
	for i := 0; i <= b.N; i++ {
		_ = dd[i%len(dd)].String()
	}
}

func TestDecimalFormat(t *testing.T) {
	t.Parallel()

	for i := int64(-1000); i <= 1000; i++ {
		equal(t, strconv.FormatInt(i, 10), fmt.Sprintf("%v", NewFromInt64(i)))
	}

	equal(t, "42", NewFromInt64(42).String())
}

func TestDecimalFormatNaN(t *testing.T) {
	t.Parallel()

	n := MustParse("-sNaN33")
	equal(t, "-NaN33", n.String())
}

func TestDecimalFormatPrec(t *testing.T) {
	t.Parallel()

	pi := MustParse("3.1415926535897932384626433")

	test := func(expected string, prec int, n Decimal) {
		t.Helper()
		var buf [32]byte
		actual := string(DefaultFormatContext.append(n, buf[:0], -1, prec, noFlags, 'f'))
		equal(t, expected, actual)
		equal(t, expected, fmt.Sprintf("%.*f", prec, n))
		equal(t, expected, n.Text('f', prec))
	}

	equal(t, "3.141592653589793", pi.String())
	equal(t, "3.141592653589793", Context{Rounding: HalfEven}.With(pi).String())
	equal(t, "3.141592653589793", Context{Rounding: HalfUp}.With(pi).String())
	equal(t, "3.141592653589793", fmt.Sprintf("%v", pi))
	equal(t, "3.141593", fmt.Sprintf("%f", pi))
	equal(t, "%!q(d64.Decimal=3.141592653589793)", fmt.Sprintf("%q", pi))

	test("3", 0, pi)
	test("3.1", 1, pi)
	test("3.14", 2, pi)
	test("3.142", 3, pi)
	test("3.141593", 6, pi)
	test("3.141592654", 9, pi)
	test("3.1415926536", 10, pi)
	test("3.141592653589793", 15, pi)
	test("3.14159265358979300000", 20, pi)
	test("3.1415926535897930"+strings.Repeat("0", 64), 80, pi)
	test("3.1415926535897930"+strings.Repeat("0", 100), 116, pi)

	pi = pi.Add(NewFromInt64(100))
	equal(t, "103.1415926535898", fmt.Sprintf("%v", pi))
	equal(t, "103.141593", fmt.Sprintf("%f", pi))
	test("103", 0, pi)
	test("103.1", 1, pi)
	test("103.14", 2, pi)
	test("103.142", 3, pi)
	test("103.141593", 6, pi)
	test("103.141592654", 9, pi)
	test("103.1415926536", 10, pi)
	test("103.141592653589800", 15, pi)
	test("103.14159265358980000000", 20, pi)
	test("103.1415926535898000"+strings.Repeat("0", 64), 80, pi)
	test("103.1415926535898000"+strings.Repeat("0", 100), 116, pi)

	pi = pi.Add(NewFromInt64(100_000))
	equal(t, "100103.1415926536", fmt.Sprintf("%v", pi))
	equal(t, "100103.141593", fmt.Sprintf("%f", pi))
	test("100103", 0, pi)
	test("100103.1", 1, pi)
	test("100103.14", 2, pi)
	test("100103.142", 3, pi)
	test("100103.141593", 6, pi)
	test("100103.141592654", 9, pi)
	test("100103.1415926536", 10, pi)
	test("100103.141592653600000", 15, pi)
	test("100103.14159265360000000000", 20, pi)
	test("100103.1415926536000000"+strings.Repeat("0", 64), 80, pi)
	test("100103.1415926536000000"+strings.Repeat("0", 100), 116, pi)

	// // Add five digits to the significand so that we round at a 2.
	pi = pi.Add(NewFromInt64(10_100_000_000))
	equal(t, "1.010010010314159e+10", fmt.Sprintf("%v", pi))
	equal(t, "10100100103.141590", fmt.Sprintf("%f", pi))
	test("10100100103", 0, pi)
	test("10100100103.1", 1, pi)
	test("10100100103.14", 2, pi)
	test("10100100103.142", 3, pi)
	test("10100100103.141590", 6, pi)
	test("10100100103.141590000", 9, pi)
	test("10100100103.1415900000", 10, pi)
	test("10100100103.141590000000000", 15, pi)
	test("10100100103.14159000000000000000", 20, pi)
	test("10100100103.1415900000000000"+strings.Repeat("0", 64), 80, pi)
	test("10100100103.1415900000000000"+strings.Repeat("0", 100), 116, pi)
}

func TestDecimalFormatPrecEdgeCases(t *testing.T) {
	t.Parallel()

	test := func(expected, input string) {
		n, err := Parse(input)
		isnil(t, err)
		equal(t, expected, fmt.Sprintf("%.3f", n))
	}

	test("0.062", "0.0625")
	test("0.063", "0.062500001")
	test("0.062", "0.0625000000000000000000000000000000001")
	test("-0.062", "-0.0625")
	test("-0.063", "-0.062500001")
	test("-0.062", "-0.0625000000000000000000000000000000001")
	test("0.188", "0.1875")
	test("0.188", "0.187500001")
	test("0.188", "0.1875000000000000000000000000000000001")
	test("-0.188", "-0.1875")
	test("-0.188", "-0.187500001")
	test("-0.188", "-0.1875000000000000000000000000000000001")
}

func TestDecimalFormatPrecEdgeCasesHalfUp(t *testing.T) {
	t.Parallel()

	ctx := Context{Rounding: HalfUp}
	test := func(expected, input string) {
		n, err := Parse(input)
		isnil(t, err)
		equal(t, expected, ctx.With(n).Text('f', -1, 3))
		equal(t, expected, fmt.Sprintf("%.3f", ctx.With(n)))
	}

	test("0.063", "0.0625")
	test("0.063", "0.062500001")
	test("0.063", "0.0625000000000000000000000000000000001")
	test("-0.063", "-0.0625")
	test("-0.063", "-0.062500001")
	test("-0.063", "-0.0625000000000000000000000000000000001")
	test("0.188", "0.1875")
	test("0.188", "0.187500001")
	test("0.188", "0.1875000000000000000000000000000000001")
	test("-0.188", "-0.1875")
	test("-0.188", "-0.187500001")
	test("-0.188", "-0.1875000000000000000000000000000000001")
}

func TestDecimalFormatPrecEdgeCases2(t *testing.T) {
	t.Parallel()

	test := func(expected string, input Decimal, prec int) {
		t.Helper()
		data := input.Append(nil, 'f', prec)
		equal(t, expected, string(data))
	}

	test("10000.0000000000", MustParse("1e4"), 10)
	test("10000000000.0000000000", MustParse("1e10"), 10)
	test("100000000000.0000000000", MustParse("1e11"), 10)
	test("100000000000000000000000000000000000000000000000000.0000000000", MustParse("1e50"), 10)
	test("0.0001000000", MustParse("1e-4"), 10)
	test("0.0000000001", MustParse("1e-10"), 10)
	test("0.0000000000", MustParse("1e-11"), 10)
	test("0.0000000000", MustParse("1e-20"), 10)
	test("0.0000000000", MustParse("1e-30"), 10)
	test("0.000000000000000000000000000001", MustParse("1e-30"), 30)
	test("0.0000000000", Zero, 10)
	test("-0.0000000000", Zero.NextMinus(), 10)
	test("0.0000000000", Zero.NextPlus(), 10)
	test("inf", Inf, 10)
	test("9999999999999999000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000.0000000000", Inf.NextMinus(), 10)
	test("inf", Inf.NextPlus(), 10)

	test("-10000.0000000000", MustParse("-1e4"), 10)
	test("-10000000000.0000000000", MustParse("-1e10"), 10)
	test("-100000000000.0000000000", MustParse("-1e11"), 10)
	test("-100000000000000000000000000000000000000000000000000.0000000000", MustParse("-1e50"), 10)
	test("-0.0001000000", MustParse("-1e-4"), 10)
	test("-0.0000000001", MustParse("-1e-10"), 10)
	test("-0.0000000000", MustParse("-1e-11"), 10)
	test("-0.0000000000", MustParse("-1e-20"), 10)
	test("-0.0000000000", MustParse("-1e-30"), 10)
	test("-0.000000000000000000000000000001", MustParse("-1e-30"), 30)
	test("-0.0000000000", NegZero, 10)
	test("-0.0000000000", Zero.NextMinus(), 10)
	test("0.0000000000", Zero.NextPlus(), 10)
	test("-inf", NegInf, 10)
	test("-inf", NegInf.NextMinus(), 10)
	test("-9999999999999999000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000.0000000000", NegInf.NextPlus(), 10)
}

func TestDecimalFormat2(t *testing.T) {
	t.Parallel()

	a := MustParse("0.0001643835616")
	equal(t, "0.000164383562", fmt.Sprintf("%.12f", a))
	b := NewFromInt64(600).Quo(NewFromInt64(10000))
	b = b.Quo(NewFromInt64(365))
	equal(t, "0.000164383562", fmt.Sprintf("%.12f", b))
}

func BenchmarkIODecimalFormat(b *testing.B) {
	d := NewFromInt64(123456789)
	for i := 0; i <= b.N; i++ {
		_ = fmt.Sprintf("%v", d)
	}
}

func TestDecimalAppend(t *testing.T) {
	t.Parallel()

	assertAppend := func(expected string, d Decimal, format byte, prec int) {
		equal(t, expected, string(d.Append([]byte{}, format, prec)))
	}

	for i := int64(-1000); i <= 1000; i++ {
		d := NewFromInt64(i)
		f := d.Append([]byte{}, 'g', 0)
		equal(t, strconv.FormatInt(i, 10), string(f))
	}

	assertAppend("NaN", QNaN, 'g', 0)
	assertAppend("inf", Inf, 'g', 0)
	assertAppend("-inf", NegInf, 'g', 0)
	assertAppend("-0", NegZero, 'g', 0)
	assertAppend("NaN", QNaN, 'f', 0)
	assertAppend("NaN", SNaN, 'f', 0)
	assertAppend("inf", Inf, 'f', 0)
	assertAppend("-inf", NegInf, 'f', 0)
	assertAppend("%w", Zero, 'w', 0)

	assertAppend("1.23456789e+8", MustParse("123456789"), 'e', 0)
	assertAppend("1.23456789e+18", MustParse("123456789e10"), 'e', 0)
	assertAppend("1.23456789e-18", MustParse("123456789e-26"), 'e', 0)
	assertAppend("1234567890000000000", MustParse("123456789e10"), 'f', 0)

	assertAppend("123456789", MustParse("123456789"), 'g', 0)
	assertAppend("1.23456789e+18", MustParse("123456789e10"), 'g', 0)
	assertAppend("1.23456789e-18", MustParse("123456789e-26"), 'g', 0)
	assertAppend("1.23456789e+18", MustParse("123456789e10"), 'g', 0)

}

func BenchmarkIODecimalAppend(b *testing.B) {
	d := NewFromInt64(123456789)
	var buf [32]byte
	for i := 0; i <= b.N; i++ {
		_ = d.Append(buf[:0], 'g', 0)
	}
}
