package decimal

import (
	"fmt"
	"io/ioutil"
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
}

const TESTDEBUG bool = true
const RUNSUITES bool = true

var tests = []string{"dectest/ddAdd.decTest"}

// "dectest/ddFMA.decTest",
// "dectest/ddMultiply.decTest"}

// TODO: Implement following tests
// "dectest/ddCompare.decTest"}
// 	"dectest/ddAbs.decTest",
// 	"dectest/ddClass.decTest",
// 	"dectest/ddCopysign.decTest",
// 	"dectest/ddDivide.decTest",
// 	"dectest/ddLogB.decTest",
// 	"dectest/ddMin.decTest",
// 	"dectest/ddMinMag.decTest",
// 	"dectest/ddMinus.decTest",
// }

// TODO(joshcarp): This test cannot fail. Proper assertions will be added once the whole suite passes
// TestFromSuite is the master tester for the dectest suite.
func TestFromSuite(t *testing.T) {
	if RUNSUITES {
		for _, file := range tests {
			if TESTDEBUG {
				fmt.Println("starting test:", file)
			}
			testVals := getInput(file)
			for _, testVal := range testVals {
				dec64vals := convertToDec64(testVal)
				testErr := runTest(dec64vals, testVal)
				// fmt.Println("running test", testVal.testName)
				if testErr != nil {
					fmt.Println(testErr)
					if dec64vals.parseError != nil {
						fmt.Println(dec64vals.parseError)
					}
				}
			}
		}
	}
}

// TODO get regexto match with three inputs for functions like FMA.
// getInput gets the test file and extracts test using regex, then returns a map object and a list of test names.
func getInput(file string) (data []testCaseStrings) {
	dat, _ := ioutil.ReadFile(file)
	dataString := string(dat)
	r := regexp.MustCompile(`(?:\n)` + // start with newline (?: non capturing group)
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
	// capturing gorups are testName, testFunc, val1,  val2, and expectedResult)
	ans := r.FindAllStringSubmatch(dataString, -1)
	for _, a := range ans {
		data = append(data, testCaseStrings{
			testName:       a[1],
			testFunc:       a[2],
			val1:           a[3],
			val2:           a[4],
			val3:           a[5],
			expectedResult: a[6],
		})
	}
	return
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
func runTest(testVals decValContainer, testValStrings testCaseStrings) error {
	calcRestul := execOp(testVals.val1, testVals.val2, testVals.val3, testValStrings.testFunc)
	if calcRestul.IsNaN() || testVals.expected.IsNaN() {
		if testVals.expected.String() != calcRestul.String() {
			return fmt.Errorf(
				"\n failed NaN TEST %s \n %v %s %v == %v \n expected result: %v ",
				testValStrings.testName,
				testValStrings.val1,
				testValStrings.testFunc,
				testValStrings.val2,
				calcRestul,
				testValStrings.expectedResult)
		}
		return nil

	} else if testVals.expected.Cmp(calcRestul) != 0 {
		return fmt.Errorf(
			"\nfailed %s \n %v %s %v == %v \n expected result: %v ",
			testValStrings.testName,
			testValStrings.val1,
			testValStrings.testFunc,
			testValStrings.val2,
			calcRestul,
			testValStrings.expectedResult)
	}
	return nil
}

// TODO: get runTest to run more functions such as FMA.
// execOp returns the calculated answer to the operation as Decimal64.
func execOp(val1, val2, val3 Decimal64, op string) Decimal64 {
	switch op {
	case "add":
		return val1.Add(val2)
	case "multiply":
		return val1.Mul(val2)
	case "abs":
		return val1.Abs()
	case "divide":
		return val1.Quo(val2)
	case "fma": // TODO: Add FMA function
		//return val1.FMA(val2, val3)
	default:
		fmt.Println("end of execOp, no tests ran", op)
	}
	return Zero64
}
