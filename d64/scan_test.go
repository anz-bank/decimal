package d64

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping TestParse in short mode.")
	}
	parseEquals := parseEquals(t)

	for i := int64(-1000); i <= 1000; i++ {
		for _, suffix := range []string{"", ".", ".0", "e0"} {

			s := strconv.Itoa(int(i))
			di := NewFromInt64(i)
			parseEquals(di, s+suffix)
		}
	}
}

func TestParseInf(t *testing.T) {
	t.Parallel()

	parseEquals := parseEquals(t)

	parseEquals(Inf, "Inf")
	parseEquals(Inf, "inf")
	parseEquals(Inf, "∞")
	parseEquals(NegInf, "-Inf")
	parseEquals(NegInf, "-inf")
	parseEquals(NegInf, "-∞")
	parseEquals(QNaN, "nan")
	parseEquals(QNaN, "NaN")
}

// TODO: Find out what the correct behavior is with bad inputs
// TODO: Does nan get returned if there are leading/trailing whitespaces?
func TestParseBadInputs(t *testing.T) {
	t.Parallel()

	test := func(input string) {
		t.Helper()
		d, _ := Parse(input)
		equal(t, SNaN.IsNaN(), d.IsNaN())
	}
	test("")
	test(" ")
	test("x")
	test("++0")
	test("--0")
	test("+-0")
	test("-+0")
	test("0..")
	test("0..2")
	test("0e")
	test("0ee")
	test("0ee2")
	test("0ex")
}

func TestParseBigExp(t *testing.T) {
	t.Parallel()

	parseEquals := parseEquals(t)

	parseEquals(Zero, "0e-9999")
	parseEquals(NegZero, "-0e-9999")
	parseEquals(Zero, "1e-9999")
	parseEquals(NegZero, "-1e-9999")

	parseEquals(Zero, "0e9999")
	parseEquals(NegZero, "-0e9999")
	parseEquals(Inf, "1e9999")
	parseEquals(NegInf, "-1e9999")
}

func TestParseLongMantissa(t *testing.T) {
	t.Parallel()

	parseEquals := parseEquals(t)

	parseEquals(One, "1000000000000000000000e-21")
	parseEquals(NewFromInt64(123), "1230000000000000000000e-19")
}

func TestDecimalScanFlakyScanState(t *testing.T) {
	t.Parallel()

	failAt := func(text string, failAt int) {
		state := flakyScanState{
			actual: &scanner{reader: strings.NewReader(text)},
			failAt: failAt,
		}
		var d Decimal
		notnil(t, d.Scan(&state, 'e'))
	}

	failAt("x", 0)
	for i := 0; i < 7; i++ {
		failAt("-1.0e-3", i)
	}
}

func BenchmarkIOParse(b *testing.B) {
	var d Decimal
	for n := 0; n < b.N; n++ {
		buf := bytes.NewBufferString("123456789")
		fmt.Fscanf(buf, "%g", &d) //nolint:errcheck
	}
}

func BenchmarkIODecimalScan(b *testing.B) {
	reader := strings.NewReader("")
	for n := 0; n < b.N; n++ {
		reader.Reset("123456789")
		var d Decimal
		if err := d.Scan(&scanner{reader: reader}, 'g'); err != nil {
			panic("Benchmarking Scan failed")
		}
	}
}

func parseEquals(t *testing.T) func(expected Decimal, input string) {
	return func(expected Decimal, input string) {
		nopanic(t, func() {
			n := MustParse(input)
			equalD64(t, expected, n)
		})

		n, err := Parse(input)
		isnil(t, err)
		equalD64(t, expected, n)

		n = SNaN
		count, err := fmt.Sscanf(input, "%g", &n)
		isnil(t, err)
		equal(t, 1, count)
		equalD64(t, expected, n)
	}
}
