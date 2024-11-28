package decimal

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
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

var (
	supportedRounding = set{"half_up": {}, "half_even": {}}
	ignoredFunctions  = set{"apply": {}}
	excludedTests     = set{
		// ddintx074 and ddintx094 expect a specific bit pattern that doesn't
		// seem to make sense
		"ddintx074": {}, "ddintx094": {},
	}
)

// TestFromSuite is the master tester for the dectest suite.
func TestFromSuite(t *testing.T) {
	t.Parallel()

	test := func(file string) func(t *testing.T) {
		return func(t *testing.T) {
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
						isnil(t, err)
						if !runTest(t, scannedContext, dec64vals, testVal) {
							runTest(t, scannedContext, dec64vals, testVal)
						}
					})
				}
			}
		}
	}

	t.Run("ddAbs", test("dectest/ddAbs.decTest"))
	t.Run("ddAdd", test("dectest/ddAdd.decTest"))
	t.Run("ddClass", test("dectest/ddClass.decTest"))
	t.Run("ddCompare", test("dectest/ddCompare.decTest"))
	t.Run("ddCopySign", test("dectest/ddCopySign.decTest"))
	t.Run("ddDivide", test("dectest/ddDivide.decTest"))
	t.Run("ddFMA", test("dectest/ddFMA.decTest"))
	t.Run("ddLogB", test("dectest/ddLogB.decTest"))
	t.Run("ddMax", test("dectest/ddMax.decTest"))
	t.Run("ddMaxMag", test("dectest/ddMaxMag.decTest"))
	t.Run("ddMin", test("dectest/ddMin.decTest"))
	t.Run("ddMinMag", test("dectest/ddMinMag.decTest"))
	t.Run("ddMinus", test("dectest/ddMinus.decTest"))
	t.Run("ddMultiply", test("dectest/ddMultiply.decTest"))
	t.Run("ddNextMinus", test("dectest/ddNextMinus.decTest"))
	t.Run("ddNextPlus", test("dectest/ddNextPlus.decTest"))
	t.Run("ddPlus", test("dectest/ddPlus.decTest"))
	t.Run("ddRound", test("dectest/ddRound.decTest"))
	t.Run("ddScaleB", test("dectest/ddScaleB.decTest"))
	t.Run("ddSubtract", test("dectest/ddSubtract.decTest"))
	t.Run("ddToIntegral", test("dectest/ddToIntegral.decTest"))
	t.Run("squareroot", test("dectest/squareroot.decTest"))

	// Future
	// t.Run("ddBase", test("dectest/ddBase.decTest"))
	// t.Run("ddCompareTotal", test("dectest/ddCompareTotal.decTest"))
	// t.Run("ddCompareTotalMag", test("dectest/ddCompareTotalMag.decTest"))
	// t.Run("ddCopyAbs.decTest", //", test("dectest/ddCopyAbs.decTest", // QAb)s)
	// t.Run("ddCopyNegate.decTest", //", test("dectest/ddCopyNegate.decTest", // QNe)g)
	// t.Run("ddDivideInt", test("dectest/ddDivideInt.decTest"))
	// t.Run("ddNextToward", test("dectest/ddNextToward.decTest"))
	// t.Run("ddRemainder", test("dectest/ddRemainder.decTest"))
	// t.Run("ddRemainderNear", test("dectest/ddRemainderNear.decTest"))

	// Wat?
	// t.Run("ddEncode", test("dectest/ddEncode.decTest"))

	// Not planned
	// -- bitwise
	// t.Run("ddAnd", test("dectest/ddAnd.decTest"))
	// t.Run("ddInvert", test("dectest/ddInvert.decTest"))
	// t.Run("ddOr", test("dectest/ddOr.decTest"))
	// t.Run("ddRotate", test("dectest/ddRotate.decTest"))
	// t.Run("ddShift", test("dectest/ddShift.decTest"))
	// t.Run("ddXor", test("dectest/ddXor.decTest"))
	//
	// -- signalling
	// t.Run("ddCompareSig", test("dectest/ddCompareSig.decTest"))
	//
	// -- nop
	// t.Run("ddCopy", test("dectest/ddCopy.decTest"))
	//
	// -- repr
	// t.Run("ddCanonical", test("dectest/ddCanonical.decTest"))
	// t.Run("ddQuantize", test("dectest/ddQuantize.decTest"))
	// t.Run("ddReduce", test("dectest/ddReduce.decTest"))
	// t.Run("ddSameQuantum", test("dectest/ddSameQuantum.decTest"))

}

