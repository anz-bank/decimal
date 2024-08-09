package decimal

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type opResult struct {
	val1, val2, val3, result Decimal64
	text                     string
}

type testCase struct {
	name           string
	function       string
	val1           string
	val2           string
	val3           string
	expectedResult string
	rounding       string
}

func (testVal *testCase) String() string {
	if testVal == nil {
		return "nil"
	}
	return fmt.Sprintf("%s %s (%v, %v, %v) -> %v", testVal.name, testVal.function, testVal.val1, testVal.val2, testVal.val3, testVal.expectedResult)
}

type set[K comparable] map[K]struct{}

func (s set[K]) Has(k K) bool {
	_, ok := s[k]
	return ok
}

var supportedRounding = set[string]{"half_up": {}, "half_even": {}}
var ignoredFunctions = set[string]{"apply": {}}

// TestFromSuite is the master tester for the dectest suite.
func TestFromSuite(t *testing.T) {
	for _, file := range []string{
		"dectest/ddAdd.decTest",
		"dectest/ddMultiply.decTest",
		"dectest/ddFMA.decTest",
		"dectest/ddClass.decTest",
		// TODO: Implement following tests
		"dectest/ddCompare.decTest",
		"dectest/ddAbs.decTest",
		// "dectest/ddCopysign.decTest",
		"dectest/ddDivide.decTest",
		// 	"dectest/ddLogB.decTest",
		"dectest/ddMin.decTest",
		"dectest/ddMinMag.decTest",
		"dectest/ddMinus.decTest",
	} {
		t.Run(file, func(t *testing.T) {
			f, _ := os.Open(file)
			scanner := bufio.NewScanner(f)
			numTests := 0
			var roundingSupported bool
			var scannedContext Context64
			for scanner.Scan() {
				testVal := getInput(scanner.Text())
				if testVal == nil {
					continue
				}
				if testVal.rounding != "" {
					roundingSupported = supportedRounding.Has(testVal.rounding)
					if roundingSupported {
						scannedContext = setRoundingFromString(testVal.rounding)
					}
				}
				if testVal.function != "" && roundingSupported {
					numTests++
					t.Run(testVal.name, func(t *testing.T) {
						dec64vals, err := convertToDec64(testVal)
						require.NoError(t, err)
						runTest(t, scannedContext, dec64vals, testVal)
					})
				}
			}
		})
	}
}

func setRoundingFromString(s string) Context64 {
	switch s {
	case "half_even":
		return Context64{roundHalfEven}
	case "half_up":
		return Context64{roundHalfUp}
	case "default":
		return DefaultContext
	default:
		panic("Rounding not supported" + s)
	}
}

func isRoundingErr(res, expected Decimal64) bool {
	var resP decParts
	resP.unpack(res)
	var expectedP decParts
	expectedP.unpack(expected)
	sigDiff := int64(resP.significand.lo - expectedP.significand.lo)
	expDiff := resP.exp - expectedP.exp
	if (sigDiff == 1 || sigDiff == -1) && (expDiff == 1 || expDiff == -1 || expDiff == 0) {
		return true
	}
	if resP.significand.lo == maxSig && resP.exp == expMax && expectedP.fl == flInf {
		return true
	}
	return false
}

var (
	testRegex     = regexp.MustCompile(`'((?:''+|[^'])*)'|(\S+)`)
	roundingRegex = regexp.MustCompile(`(?:rounding:[\s]*)(?P<rounding>[\S]*)`)
)

