package decimal

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseDecimal64(t *testing.T) {
	parseEquals64 := parseEquals64(t)

	for i := int64(-1000); i <= 1000; i++ {
		for _, suffix := range []string{"", ".", ".0", "e0"} {
			s := strconv.Itoa(int(i))
			di := NewDecimal64FromInt64(i)
			parseEquals64(di, s+suffix)
		}
	}
}

func TestParseDecimal64Inf(t *testing.T) {
	parseEquals64 := parseEquals64(t)

	parseEquals64(Infinity64, "Inf")
	parseEquals64(Infinity64, "inf")
	parseEquals64(Infinity64, "∞")
	parseEquals64(NegInfinity64, "-Inf")
	parseEquals64(NegInfinity64, "-inf")
	parseEquals64(NegInfinity64, "-∞")
	parseEquals64(QNaN64, "nan")
	parseEquals64(QNaN64, "NaN")
}

func TestParseDecimal64BadInputs(t *testing.T) {
	for _, input := range []string{
		"", " ", "x",
		" 0", "0 ",
		"++0", "--0", "+-0", "-+0",
		"0..", "0..2",
		"0e", "0ee", "0ee2", "0ex",
	} {
		require.Panics(t, func() {
			MustParseDecimal64(input)
		}, "%v", input)
		_, err := ParseDecimal64(input)
		require.Error(t, err, "%v", input)
	}
}

func TestParseDecimal64BigExp(t *testing.T) {
	parseEquals64 := parseEquals64(t)

	parseEquals64(Zero64, "0e-9999")
	parseEquals64(NegZero64, "-0e-9999")
	parseEquals64(Zero64, "1e-9999")
	parseEquals64(NegZero64, "-1e-9999")

	parseEquals64(Zero64, "0e9999")
	parseEquals64(NegZero64, "-0e9999")
	parseEquals64(Infinity64, "1e9999")
	parseEquals64(NegInfinity64, "-1e9999")
}

func TestParseDecimal64LongMantissa(t *testing.T) {
	parseEquals64 := parseEquals64(t)

	parseEquals64(One64, "1000000000000000000000e-21")
	parseEquals64(NewDecimal64FromInt64(123), "1230000000000000000000e-19")
}

func TestDecimal64ScanFlakyScanState(t *testing.T) {
	requireFailAt := func(text string, failAt int) {
		state := flakyScanState{
			actual: &stringScanner{reader: strings.NewReader(text)},
			failAt: failAt,
		}
		var d Decimal64
		require.Error(t, d.Scan(&state, 'e'))
	}

	requireFailAt("x", 0)
	for i := 0; i < 7; i++ {
		requireFailAt("-1.0e-3", i)
	}
}

func parseEquals64(t *testing.T) func(expected Decimal64, input string) {
	require := require.New(t)

	return func(expected Decimal64, input string) {
		require.NotPanics(func() {
			n := MustParseDecimal64(input)
			require.Equal(expected, n, "%s", input)
		}, "%s", input)

		n, err := ParseDecimal64(input)
		require.NoError(err, "%s", input)
		require.Equal(expected, n, "%s", input)

		n = SNaN64
		count, err := fmt.Sscanf(input, "%g", &n)
		require.NoError(err, "%s", input)
		require.Equal(1, count, "%s", input)
		require.Equal(expected, n, "%s", input)
	}
}
