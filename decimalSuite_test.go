package decimal

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
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

type set map[string]struct{}

func (s set) Has(k string) bool {
	_, ok := s[k]
	return ok
}

var supportedRounding = set{"half_up": {}, "half_even": {}}
var ignoredFunctions = set{"apply": {}}

// TestFromSuite is the master tester for the dectest suite.
func TestFromSuite(t *testing.T) {
	t.Parallel()

	for _, file := range []string{
		"dectest/ddAbs.decTest",
		"dectest/ddAdd.decTest",
		"dectest/ddClass.decTest",
		"dectest/ddCompare.decTest",
		"dectest/ddCopysign.decTest",
		"dectest/ddDivide.decTest",
		"dectest/ddFMA.decTest",
		"dectest/ddLogB.decTest",
		"dectest/ddMax.decTest",
		"dectest/ddMaxMag.decTest",
		"dectest/ddMin.decTest",
		"dectest/ddMinMag.decTest",
		"dectest/ddMinus.decTest",
		"dectest/ddMultiply.decTest",
		"dectest/ddNextMinus.decTest",
		"dectest/ddNextPlus.decTest",
		"dectest/ddPlus.decTest",
		"dectest/ddSubtract.decTest",

		// Future
		// "dectest/ddBase.decTest",
		// "dectest/ddNextToward.decTest",
		// "dectest/ddRemainder.decTest",
		// "dectest/ddRemainderNear.decTest",
		// "dectest/ddScaleB.decTest",
		// "dectest/ddQuantize.decTest",
		// "dectest/ddToIntegral.decTest",

		// Not planned
		// "dectest/ddAnd.decTest",
		// "dectest/ddCanonical.decTest",
		// "dectest/ddCompareSig.decTest",
		// "dectest/ddCompareTotal.decTest",
		// "dectest/ddCompareTotalMag.decTest",
		// "dectest/ddCopy.decTest",
		// "dectest/ddCopyAbs.decTest",
		// "dectest/ddCopyNegate.decTest",
		// "dectest/ddDivideInt.decTest",
		// "dectest/ddEncode.decTest",
		// "dectest/ddInvert.decTest",
		// "dectest/ddOr.decTest",
		// "dectest/ddReduce.decTest",
		// "dectest/ddRotate.decTest",
		// "dectest/ddSameQuantum.decTest",
		// "dectest/ddShift.decTest",
		// "dectest/ddXor.decTest",
	} {
		file := file
		t.Run(file, func(t *testing.T) {
			t.Parallel()

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
	i := 0
	for ; i < len(fields); i++ {
		if fields[i] == "->" {
			break
		}
	}
	if i == len(fields) {
		panic("missing ->")
	}
	if i < 5 {
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

var textResults = set{"class": {}}

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
	case "compare":
		return opResult{result: a.Cmp64(b)}
	case "copysign":
		return opResult{result: a.CopySign(b)}
	case "divide":
		return opResult{result: context.Quo(a, b)}
	case "fma":
		return opResult{result: context.FMA(a, b, c)}
	case "logb":
		return opResult{result: a.Logb()}
	case "max":
		return opResult{result: a.Max(b)}
	case "maxmag":
		return opResult{result: a.MaxMag(b)}
	case "min":
		return opResult{result: a.Min(b)}
	case "minmag":
		return opResult{result: a.MinMag(b)}
	case "minus":
		return opResult{result: a.Neg()}
	case "nextminus":
		return opResult{result: a.NextMinus()}
	case "nextplus":
		return opResult{result: a.NextPlus()}
	case "plus":
		return opResult{result: a}
	case "subtract":
		return opResult{result: context.Add(a, b.Neg())}
	case "class":
		return opResult{text: a.Class()}
	default:
		panic(fmt.Errorf("unhandled op: %s", op))
	}
}
