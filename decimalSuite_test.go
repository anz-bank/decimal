package decimal

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"testing"
)

type decValContainer struct {
	val1, val2, val3, expected Decimal64
	parseError                 error
}
type testCaseStrings struct {
	testName       string
	testFunc       string
	val1           string
	val2           string
	val3           string
	expectedResult string
	rounding       string
}

const PrintFiles bool = true
const PrintTests bool = false
const RunTests bool = true
const IgnorePanics bool = false
const IgnoreRounding bool = false

var tests = []string{"",
	"dectest/ddAdd.decTest",
	"dectest/ddMultiply.decTest",
	"dectest/ddFMA.decTest",
	// TODO: Implement following tests
	// "dectest/ddCompare.decTest",
	// 	"dectest/ddAbs.decTest",
	// 	"dectest/ddClass.decTest",
	// 	"dectest/ddCopysign.decTest",
	// 	"dectest/ddDivide.decTest",
	// 	"dectest/ddLogB.decTest",
	// 	"dectest/ddMin.decTest",
	// 	"dectest/ddMinMag.decTest",
	// 	"dectest/ddMinus.decTest",
}
var supportedRounding = []string{"half_up",
	"half_even"}
var ignoredFunctions = []string{"apply"}

// TODO(joshcarp): This test cannot fail. Proper assertions will be added once the whole suite passes
// TestFromSuite is the master tester for the dectest suite.
func TestFromSuite(t *testing.T) {
	if RunTests {
		for _, file := range tests {
			if PrintFiles {
				fmt.Println("starting test:", file)
			}
			f, _ := os.Open(file)
			scanner := bufio.NewScanner(f)
			numTests := 0
			failedTests := 0
			var roundingSupported bool
			var scannedContext Context64

			for scanner.Scan() {
				testVal, err := getInput(scanner.Text())
				if testVal.rounding != "" {
					roundingSupported = isInList(testVal.rounding, supportedRounding)
					scannedContext = setRoundingFromString(testVal.rounding)
				}
				if err == nil && roundingSupported {
					numTests++
					dec64vals := convertToDec64(testVal)
					calcRestul, testErr := runTest(scannedContext, dec64vals, testVal)
					if PrintTests {
						fmt.Printf("%s %s %v %v %v -> %v\n", testVal.testName, testVal.testFunc, testVal.val1, testVal.val2, testVal.val3, testVal.expectedResult)
					}
					if testErr != nil && !(isRoundingErr(calcRestul, dec64vals.expected) && IgnoreRounding) {
						fmt.Println(testErr)
						fmt.Println("Rounding mode:", supportedRounding[scannedContext.roundingMode])
						failedTests++
						fmt.Printf("%s %s %v %v %v -> %v\n", testVal.testName, testVal.testFunc, testVal.val1, testVal.val2, testVal.val3, testVal.expectedResult)
						if dec64vals.parseError != nil {
							fmt.Println(dec64vals.parseError)
						}
					}
				}
			}
			if PrintFiles {
				fmt.Println("Number of tests ran:", numTests, "Number of failed tests:", failedTests)
			}
		}
		fmt.Printf("decimalSuite_test settings (These should only be true for debug):\n Ignore Rounding errors: %v\n Ignore Panics: %v\n", IgnoreRounding, IgnorePanics)
	}
}

func isInList(s string, list []string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}

func setRoundingFromString(s string) Context64 {
	switch s {
	case "half_even":
		return Context64{roundHalfEven}
	case "half_up":
		return Context64{roundHalfUp}
	default:
		return DefaultContext
	}

}
func isRoundingErr(res, expected Decimal64) bool {
	resP := res.getParts()
	expectedP := expected.getParts()
	sigDiff := int64(resP.significand - expectedP.significand)
	expDiff := resP.exp - expectedP.exp
	if (sigDiff == 1 || sigDiff == -1) && (expDiff == 1 || expDiff == -1 || expDiff == 0) {
		return true
	}
	if resP.significand == maxSig && resP.exp == expMax && expectedP.fl == flInf {
		return true
	}
	return false
}

