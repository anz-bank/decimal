package decimal

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test64PrecScal(t *testing.T) {
	t.Parallel()

	test := func(s string, prec int) {
		t.Helper()
		expected := MustParse64(s)
		actual := new64(precScale(prec).bits)
		assert.Equal(t, expected, actual)
	}

	test("1", 0)
	test("0.1", 1)
	test("0.01", 2)
	test("0.001", 3)
	test("0.0001", 4)
}

func TestDecimal64String(t *testing.T) {
	assert.Equal(t, strconv.Itoa(0), New64FromInt64(0).String())
	for i := int64(-1000); i <= 1000; i++ {
		assert.Equal(t, strconv.Itoa(int(i)), New64FromInt64(i).String())
	}

	for f := 1; f < 1000; f += 11 {
		fdigits := strings.TrimRight(fmt.Sprintf("%03d", f), "0")
		fraction := New64FromInt64(int64(f)).Quo(New64FromInt64(1000))
		for i := int64(0); i <= 100; i++ {
			require.NotPanics(t, func() {
				assert.Equal(t,
					strconv.Itoa(int(i))+"."+fdigits,
					New64FromInt64(i).Add(fraction).String(),
					"%d.%03d", f, i,
				)
			}, "%d.%03d", f, i)
		}
		for i := int64(-100); i < 0; i++ {
			require.NotPanics(t, func() {
				assert.Equal(t,
					strconv.Itoa(int(i))+"."+fdigits,
					New64FromInt64(i).Sub(fraction).String(),
					"%d.%03d", f, i,
				)
			}, "%d.%03d", f, i)
		}
	}
}

