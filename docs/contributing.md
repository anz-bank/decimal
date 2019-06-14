## Recommended preliminary learning

To gain familiarity with this project, it is recommended to develop a further understanding of the floating point standard, below are some links to get you started
- [IEEE 754-2008 revision - Wikipedia](https://en.wikipedia.org/wiki/IEEE_754-2008_revision)

- [IEEE Standard for Floating-Point
Arithmetic - PDF](http://www.dsc.ufcg.edu.br/~cnum/modulos/Modulo2/IEEE754_2008.pdf)

- [Decimal specification from IBM](http://speleotrove.com/decimal/)

## Running unit tests

### IBM unit tests

Our unit tests are provided by Mike Cowlishaw from IBM, and are referenced in our test suite `decimalSuite_test.go`

You can define which unit tests to run, for example:

```go
var tests = []string {
    "dectest/ddAdd.decTest",
}
```

Run this command to run all the unit tests:

```bash
go test -v .
```

### Custom unit test

If you want to debug your code, the best way to go about this is to create a new file, e.g. `decimal64NewTest_test.go`, you will need `_test.go` for unit tests to work.

You can try this sample unit test for the Fused-Multiply-Add operator (FMA):

```go
package decimal

import (
    "fmt"
    "testing"
)

func TestNewTest(t *testing.T) {
	a := MustParseDecimal64("1")
	b := MustParseDecimal64("2.5")
	c := MustParseDecimal64("9999999999999999e4")
	ans := a.FMA(b, c)
	fmt.Println(ans)
}
```

In your terminal, you can run your unit test with this command:

```bash
go test -v -run TestNewTest
```
