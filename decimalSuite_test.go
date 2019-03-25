package decimal

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"testing"
)

type decValContainer struct {
	val1, val2, expected Decimal64
	parseError           error
}
type testCaseStrings struct {
	testName       string
	testFunc       string
	val1           string
	val2           string
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

// TODO: any tests that are failing a particular test in the test suite will be turned into a unit test.
// func TestNew(t *testing.T) {
// require := require.New(t)
// propper rounding
// require.Equal(MustParseDecimal64("4444444444444445"), MustParseDecimal64("4444444444444444").Add(MustParseDecimal64("0.5001")))
// require.Equal(MustParseDecimal64("0.23"), MustParseDecimal64("1.3").Add(MustParseDecimal64("-1.07")))
// fmt.Println("sjoidgf", MustParseDecimal64("4444444444444444").Add(MustParseDecimal64("1.5001")))
// require.Equal(MustParseDecimal64("12345678901234.29"), MustParseDecimal64("12345678901234").Add(MustParseDecimal64("0.2951")))
//
// }

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
		`(?P<val2>\+?-?[^->]?[^\t\f\v\' ]*)` + //testvals2 same as 1 but specifically dont match with '->'
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
			expectedResult: a[5],
		})
	}
	return

}

// convertToDec64 converts the map object strings to decimal64s.
func convertToDec64(testvals testCaseStrings) (dec64vals decValContainer) {
	var err1, err2, expectedErr error
	dec64vals.val1, err1 = ParseDecimal64(testvals.val1)
	dec64vals.val2, err2 = ParseDecimal64(testvals.val2)
	dec64vals.expected, expectedErr = ParseDecimal64(testvals.expectedResult)

	if err1 != nil || err2 != nil || expectedErr != nil {
		dec64vals.parseError = fmt.Errorf("\nerror parsing in test: %s: \n val 1:%s: \n val 2: %s\n expected: %s ",
			testvals.testName,
			err1,
			err2,
			expectedErr)
	}
	return
}

// runTest completes the tests and returns a boolean and string on if the test passes.
func runTest(testVals decValContainer, testValStrings testCaseStrings) error {
	calcRestul := execOp(testVals.val1, testVals.val2, testValStrings.testFunc)
	flavor1, _, _, _ := calcRestul.parts()
	flavor2, _, _, _ := testVals.expected.parts()
	if flavor1 == flSNaN || flavor2 == flSNaN {
		if testVals.expected.Cmp(calcRestul) == -2 {
			return nil
		}

		return fmt.Errorf(
			"\nfailed NaN %s \n %v %s %v == %v \n expected result: %v ",
			testValStrings.testName,
			testValStrings.val1,
			testValStrings.testFunc,
			testValStrings.val2,
			calcRestul,
			testVals.expected)

	} else if testVals.expected.Cmp(calcRestul) != 0 {
		return fmt.Errorf(
			"\nfailed %s \n %v %s %v == %v \n expected result: %v ",
			testValStrings.testName,
			testValStrings.val1,
			testValStrings.testFunc,
			testValStrings.val2,
			calcRestul,
			testVals.expected)
	}
	return nil
}

// TODO: get runTest to run more functions such as FMA.
// execOp returns the calculated answer to the operation as Decimal64.
func execOp(val1, val2 Decimal64, op string) Decimal64 {
	switch op {
	case "add":
		return val1.Add(val2)
	case "multiply":
		return val1.Mul(val2)
	case "abs":
		return val1.Abs()
	case "divide":
		return val1.Quo(val2)
	default:
		panic("end of operation function no tests ran")
	}
}