func TestDecimal64StringEdgeCases(t *testing.T) {
	t.Parallel()

	test := func(expected, source string) {
		t.Helper()
		assert.Equal(t, strings.TrimSpace(expected), MustParse64(strings.TrimSpace(source)).String())
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

func BenchmarkDecimal64String(b *testing.B) {
	d := New64FromInt64(123456789)
	for i := 0; i <= b.N; i++ {
		_ = d.String()
	}
}

func TestDecimal64Format(t *testing.T) {
	require := require.New(t)

	for i := int64(-1000); i <= 1000; i++ {
		require.Equal(
			strconv.FormatInt(i, 10),
			fmt.Sprintf("%v", New64FromInt64(i)),
			"%d", i,
		)
	}

	require.Equal("42", New64FromInt64(42).String())
}

func TestDecimal64FormatNaN(t *testing.T) {
	t.Parallel()

	n := MustParse64("-sNan33")
	assert.Equal(t, "-NaN33", n.String())
}

func TestDecimal64FormatPrec(t *testing.T) {
	t.Parallel()
	pi := MustParse64("3.1415926535897932384626433")

	test := func(expected string, prec int, n Decimal64) {
		t.Helper()
		buf := string(*DefaultFormatContext64.append(n, nil, 'f', -1, prec))
		assert.Equal(t, expected, string(buf))
		assert.Equal(t, expected, fmt.Sprintf("%.*f", prec, n))
		assert.Equal(t, expected, n.Text('f', prec))
	}

	assert.Equal(t, "3.141592653589793", pi.String())
	assert.Equal(t, "3.141592653589793", Context64{Rounding: HalfEven}.With(pi).String())
	assert.Equal(t, "3.141592653589793", Context64{Rounding: HalfUp}.With(pi).String())
	assert.Equal(t, "3.141592653589793", fmt.Sprintf("%v", pi))
	assert.Equal(t, "3.141593", fmt.Sprintf("%f", pi))
	assert.Equal(t, "%!q(decimal.Decimal64=3.141592653589793)", fmt.Sprintf("%q", pi))

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

	pi = pi.Add(New64FromInt64(100))
	assert.Equal(t, "103.1415926535898", fmt.Sprintf("%v", pi))
	assert.Equal(t, "103.141593", fmt.Sprintf("%f", pi))
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

	pi = pi.Add(New64FromInt64(100_000))
	assert.Equal(t, "100103.1415926536", fmt.Sprintf("%v", pi))
	assert.Equal(t, "100103.141593", fmt.Sprintf("%f", pi))
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
	pi = pi.Add(New64FromInt64(10_100_000_000))
	assert.Equal(t, "1.010010010314159e+10", fmt.Sprintf("%v", pi))
	assert.Equal(t, "10100100103.141590", fmt.Sprintf("%f", pi))
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

func TestDecimal64FormatPrecEdgeCases(t *testing.T) {
	t.Parallel()

	test := func(expected, input string) {
		n, err := Parse64(input)
		require.NoError(t, err)
		assert.Equal(t, expected, fmt.Sprintf("%.3f", n))
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

func TestDecimal64FormatPrecEdgeCasesHalfUp(t *testing.T) {
	t.Parallel()

	ctx := Context64{Rounding: HalfUp}
	test := func(expected, input string) {
		n, err := Parse64(input)
		require.NoError(t, err)
		assert.Equal(t, expected, ctx.With(n).Text('f', -1, 3))
		assert.Equal(t, expected, fmt.Sprintf("%.3f", ctx.With(n)))
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

func TestDecimal64FormatPrecEdgeCases2(t *testing.T) {
	t.Parallel()

	test := func(expected string, input Decimal64, prec int) {
		t.Helper()
		data := input.Append(nil, 'f', prec)
		assert.Equal(t, expected, string(data))
	}

	test("10000.0000000000", MustParse64("1e4"), 10)
	test("10000000000.0000000000", MustParse64("1e10"), 10)
	test("100000000000.0000000000", MustParse64("1e11"), 10)
	test("100000000000000000000000000000000000000000000000000.0000000000", MustParse64("1e50"), 10)
	test("0.0001000000", MustParse64("1e-4"), 10)
	test("0.0000000001", MustParse64("1e-10"), 10)
	test("0.0000000000", MustParse64("1e-11"), 10)
	test("0.0000000000", MustParse64("1e-20"), 10)
	test("0.0000000000", MustParse64("1e-30"), 10)
	test("0.000000000000000000000000000001", MustParse64("1e-30"), 30)
	test("0.0000000000", Zero64, 10)
	test("-0.0000000000", Zero64.NextMinus(), 10)
	test("0.0000000000", Zero64.NextPlus(), 10)
	test("inf", Infinity64, 10)
	test("9999999999999999000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000.0000000000", Infinity64.NextMinus(), 10)
	test("inf", Infinity64.NextPlus(), 10)

	test("-10000.0000000000", MustParse64("-1e4"), 10)
	test("-10000000000.0000000000", MustParse64("-1e10"), 10)
	test("-100000000000.0000000000", MustParse64("-1e11"), 10)
	test("-100000000000000000000000000000000000000000000000000.0000000000", MustParse64("-1e50"), 10)
	test("-0.0001000000", MustParse64("-1e-4"), 10)
	test("-0.0000000001", MustParse64("-1e-10"), 10)
	test("-0.0000000000", MustParse64("-1e-11"), 10)
	test("-0.0000000000", MustParse64("-1e-20"), 10)
	test("-0.0000000000", MustParse64("-1e-30"), 10)
	test("-0.000000000000000000000000000001", MustParse64("-1e-30"), 30)
	test("-0.0000000000", NegZero64, 10)
	test("-0.0000000000", Zero64.NextMinus(), 10)
	test("0.0000000000", Zero64.NextPlus(), 10)
	test("-inf", NegInfinity64, 10)
	test("-inf", NegInfinity64.NextMinus(), 10)
	test("-9999999999999999000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000.0000000000", NegInfinity64.NextPlus(), 10)
}

func TestDecimal64Format2(t *testing.T) {
	t.Parallel()

	a := MustParse64("0.0001643835616")
	require.Equal(t, "0.000164383562", fmt.Sprintf("%.12f", a))
	b := New64FromInt64(600).Quo(New64FromInt64(10000))
	b = b.Quo(New64FromInt64(365))
	require.Equal(t, "0.000164383562", fmt.Sprintf("%.12f", b))
}

func BenchmarkDecimal64Format(b *testing.B) {
	d := New64FromInt64(123456789)
	for i := 0; i <= b.N; i++ {
		_ = fmt.Sprintf("%v", d)
	}
}

func TestDecimal64Append(t *testing.T) {
	assertAppend := func(expected string, d Decimal64, format byte, prec int) {
		assert.Equal(t, expected, string(d.Append([]byte{}, format, prec)))
	}

	for i := int64(-1000); i <= 1000; i++ {
		d := New64FromInt64(i)
		f := d.Append([]byte{}, 'g', 0)
		assert.Equal(t, strconv.FormatInt(i, 10), string(f), "%d", i)
	}

	assertAppend("NaN", QNaN64, 'g', 0)
	assertAppend("inf", Infinity64, 'g', 0)
	assertAppend("-inf", NegInfinity64, 'g', 0)
	assertAppend("-0", NegZero64, 'g', 0)
	assertAppend("NaN", QNaN64, 'f', 0)
	assertAppend("NaN", SNaN64, 'f', 0)
	assertAppend("inf", Infinity64, 'f', 0)
	assertAppend("-inf", NegInfinity64, 'f', 0)
	assertAppend("%w", Zero64, 'w', 0)

	assertAppend("1.23456789e+8", MustParse64("123456789"), 'e', 0)
	assertAppend("1.23456789e+18", MustParse64("123456789e10"), 'e', 0)
	assertAppend("1.23456789e-18", MustParse64("123456789e-26"), 'e', 0)
	assertAppend("1234567890000000000", MustParse64("123456789e10"), 'f', 0)

	assertAppend("123456789", MustParse64("123456789"), 'g', 0)
	assertAppend("1.23456789e+18", MustParse64("123456789e10"), 'g', 0)
	assertAppend("1.23456789e-18", MustParse64("123456789e-26"), 'g', 0)
	assertAppend("1.23456789e+18", MustParse64("123456789e10"), 'g', 0)

}

func BenchmarkDecimal64Append(b *testing.B) {
	d := New64FromInt64(123456789)
	buf := make([]byte, 10)
	for i := 0; i <= b.N; i++ {
		_ = d.Append(buf, 'g', 0)
	}
}