// getInput gets the test file and extracts test using regex, then returns a map object and a list of test names.
func getInput(line string) *testCase {
	// TODO: Figure out what this comment means.
	// Add regex to match to  rounding: rounding mode here

	m := testRegex.FindAllStringSubmatch(line, -1)
	if m == nil || !strings.HasPrefix(m[0][2], "dd") {
		m := roundingRegex.FindStringSubmatch(line)
		if m == nil {
			return nil
		}
		return &testCase{rounding: m[1]}
	}
	fields := make([]string, 0, len(m))
	for _, f := range m {
		fields = append(fields, strings.ReplaceAll(f[1], "''", "'")+f[2])
	}
	if i := slices.Index(fields, "->"); i < 5 {
		if i == -1 {
			panic(fmt.Errorf("malformed input: %s", line))
		}
		head, tail := fields[:i], fields[i:]
		for ; i < 5; i++ {
			head = append(append([]string{}, head...), "")
		}
		fields = append(head, tail...)
	}
	test := &testCase{
		name:           fields[0],
		function:       fields[1],
		val1:           fields[2],
		val2:           fields[3],
		val3:           fields[4],
		expectedResult: fields[6], // field[6] == "->"
	}
	if ignoredFunctions.Has(test.function) {
		return nil
	}
	// # represents a null value, which isn't meaningful for this package.
	if test.val1 == "#" || test.val2 == "#" {
		return nil
	}
	return test
}

// convertToDec64 converts the map object strings to decimal64s.
func convertToDec64(testvals *testCase) (opResult, error) {
	var r opResult
	var err error
	parseNotEmpty := func(s string) (Decimal64, error) {
		if s == "" {
			return QNaN64, nil
		}
		return Parse64(s)
	}
	r.val1, err = parseNotEmpty(testvals.val1)
	if err != nil {
		return opResult{}, fmt.Errorf("error parsing val1: %w", err)
	}
	r.val2, err = parseNotEmpty(testvals.val2)
	if err != nil {
		return opResult{}, fmt.Errorf("error parsing val2: %w", err)
	}
	r.val3, err = parseNotEmpty(testvals.val3)
	if err != nil {
		return opResult{}, fmt.Errorf("error parsing val3: %w", err)
	}
	if textResults.Has(testvals.function) {
		r.text = testvals.expectedResult
	} else {
		r.result, err = parseNotEmpty(testvals.expectedResult)
		if err != nil {
			return opResult{}, fmt.Errorf("error parsing expected: %w", err)
		}
	}
	return r, nil
}

// runTest completes the tests and compares actual and expected results.
func runTest(t *testing.T, context Context64, expected opResult, testValStrings *testCase) bool {
	actual := execOp(context, expected.val1, expected.val2, expected.val3, testValStrings.function)

	if actual.text != "" {
		if testValStrings.function == "compare" && actual.text == "-2" && expected.result.IsNaN() {
			return true
		}
		if actual.text != testValStrings.expectedResult {
			return assert.Failf(t, "unexpected result", "test:\n%s\ncalculated text: %s", testValStrings, actual.text)
		}
		return true
	}
	if actual.result.IsNaN() || expected.result.IsNaN() {
		e := expected.result.String()
		a := actual.result.String()
		if e != a {
			return assert.Failf(t, "failed NaN test", "test:\n%s\ncalculated result: %v", testValStrings, actual.result)
		}
		return true
	}
	if expected.result.Cmp(actual.result) != 0 {
		return assert.Fail(t, "failed", "test:\n%s\ncalculated result: %v", testValStrings, actual.result)
	}
	return true
}

var textResults = set[string]{"class": {}}

// TODO: get runTest to run more functions such as FMA.
// execOp returns the calculated answer to the operation as Decimal64.
func execOp(context Context64, a, b, c Decimal64, op string) opResult {
	switch op {
	case "add":
		return opResult{result: context.Add(a, b)}
	case "multiply":
		return opResult{result: context.Mul(a, b)}
	case "abs":
		return opResult{result: a.Abs()}
	case "minus":
		return opResult{result: a.Neg()}
	case "divide":
		return opResult{result: context.Quo(a, b)}
	case "fma":
		return opResult{result: context.FMA(a, b, c)}
	case "compare":
		return opResult{result: a.Cmp64(b)}
	case "min":
		return opResult{result: a.Min(b)}
	case "minmag":
		return opResult{result: a.MinMag(b)}
	case "class":
		return opResult{text: a.Class()}
	default:
		fmt.Println("end of execOp, no tests ran", op)
	}
	return opResult{result: Zero64}
}
