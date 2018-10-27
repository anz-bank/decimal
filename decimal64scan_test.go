package decimal

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

// type FlakyScanner struct {
// 	actualScanner io.ByteScanner
// }

// func (fs *FlakyScanner) UnreadByte() error {
// 	return fs.actualScanner.UnreadByte()
// }

// func (fs *FlakyScanner) ReadByte() (byte, error) {
// 	b, err := fs.actualScanner.ReadByte()
// 	if err != nil {
// 		if err == io.EOF {
// 			return b, fmt.Errorf("FlakyScanner pretending to fail")
// 		}
// 	}
// 	return b, err
// }

func parseEquals64(t *testing.T, expected Decimal64, input string) {
	require.NotPanics(t, func() {
		MustParseDecimal64(input)
	})
	n, err := ParseDecimal64(input)
	require.NoError(t, err)
	require.Equal(t, expected, n, "%s", input)
}

func TestParseDecimal64(t *testing.T) {
	for i := int64(-1000); i <= 1000; i++ {
		for _, suffix := range []string{"", ".", ".0", "e0"} {
			s := strconv.Itoa(int(i))
			di := NewDecimal64FromInt64(i)
			parseEquals64(t, di, s+suffix)
		}
	}
}

func TestParseDecimal64Inf(t *testing.T) {
	parseEquals64(t, Infinity64, "Inf")
	parseEquals64(t, Infinity64, "inf")
	parseEquals64(t, Infinity64, "∞")
	parseEquals64(t, NegInfinity64, "-Inf")
	parseEquals64(t, NegInfinity64, "-inf")
	parseEquals64(t, NegInfinity64, "-∞")
}

func TestParseDecimal64BadInputs(t *testing.T) {
	for _, input := range []string{
		"", " ", "x",
		" 0", "0 ",
		"++0", "--0", "+-0", "-+0",
		"0..", "0..2",
		"0e", "0ee", "0ex",
	} {
		require.Panics(t, func() {
			MustParseDecimal64(input)
		})
		_, err := ParseDecimal64(input)
		require.Error(t, err, "%v", input)
	}
}

func TestParseDecimal64BigExp(t *testing.T) {
	parseEquals64(t, Zero64, "0e-9999")
	parseEquals64(t, NegZero64, "-0e-9999")
	parseEquals64(t, Zero64, "1e-9999")
	parseEquals64(t, NegZero64, "-1e-9999")

	parseEquals64(t, Zero64, "0e9999")
	parseEquals64(t, NegZero64, "-0e9999")
	parseEquals64(t, Infinity64, "1e9999")
	parseEquals64(t, NegInfinity64, "-1e9999")
}

// func TestParseDecimal64BadScanner
