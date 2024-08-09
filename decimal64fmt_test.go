package decimal

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecimal64String(t *testing.T) {
	require := require.New(t)

	require.Equal(strconv.Itoa(0), New64FromInt64(0).String())
	for i := int64(-1000); i <= 1000; i++ {
		require.Equal(strconv.Itoa(int(i)), New64FromInt64(i).String())
	}

	for f := 1; f < 1000; f += 11 {
		fdigits := strings.TrimRight(fmt.Sprintf("%03d", f), "0")
		fraction := New64FromInt64(int64(f)).Quo(New64FromInt64(1000))
		for i := int64(0); i <= 100; i++ {
			require.Equal(
				strconv.Itoa(int(i))+"."+fdigits,
				New64FromInt64(i).Add(fraction).String(),
				"%d.%03d", f, i,
			)
		}
		for i := int64(-100); i < 0; i++ {
			require.Equal(
				strconv.Itoa(int(i))+"."+fdigits,
				New64FromInt64(i).Sub(fraction).String(),
				"%d.%03d", f, i,
			)
		}
	}
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

	assert.Equal(t, "3.141592653589793", fmt.Sprintf("%v", pi))
	assert.Equal(t, "3.141593", fmt.Sprintf("%f", pi))
	assert.Equal(t, "3", fmt.Sprintf("%.0f", pi))
	assert.Equal(t, "3.1", fmt.Sprintf("%.1f", pi))
	assert.Equal(t, "3.14", fmt.Sprintf("%.2f", pi))
	assert.Equal(t, "3.142", fmt.Sprintf("%.3f", pi))
	assert.Equal(t, "3.141593", fmt.Sprintf("%.6f", pi))
	assert.Equal(t, "3.141592654", fmt.Sprintf("%.9f", pi))
	assert.Equal(t, "3.1415926536", fmt.Sprintf("%.10f", pi))
	assert.Equal(t, "3.141592653589793", fmt.Sprintf("%.15f", pi))
	assert.Equal(t, "3.14159265358979300000", fmt.Sprintf("%.20f", pi))
	assert.Equal(t, "3.1415926535897930"+strings.Repeat("0", 64), fmt.Sprintf("%.80f", pi))
	assert.Equal(t, "3.1415926535897930"+strings.Repeat("0", 100), fmt.Sprintf("%.116f", pi))

	pi = pi.Add(New64FromInt64(100))
	assert.Equal(t, "103.1415926535898", fmt.Sprintf("%v", pi))
	assert.Equal(t, "103.141593", fmt.Sprintf("%f", pi))
	assert.Equal(t, "103", fmt.Sprintf("%.0f", pi))
	assert.Equal(t, "103.1", fmt.Sprintf("%.1f", pi))
	assert.Equal(t, "103.14", fmt.Sprintf("%.2f", pi))
	assert.Equal(t, "103.142", fmt.Sprintf("%.3f", pi))
	assert.Equal(t, "103.141593", fmt.Sprintf("%.6f", pi))
	assert.Equal(t, "103.141592654", fmt.Sprintf("%.9f", pi))
	assert.Equal(t, "103.1415926536", fmt.Sprintf("%.10f", pi))
	assert.Equal(t, "103.141592653589800", fmt.Sprintf("%.15f", pi))
	assert.Equal(t, "103.14159265358980000000", fmt.Sprintf("%.20f", pi))
	assert.Equal(t, "103.1415926535898000"+strings.Repeat("0", 64), fmt.Sprintf("%.80f", pi))
	assert.Equal(t, "103.1415926535898000"+strings.Repeat("0", 100), fmt.Sprintf("%.116f", pi))

	pi = pi.Add(New64FromInt64(100_000))
	assert.Equal(t, "100103.1415926536", fmt.Sprintf("%v", pi))
	assert.Equal(t, "100103.141593", fmt.Sprintf("%f", pi))
	assert.Equal(t, "100103", fmt.Sprintf("%.0f", pi))
	assert.Equal(t, "100103.1", fmt.Sprintf("%.1f", pi))
	assert.Equal(t, "100103.14", fmt.Sprintf("%.2f", pi))
	assert.Equal(t, "100103.142", fmt.Sprintf("%.3f", pi))
	assert.Equal(t, "100103.141593", fmt.Sprintf("%.6f", pi))
	assert.Equal(t, "100103.141592654", fmt.Sprintf("%.9f", pi))
	assert.Equal(t, "100103.1415926536", fmt.Sprintf("%.10f", pi))
	assert.Equal(t, "100103.141592653600000", fmt.Sprintf("%.15f", pi))
	assert.Equal(t, "100103.14159265360000000000", fmt.Sprintf("%.20f", pi))
	assert.Equal(t, "100103.1415926536000000"+strings.Repeat("0", 64), fmt.Sprintf("%.80f", pi))
	assert.Equal(t, "100103.1415926536000000"+strings.Repeat("0", 100), fmt.Sprintf("%.116f", pi))

	// Add five digits to the significand so we round at a 2.
	pi = pi.Add(New64FromInt64(10_100_000_000))
	assert.Equal(t, "10100100103.14159", fmt.Sprintf("%v", pi))
	assert.Equal(t, "10100100103.141590", fmt.Sprintf("%f", pi))
	assert.Equal(t, "10100100103", fmt.Sprintf("%.0f", pi))
	assert.Equal(t, "10100100103.1", fmt.Sprintf("%.1f", pi))
	assert.Equal(t, "10100100103.14", fmt.Sprintf("%.2f", pi))
	assert.Equal(t, "10100100103.142", fmt.Sprintf("%.3f", pi))
	assert.Equal(t, "10100100103.141590", fmt.Sprintf("%.6f", pi))
	assert.Equal(t, "10100100103.141590000", fmt.Sprintf("%.9f", pi))
	assert.Equal(t, "10100100103.1415900000", fmt.Sprintf("%.10f", pi))
	assert.Equal(t, "10100100103.141590000000000", fmt.Sprintf("%.15f", pi))
	assert.Equal(t, "10100100103.14159000000000000000", fmt.Sprintf("%.20f", pi))
	assert.Equal(t, "10100100103.1415900000000000"+strings.Repeat("0", 64), fmt.Sprintf("%.80f", pi))
	assert.Equal(t, "10100100103.1415900000000000"+strings.Repeat("0", 100), fmt.Sprintf("%.116f", pi))
}

