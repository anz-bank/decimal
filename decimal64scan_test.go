package decimal

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestParse64(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping TestParse64 in short mode.")
	}
	parseEquals64 := parseEquals64(t)

	for i := int64(-1000); i <= 1000; i++ {
		for _, suffix := range []string{"", ".", ".0", "e0"} {

			s := strconv.Itoa(int(i))
			di := New64FromInt64(i)
			parseEquals64(di, s+suffix)
		}
	}
}

func TestParse64Inf(t *testing.T) {
	t.Parallel()

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

// TODO: Find out what the correct behavior is with bad inputs
// TODO: Does nan get returned if there are leading/trailing whitespaces?
func TestParse64BadInputs(t *testing.T) {
	t.Parallel()

	test := func(input string) {
		t.Helper()
		d, _ := Parse64(input)
		equal(t, SNaN64.IsNaN(), d.IsNaN())
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

func TestParse64BigExp(t *testing.T) {
	t.Parallel()

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

func TestParse64LongMantissa(t *testing.T) {
	t.Parallel()

	parseEquals64 := parseEquals64(t)

	parseEquals64(One64, "1000000000000000000000e-21")
	parseEquals64(New64FromInt64(123), "1230000000000000000000e-19")
}

func TestDecimal64ScanFlakyScanState(t *testing.T) {
	t.Parallel()

	failAt := func(text string, failAt int) {
		state := flakyScanState{
			actual: &scanner{reader: strings.NewReader(text)},
			failAt: failAt,
		}
		var d Decimal64
		notnil(t, d.Scan(&state, 'e'))
	}

	failAt("x", 0)
	for i := 0; i < 7; i++ {
		failAt("-1.0e-3", i)
	}
}

func BenchmarkIOParse64(b *testing.B) {
	var d Decimal64
	for n := 0; n < b.N; n++ {
		buf := bytes.NewBufferString("123456789")
		fmt.Fscanf(buf, "%g", &d) //nolint:errcheck
	}
}

func BenchmarkIODecimal64Scan(b *testing.B) {
	reader := strings.NewReader("")
	for n := 0; n < b.N; n++ {
		reader.Reset("123456789")
		var d Decimal64
		if err := d.Scan(&scanner{reader: reader}, 'g'); err != nil {
			panic("Benchmarking Scan failed")
		}
	}
}

func parseEquals64(t *testing.T) func(expected Decimal64, input string) {
	return func(expected Decimal64, input string) {
		nopanic(t, func() {
			n := MustParse64(input)
			equalD64(t, expected, n)
		})

		n, err := Parse64(input)
		isnil(t, err)
		equalD64(t, expected, n)

		n = SNaN64
		count, err := fmt.Sscanf(input, "%g", &n)
		isnil(t, err)
		equal(t, 1, count)
		equalD64(t, expected, n)
	}
}