// getInput gets the test file and extracts test using regex, then returns a map object and a list of test names.
func getInput(line string) (testCaseStrings, error) {
	testRegex := regexp.MustCompile(
		`(?P<testName>dd[\w]*)` + // first capturing group: testfunc made of anything that isn't a whitespace
			`(?:\s*)` + // match any whitespace (?: non capturing group)
			`(?P<testFunc>[\S]*)` + // testfunc made of anything that isn't a whitespace
			`(?:\s*\'?)` + // after can be any number of spaces and quotations if they exist (?: non capturing group)
			`(?P<val1>\+?-?[^\t\f\v\' ]*)` + // first test val is anything that isnt a whitespace or a quoteation mark
			`(?:'?\s*'?)` + // match any quotation marks and any space (?: non capturing group)
			`(?P<val2>\+?-?[^\t\f\v\' ]*)` + // second test val is anything that isnt a whitespace or a quoteation mark
			`(?:'?\s*'?)` +
			`(?P<val3>\+?-?[^->]?[^\t\f\v\' ]*)` + //testvals3 same as 1 but specifically dont match with '->'
			`(?:'?\s*->\s*'?)` + // matches the indicator to answer and surrounding whitespaces (?: non capturing group)
			`(?P<expectedResult>\+?-?[^\r\n\t\f\v\' ]*)`) // matches the answer that's anything that is plus minus but not quotations

	// Add regex to match to  rounding: rounding mode her

	// capturing gorups are testName, testFunc, val1,  val2, and expectedResult)
	ans := testRegex.FindStringSubmatch(line)

	if len(ans) == 0 {
		roundingRegex := regexp.MustCompile(`(?:rounding:[\s]*)(?P<rounding>[\S]*)`)
		ans = roundingRegex.FindStringSubmatch(line)
		if len(ans) == 0 {
			return testCaseStrings{}, fmt.Errorf("No test cases")
		}
		return testCaseStrings{rounding: ans[1]}, fmt.Errorf("No test cases or rounding")
	}
	if isInList(ans[2], ignoredFunctions) {
		return testCaseStrings{}, fmt.Errorf("No test cases")
	}
	data := testCaseStrings{
		testName:       ans[1],
		testFunc:       ans[2],
		val1:           ans[3],
		val2:           ans[4],
		val3:           ans[5],
		expectedResult: ans[6],
	}
	return data, nil
}

// convertToDec64 converts the map object strings to decimal64s.
func convertToDec64(testvals testCaseStrings) (dec64vals decValContainer) {
	var err1, err2, err3, expectedErr error
	dec64vals.val1, err1 = ParseDecimal64(testvals.val1)
	dec64vals.val2, err2 = ParseDecimal64(testvals.val2)
	dec64vals.val3, err3 = ParseDecimal64(testvals.val3)
	dec64vals.expected, expectedErr = ParseDecimal64(testvals.expectedResult)

	if err1 != nil || err2 != nil || expectedErr != nil {
		dec64vals.parseError = fmt.Errorf("\nerror parsing in test: %s: \n val 1:%s: \n val 2: %s  \n val 3: %s\n expected: %s ",
			testvals.testName,
			err1,
			err2,
			err3,
			expectedErr)
	}
	return
}

// runTest completes the tests and returns a boolean and string on if the test passes.
func runTest(context Context64, testVals decValContainer, testValStrings testCaseStrings) (Decimal64, error) {
	calcRestul := execOp(context, testVals, testValStrings.testFunc)
	if calcRestul.IsNaN() || testVals.expected.IsNaN() {
		if testVals.expected.String() != calcRestul.String() {
			return calcRestul, fmt.Errorf(
				"\n failed NaN TEST %s \n %v %s %v %v== %v \n expected result: %v ",
				testValStrings.testName,
				testValStrings.val1,
				testValStrings.testFunc,
				testValStrings.val2,
				testValStrings.val3,
				calcRestul,
				testValStrings.expectedResult)
		}
		return calcRestul, nil
	} else if testVals.expected.Cmp(calcRestul) != 0 {
		return calcRestul, fmt.Errorf(
			"\nfailed %s \n %v %s %v %v== %v \n expected result: %v ",
			testValStrings.testName,
			testValStrings.val1,
			testValStrings.testFunc,
			testValStrings.val2,
			testValStrings.val3,
			calcRestul,
			testValStrings.expectedResult)
	}
	return calcRestul, nil
}

// TODO: get runTest to run more functions such as FMA.
// execOp returns the calculated answer to the operation as Decimal64.
func execOp(context Context64, testVals decValContainer, op string) Decimal64 {
	if IgnorePanics {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("failed", r, testVals.val1, testVals.val2)
			}
		}()
	}
	switch op {
	case "add":
		return context.Add(testVals.val1, testVals.val2)
	case "multiply":
		return context.Mul(testVals.val1, testVals.val2)
	case "abs":
		return testVals.val1.Abs()
	case "divide":
		return testVals.val1.Quo(testVals.val2)
	case "fma":
		return context.FMA(testVals.val1, testVals.val2, testVals.val3)
	case "compare":
		return NewDecimal64FromInt64(int64(testVals.val1.Cmp(testVals.val2)))
	default:
		fmt.Println("end of execOp, no tests ran", op)
	}
	return Zero64
}