func setRoundingFromString(s string) Context64 {
	switch s {
	case "half_even":
		return Context64{HalfEven}
	case "half_up":
		return Context64{HalfUp}
	case "default":
		return DefaultContext64
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
	if m == nil || !strings.HasPrefix(m[0][2], "dd") && !strings.HasPrefix(m[0][2], "sqtx") {
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
	if excludedTests.Has(test.name) {
		return nil
	}
	if ignoredFunctions.Has(test.function) {
		return nil
	}

	// # represents a null value, which isn't meaningful for Decimal64.
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
		if hexBits, cut := strings.CutPrefix(s, "#"); cut {
			bits, err := strconv.ParseUint(hexBits, 16, 64)
			if err != nil {
				return Decimal64{}, err
			}
			return new64(bits), nil
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
func runTest(t *testing.T, context Context64, expected opResult, testValStrings *testCase) pass {
	return replayOnFail(t, func() {
		actual := execOp(context, expected.val1, expected.val2, expected.val3, testValStrings.function)
		switch {
		case actual.text != "":
			if testValStrings.function == "compare" && actual.text == "-2" && expected.result.IsNaN() {
				return
			}
			if actual.text != testValStrings.expectedResult {
				t.Errorf("test:\n%s\ncalculated text: %s", testValStrings, actual.text)
			}
		case actual.result.IsNaN() || expected.result.IsNaN():
			e := expected.result.String()
			a := actual.result.String()
			if e != a {
				t.Errorf("test:\n%s\ncalculated result: %v", testValStrings, actual.result)
			}
		case expected.result.Cmp(actual.result) != 0:
			t.Errorf("test:\n%s\ncalculated result: %v", testValStrings, actual.result)
		}
	})
}

var textResults = set{"class": {}}

var ops = map[string]func(ctx Context64, a, b, c Decimal64) any{
	"add":         func(ctx Context64, a, b, c Decimal64) any { return ctx.Add(a, b) },
	"abs":         func(ctx Context64, a, b, c Decimal64) any { return a.Abs() },
	"class":       func(ctx Context64, a, b, c Decimal64) any { return a.Class() },
	"compare":     func(ctx Context64, a, b, c Decimal64) any { return a.Cmp64(b) },
	"copysign":    func(ctx Context64, a, b, c Decimal64) any { return a.CopySign(b) },
	"divide":      func(ctx Context64, a, b, c Decimal64) any { return ctx.Quo(a, b) },
	"fma":         func(ctx Context64, a, b, c Decimal64) any { return ctx.FMA(a, b, c) },
	"logb":        func(ctx Context64, a, b, c Decimal64) any { return a.Logb() },
	"max":         func(ctx Context64, a, b, c Decimal64) any { return a.Max(b) },
	"maxmag":      func(ctx Context64, a, b, c Decimal64) any { return a.MaxMag(b) },
	"min":         func(ctx Context64, a, b, c Decimal64) any { return a.Min(b) },
	"minmag":      func(ctx Context64, a, b, c Decimal64) any { return a.MinMag(b) },
	"minus":       func(ctx Context64, a, b, c Decimal64) any { return a.Neg() },
	"multiply":    func(ctx Context64, a, b, c Decimal64) any { return ctx.Mul(a, b) },
	"nextminus":   func(ctx Context64, a, b, c Decimal64) any { return a.NextMinus() },
	"nextplus":    func(ctx Context64, a, b, c Decimal64) any { return a.NextPlus() },
	"plus":        func(ctx Context64, a, b, c Decimal64) any { return a },
	"scaleb":      func(ctx Context64, a, b, c Decimal64) any { return a.ScaleB(b) },
	"round":       func(ctx Context64, a, b, c Decimal64) any { return ctx.Round(a, b) },
	"tointegralx": func(ctx Context64, a, b, c Decimal64) any { return ctx.ToIntegral(a) },
	"subtract":    func(ctx Context64, a, b, c Decimal64) any { return ctx.Add(a, b.Neg()) },
	"squareroot":  func(ctx Context64, a, b, c Decimal64) any { return a.Sqrt() },
	// "quantize":    func(ctx Context64, a, b, c Decimal64) any { return ctx.Quantize(a, b) },
}

// TODO: get runTest to run more functions such as FMA.
// execOp returns the calculated answer to the operation as Decimal64.
func execOp(ctx Context64, a, b, c Decimal64, op string) opResult {
	if f, has := ops[op]; has {
		switch a := f(ctx, a, b, c).(type) {
		case string:
			return opResult{text: a}
		case Decimal64:
			return opResult{result: a}
		default:
			panic("wat?")
		}
	}
	panic(fmt.Errorf("unhandled op: %s", op))
}
