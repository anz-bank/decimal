package decimal

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"testing"
)

type decValContainer struct {
	val1           Decimal64
	val1Err        error
	val2           Decimal64
	val2Err        error
	expectedResult Decimal64
	expecValErr    error
}
type testCaseStrings struct {
	name           string
	op             string
	val1           string
	val2           string
	expectedResult string
}

// master tester for the suite, currently uses print statements but will implement assert statements as test cases pass
func TestFromSuite(t *testing.T) {
	// require := require.New(t)
	testVals := getInput("dectest/ddAdd.decTest")
	var TestResult string
	var TestPass bool
	for i := range testVals {
		Dec64Vals := convertToDec64(testVals[i])
		TestPass, TestResult = doTests(Dec64Vals, testVals[i]) // do more here

		if TestPass == false {
			fmt.Printf("\n%s: \n %s\n", testVals[i].name, TestResult)
		}
		if Dec64Vals.val1Err != nil || Dec64Vals.val2Err != nil || Dec64Vals.expecValErr != nil {
			fmt.Printf("\nError parsing in test: %s: \n val 1:%s: \n Val2: %s\n val 3: %s \n", testVals[i].name, Dec64Vals.val1Err, Dec64Vals.val2Err, Dec64Vals.expecValErr)

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

//getInput gets the test file and extracts test using regex, then returns a map object and a list of test names
func getInput(file string) (data []testCaseStrings) {
	dat, _ := ioutil.ReadFile("dectest/ddAdd.decTest")
	dataString := string(dat)
	//  "(\n)") // start with newline
	//  "(?P<TestName>dd[\w]*)") // TestName must start with dd (double decimal)
	//  "(?P<TestFunc>[\S]*)") // testfunc made of anything that isn't a whitespace
	//  "(\s*\'?)") // after can be any number of spaces and quotations if they exist
	// "(?P<TestVals_1>(\+|-)*[^\t\f\v\' ]*)") // first test val is anything that isnt a whitespace or a quoteation mark
	//  "(?P<TestVals_1>(\+|-)*[^\t\f\v\' ]*)") // first test val is anything that isnt a whitespace or a quoteation mark
	// ('?\s*'?) match any quotation marks and any spaces
	// (?P<TestVals_2>(\+|-[^->])?[^\t\f\v\' ]*) testvals2 same as 1 but no '->'
	// ('?\s*->\s*'?) matches the indicator to answer
	// (?P<answer>(\+|-)*[^\t\f\v\' ]*) matches the answer that's anything that is plus minus but not quotations
	// split into groups: testName, TestFunct, TestVals (x2) and TestAns
	regex := `(\n)(?P<TestName>dd[\w]*)(\s*)(?P<TestFunc>[\S]*)(\s*\'?)(?P<TestVals_1>(\+|-)*[^\t\f\v\' ]*)('?\s*'?)(?P<TestVals_2>(\+|-[^->])?[^\t\f\v\' ]*)('?\s*->\s*'?)(?P<answer>(\+|-)*[^\r\n\t\f\v\' ]*)`
	r := regexp.MustCompile(regex)
	ans := r.FindAllStringSubmatch(dataString, -1)
	var datum testCaseStrings
	for i := range ans {
		datum.op = ans[i][4]
		datum.val1 = ans[i][6]
		datum.val2 = ans[i][9]
		datum.expectedResult = ans[i][12]
		datum.name = ans[i][2]
		data = append(data, datum)

	}
	return data

}

//convertToDec64 converts the map object strings to decimal64s
func convertToDec64(testvals testCaseStrings) (dec64Vals decValContainer) {
	dec64Vals.val1, dec64Vals.val1Err = ParseDecimal64(testvals.val1)
	dec64Vals.expectedResult, dec64Vals.expecValErr = ParseDecimal64(testvals.expectedResult)
	dec64Vals.val2, dec64Vals.val2Err = ParseDecimal64(testvals.val2)
	return
}

// TODO: get doTests to run more functions
//doTests completes the tests and returns a boolean and string on if the test passes
func doTests(testVals decValContainer, testValStrings testCaseStrings) (testStatus bool, testString string) {
	switch testValStrings.op {
	case "add":
		testString = fmt.Sprintf("%v + %v != %v \n(expected %v)", testValStrings.val1, testValStrings.val2, testVals.val1.Add(testVals.val2), testVals.expectedResult)
		testStatus = testVals.expectedResult == testVals.val1.Add(testVals.val2)
		return
	case "abs":

		// testStatus = testAns == testVal1.Abs()
		return
	case "divide":

		// testStatus = testAns == testVal1.Quo(testVal2)
		return
	case "minus":

		// testStatus = testAns == testVal1.Sub(testVal2)
		return
	case "multiply":
		// testString = fmt.Sprintf("%v * %v != %v (expected %v)", testVal1, testVal2, testVal1.Mul(testVal2), testAns)
		// testStatus = testAns == testVal1.Mul(testVal2)
		return
	default:
		testStatus = false
		return
	}

}
