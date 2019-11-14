package decimal

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecimal64String(t *testing.T) {
	require := require.New(t)

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

func BenchmarkDecimal64Format(b *testing.B) {
	d := New64FromInt64(123456789)
	for i := 0; i <= b.N; i++ {
		_ = fmt.Sprintf("%v", d)
	}
}

func TestDecimal64Append(t *testing.T) {
	require := require.New(t)

	requireAppend := func(expected string, d Decimal64, format byte, prec int) {
		require.Equal(expected, string(d.Append([]byte{}, format, prec)))
	}

	for i := int64(-1000); i <= 1000; i++ {
		require.Equal(
			strconv.FormatInt(i, 10),
			string(New64FromInt64(i).Append([]byte{}, 'g', 0)),
		)
	}

	requireAppend("NaN", QNaN64, 'g', 0)
	requireAppend("inf", Infinity64, 'g', 0)
	requireAppend("-inf", NegInfinity64, 'g', 0)
	requireAppend("-0", NegZero64, 'g', 0)
	requireAppend("NaN", QNaN64, 'f', 0)
	requireAppend("NaN", SNaN64, 'f', 0)
	requireAppend("inf", Infinity64, 'f', 0)
	requireAppend("-inf", NegInfinity64, 'f', 0)
	requireAppend("%w", Zero64, 'w', 0)

	requireAppend("1.23456789e+8", MustParse64("123456789"), 'e', 0)
	requireAppend("1.23456789e+18", MustParse64("123456789e10"), 'e', 0)
	requireAppend("1.23456789e-18", MustParse64("123456789e-26"), 'e', 0)
	requireAppend("1234567890000000000", MustParse64("123456789e10"), 'f', 0)

	requireAppend("123456789", MustParse64("123456789"), 'g', 0)
	requireAppend("1.23456789e+18", MustParse64("123456789e10"), 'g', 0)
	requireAppend("1.23456789e-18", MustParse64("123456789e-26"), 'g', 0)
	requireAppend("1.23456789e+18", MustParse64("123456789e10"), 'g', 0)

}

func BenchmarkDecimal64Append(b *testing.B) {
	d := New64FromInt64(123456789)
	buf := make([]byte, 10)
	for i := 0; i <= b.N; i++ {
		_ = d.Append(buf, 'g', 0)
	}
}
