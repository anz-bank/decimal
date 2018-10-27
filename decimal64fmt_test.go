package decimal

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecimal64String(t *testing.T) {
	require := require.New(t)

	for i := int64(-1000); i <= 1000; i++ {
		require.Equal(strconv.Itoa(int(i)), NewDecimal64FromInt64(i).String())
	}

	for f := 1; f < 100; f += 2 {
		fraction := NewDecimal64FromInt64(int64(f)).Quo(NewDecimal64FromInt64(100))
		for i := int64(0); i <= 100; i++ {
			require.Equal(
				strconv.Itoa(int(i))+"."+strconv.Itoa(100 + f)[1:3],
				NewDecimal64FromInt64(i).Add(fraction).String(),
			)
		}
		for i := int64(-100); i < 0; i++ {
			require.Equal(
				strconv.Itoa(int(i))+"."+strconv.Itoa(100 + f)[1:3],
				NewDecimal64FromInt64(i).Sub(fraction).String(),
			)
		}
	}
}

func TestDecimal64Format(t *testing.T) {
	require := require.New(t)

	for i := int64(-1000); i <= 1000; i++ {
		require.Equal(strconv.FormatInt(i, 10), fmt.Sprintf("%v", NewDecimal64FromInt64(i)))
	}

	require.Equal("%!s(*decimal.Decimal64=42)", fmt.Sprintf("%s", NewDecimal64FromInt64(42)))
}

func TestDecimal64Append(t *testing.T) {
	require := require.New(t)

	requireAppend := func(expected string, d Decimal64, format byte, prec int) {
		require.Equal(expected, string(d.Append([]byte{}, format, prec)))
	}

	// for i := int64(-1000); i <= 1000; i++ {
	// 	require.Equal(
	// 		strconv.FormatInt(i, 10),
	// 		string(NewDecimal64FromInt64(i).Append([]byte{}, 'g', 0)),
	// 	)
	// }

	// requireAppend("NaN", QNaN64, 'g', 0)
	// requireAppend("inf", Infinity64, 'g', 0)
	// requireAppend("-inf", NegInfinity64, 'g', 0)
	// requireAppend("-0", NegZero64, 'g', 0)

	requireAppend("1.23456789e+8", MustParseDecimal64("123456789"), 'e', 0)
	requireAppend("1.23456789e+18", MustParseDecimal64("123456789e10"), 'e', 0)
	requireAppend("1.23456789e-18", MustParseDecimal64("123456789e-26"), 'e', 0)
	requireAppend("1234567890000000000", MustParseDecimal64("123456789e10"), 'f', 0)

	requireAppend("123456789", MustParseDecimal64("123456789"), 'g', 0)
	requireAppend("1.23456789e+18", MustParseDecimal64("123456789e10"), 'g', 0)
	requireAppend("1.23456789e-18", MustParseDecimal64("123456789e-26"), 'g', 0)
	requireAppend("1.23456789e+18", MustParseDecimal64("123456789e10"), 'g', 0)

	requireAppend("nan", QNaN64, 'f', 0)
	requireAppend("nan", SNaN64, 'f', 0)

	requireAppend("inf", Infinity64, 'f', 0)
	requireAppend("-inf", NegInfinity64, 'f', 0)

	requireAppend("%w", Zero64, 'w', 0)
}