func TestDecimal64FormatPrecEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct{ expected, input string }{
		{"0.062", "0.0625"},
		{"0.063", "0.062500001"},
		{"0.062", "0.0625000000000000000000000000000000001"},
		{"-0.062", "-0.0625"},
		{"-0.063", "-0.062500001"},
		{"-0.062", "-0.0625000000000000000000000000000000001"},
		{"0.188", "0.1875"},
		{"0.188", "0.187500001"},
		{"0.188", "0.1875000000000000000000000000000000001"},
		{"-0.188", "-0.1875"},
		{"-0.188", "-0.187500001"},
		{"-0.188", "-0.1875000000000000000000000000000000001"},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			n, err := Parse64(test.input)
			require.NoError(t, err)
			assert.Equal(t, test.expected, fmt.Sprintf("%.3f", n))
		})
	}
}

func TestDecimal64FormatPrecEdgeCasesHalfAway(t *testing.T) {
	t.Parallel()

	tests := []struct{ expected, input string }{
		{"0.063", "0.0625"},
		{"0.063", "0.062500001"},
		{"0.063", "0.0625000000000000000000000000000000001"},
		{"-0.063", "-0.0625"},
		{"-0.063", "-0.062500001"},
		{"-0.063", "-0.0625000000000000000000000000000000001"},
		{"0.188", "0.1875"},
		{"0.188", "0.187500001"},
		{"0.188", "0.1875000000000000000000000000000000001"},
		{"-0.188", "-0.1875"},
		{"-0.188", "-0.187500001"},
		{"-0.188", "-0.1875000000000000000000000000000000001"},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			n, err := Parse64(test.input)
			require.NoError(t, err)
			assert.Equal(t, test.expected, fmt.Sprintf("%.3f", n.RoundHalfAwayFromZero().Float64()))
		})
	}
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
		assert.Equal(t,
			strconv.FormatInt(i, 10),
			string(New64FromInt64(i).Append([]byte{}, 'g', 0)),
		)
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
