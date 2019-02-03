package decimal

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"testing"
)

type parseResult struct {
	d   Decimal64
	err error
}
type decValContainer struct {
	val1, val2, expected parseResult
}
type testCaseStrings struct {
	name           string
	op             string
	val1           string
	val2           string
	expectedResult string
}

// //TODO(joshcarp): This test cannot fail. Proper assertions will be added once the whole suite passes
// TestFromSuite is the master tester for the dectest suite
func TestFromSuite(t *testing.T) {
	// require := require.New(t)
	testVals := getInput("dectest/ddAdd.decTest")
	var TestResult string
	var TestFailed bool
	for i := range testVals {
		Dec64Vals := convertToDec64(testVals[i])
		TestFailed, TestResult = doTest(Dec64Vals, testVals[i]) // do more here

		if TestFailed {
			fmt.Printf("\n%s: \n %s\n", testVals[i].name, TestResult)
		}
		if Dec64Vals.val1.err != nil || Dec64Vals.val2.err != nil || Dec64Vals.expected.err != nil {
			fmt.Printf("\nError parsing in test: %s: \n val 1:%s: \n Val2: %s\n val 3: %s \n", testVals[i].name, Dec64Vals.val1.err, Dec64Vals.val1.err, Dec64Vals.expected.err)

		}
	}

}

// TODO: any tests that are failing a particular test in the test suite will be turned into a unit test
// func TestNew(t *testing.T) {
// 	// require := require.New(t)
// 	// // propper rounding
// 	// require.Equal(MustParseDecimal64("4444444444444445"), MustParseDecimal64("4444444444444444").Add(MustParseDecimal64("0.5001")))
// 	// require.Equal(MustParseDecimal64("0.23"), MustParseDecimal64("1.3").Add(MustParseDecimal64("-1.07")))
// 	// fmt.Println("sjoidgf", MustParseDecimal64("4444444444444444").Add(MustParseDecimal64("1.5001")))
// 	// // require.Equal(MustParseDecimal64("12345678901234.29"), MustParseDecimal64("12345678901234").Add(MustParseDecimal64("0.2951")))
//
// }

// getInput gets the test file and extracts test using regex, then returns a map object and a list of test names
func getInput(file string) (data []testCaseStrings) {
	dat, _ := ioutil.ReadFile("dectest/ddAdd.decTest")
	dataString := string(dat)
	r := regexp.MustCompile((`(?:\n)` + // start with newline (?: non capturing group)
		`(?P<name>dd[\w]*)` + // first capturing group: testfunc made of anything that isn't a whitespace
		`(?:\s*)` + // match any whitespace (?: non capturing group)
		`(?P<op>[\S]*)` + // testfunc made of anything that isn't a whitespace
		`(?:\s*\'?)` + // after can be any number of spaces and quotations if they exist (?: non capturing group)
		`(?P<val1>\+?-?[^\t\f\v\' ]*)` + // first test val is anything that isnt a whitespace or a quoteation mark
		`(?:'?\s*'?)` + // match any quotation marks and any space (?: non capturing group)
		`(?P<val2>\+?-?[^->]?[^\t\f\v\' ]*)` + //testvals2 same as 1 but specifically dont match with '->'
		`(?:'?\s*->\s*'?)` + // matches the indicator to answer and surrounding whitespaces (?: non capturing group)
		`(?P<expectedResult>\+?-?[^\r\n\t\f\v\' ]*)`)) // matches the answer that's anything that is plus minus but not quotations
	// capturing gorups are testName, TestFunct, TestVals_1,  TestVals_2, and answer)

	ans := r.FindAllStringSubmatch(dataString, -1)
	for _, a := range ans {
		data = append(data, testCaseStrings{
			name:           a[1],
			op:             a[2],
			val1:           a[3],
			val2:           a[4],
			expectedResult: a[5]})

	}
	return

}

// convertToDec64 converts the map object strings to decimal64s
func convertToDec64(testvals testCaseStrings) (dec64Vals decValContainer) {
	dec64Vals.val1.d, dec64Vals.val1.err = ParseDecimal64(testvals.val1)
	dec64Vals.expected.d, dec64Vals.expected.err = ParseDecimal64(testvals.expectedResult)
	dec64Vals.val2.d, dec64Vals.val2.err = ParseDecimal64(testvals.val2)
	return
}

// doTest completes the tests and returns a boolean and string on if the test passes
func doTest(testVals decValContainer, testValStrings testCaseStrings) (testFailed bool, testString string) {
	switch testValStrings.op {
	case "add":
		testString = fmt.Sprintf("%v + %v != %v \n(expected %v)", testValStrings.val1, testValStrings.val2, testVals.val1.d.Add(testVals.val2.d), testVals.expected.d)
		if testVals.expected.d.Cmp(testVals.val1.d.Add(testVals.val2.d)) != 0 {
			return true, testString
		}

	// TODO: get doTest to run more functions
	// 	return
	// case "abs":
	//
	// 	// testStatus = testAns == testVal1.Abs()
	// 	return
	// case "divide":
	//
	// 	// testStatus = testAns == testVal1.Quo(testVal2)
	// 	return
	// case "minus":
	//
	// 	// testStatus = testAns == testVal1.Sub(testVal2)
	// 	return
	// case "multiply":
	// 	// testString = fmt.Sprintf("%v * %v != %v (expected %v)", testVal1, testVal2, testVal1.Mul(testVal2), testAns)
	// 	// testStatus = testAns == testVal1.Mul(testVal2)
	// 	return
	default:
		testFailed = true
		return
	}
	return
}
