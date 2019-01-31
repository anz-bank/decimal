package decimal

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"testing"
)

func getInput(file string) (map[string]map[string]string, []string) {
	dat, _ := ioutil.ReadFile("dectest/ddAdd.decTest")
	dataString := string(dat)
	//  "(\n|\r)") // start with newline
	//  "(?P<TestName>dd[\w]*)") // TestName must start with dd (double decimal)
	//  "(?P<TestFunc>[\S]*)") // testfunc made of anything that isn't a whitespace
	//  "(\s*\'?)") // after can be any number of spaces and quotations if they exist
	// "(?P<TestVals_1>(\+|-)*[^\r\n\t\f\v\' ]*)") // first test val is anything that isnt a whitespace or a quoteation mark
	//  "(?P<TestVals_1>(\+|-)*[^\r\n\t\f\v\' ]*)") // first test val is anything that isnt a whitespace or a quoteation mark
	// ('?\s*'?) match any quotation marks and any spaces
	// (?P<TestVals_2>(\+|-[^->])?[^\r\n\t\f\v\' ]*) testvals2 same as 1 but no '->'
	// ('?\s*->\s*'?) matches the indicator to answer
	// (?P<answer>(\+|-)*[^\r\n\t\f\v\' ]*) matches the answer that's anything that is plus minus but not quotations

	// split into groups: testName, TestFunct, TestVals (x2) and TestAns
	regex := `(\n|\r)(?P<TestName>dd[\w]*)(\s*)(?P<TestFunc>[\S]*)(\s*\'?)(?P<TestVals_1>(\+|-)*[^\r\n\t\f\v\' ]*)('?\s*'?)(?P<TestVals_2>(\+|-[^->])?[^\r\n\t\f\v\' ]*)('?\s*->\s*'?)(?P<answer>(\+|-)*[^\r\n\t\f\v\' ]*)`
	r := regexp.MustCompile(regex)
	ans := r.FindAllStringSubmatch(dataString, -1)
	testList := []string{}
	var data = map[string]map[string]string{}

	for i := 0; i < len(ans); i++ {
		data[ans[i][2]] = make(map[string]string)
		data[ans[i][2]]["TestFunc"] = ans[i][4]
		data[ans[i][2]]["TestVal1"] = ans[i][6]
		data[ans[i][2]]["TestVal2"] = ans[i][9]
		data[ans[i][2]]["TestAns"] = ans[i][12]
		testList = append(testList, ans[i][2])

	}
	return data, testList

}
func convertToDec64(testVals map[string]string, testName string) (testVal1, testVal2, testAns Decimal64, parseErr string) {
	var err1, err2, err3 error
	// var TestResult []string

	testVal1, err1 = ParseDecimal64(testVals["TestVal1"])
	testAns, err3 = ParseDecimal64(testVals["TestAns"])

	testVal2, err2 = ParseDecimal64(testVals["TestVal2"])
	parseErr = checkErrors(err1, err2, err3, testVals)

	return
}

func checkErrors(err1, err2, err3 error, testVals map[string]string) (errReturn string) {
	var errReturnSlice []string
	if err2 != nil {
		errReturnSlice = append(errReturnSlice, fmt.Sprintf("error parsing input of value: %s \n", testVals["TestVal2"]))
	}
	if err1 != nil {
		errReturnSlice = append(errReturnSlice, fmt.Sprintf("error parsing input of value: %s \n", testVals["TestVal1"]))
	}

	if err3 != nil {
		errReturnSlice = append(errReturnSlice, fmt.Sprintf("error parsing input of value: %s \n", testVals["TestAns"]))
	}
	return strings.Join(errReturnSlice, "")
}

func doTests(testVal1, testVal2, testAns Decimal64, testfunc string) (testStatus bool, testString string) {

	fmt.Println()
	switch testfunc {
	case "add":
		testString = fmt.Sprintf("%v + %v != %v \n(expected %v)", testVal1, testVal2, testVal1.Add(testVal2), testAns)
		testStatus = testAns == testVal1.Add(testVal2)
		return
	case "abs":

		testStatus = testAns == testVal1.Abs()
		return
	case "divide":

		testStatus = testAns == testVal1.Quo(testVal2)
		return
	case "minus":

		testStatus = testAns == testVal1.Sub(testVal2)
		return
	case "multiply":
		testString = fmt.Sprintf("%v * %v != %v (expected %v)", testVal1, testVal2, testVal1.Mul(testVal2), testAns)
		testStatus = testAns == testVal1.Mul(testVal2)
		return
	default:
		testStatus = false
		return
	}

}

func TestFromCTests(t *testing.T) {
	// require := require.New(t)
	testVals, testNames := getInput("dectest/ddAdd.decTest")

	var testName string
	var TestResult string
	var TestPass bool
	var testVal1, testVal2, testAns Decimal64
	var parseResult string

	for i := range testNames {
		testName = testNames[i]
		testVal1, testVal2, testAns, parseResult = convertToDec64(testVals[testName], testName)
		TestPass, TestResult = doTests(testVal1, testVal2, testAns, testVals[testName]["TestFunc"])
		if TestPass == false {
			fmt.Printf("%s: \n %s\n %s\n", testName, TestResult, parseResult)
		}
	}

}

// func TestNew(t *testing.T) {
// 	// require := require.New(t)
// 	// // propper rounding
// 	// require.Equal(MustParseDecimal64("4444444444444445"), MustParseDecimal64("4444444444444444").Add(MustParseDecimal64("0.5001")))
// 	// require.Equal(MustParseDecimal64("0.23"), MustParseDecimal64("1.3").Add(MustParseDecimal64("-1.07")))
// 	// fmt.Println("sjoidgf", MustParseDecimal64("4444444444444444").Add(MustParseDecimal64("1.5001")))
// 	// // require.Equal(MustParseDecimal64("12345678901234.29"), MustParseDecimal64("12345678901234").Add(MustParseDecimal64("0.2951")))
//
// }
