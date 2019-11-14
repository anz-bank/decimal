|:no_entry: NOTICE: USE OF THIS LIBRARY IN ANZ SYSTEMS IS STRICTLY PROHIBITED. :no_entry:|
|-|
|This library is currently in development and has not undergone a thorough correctness assessment or security audit. It is therefore NOT suitable for production use until full approval is granted by the ANZ open source committee and this notice removed.|

# decimal

This library implements fixed-precision decimal numbers based on IEEE 754R standard;
<https://ieeexplore.ieee.org/document/4674342>
More info can be found at:
<http://speleotrove.com/decimal/>

# Features
- Decimal64, partial implementation of the ieee-754R standard
- Half up and half even rounding
- Up to 3 times faster than arbitrary precision decimal libraries in Go

# Goals
- To implement 128 bit decimal

# Installation and use

Run `go get github.com/anz-bank/decimal`


```go
package main

import (
	"fmt"

	"github.com/anz-bank/decimal"
)

func main() {
	var a decimal.Decimal64
	b := decimal.MustParse64("0.1")
	c := decimal.MustParse64("0.3")
	d := decimal.New64FromInt64(123456)

	fmt.Println(a, b, c, d)
}

```

# Docs
<https://godoc.org/github.com/anz-bank/decimal>

# Why decimal
Binary floating point numbers are fundamentally flawed when it comes to representing exact numbers in a decimal world. Just like 1/3 can't be represented in base 10 (it evaluates to 0.3333333333 repeating), 1/10 can't be represented in binary.
The solution is to use a decimal floating point number.
Binary floating point numbers (often just called floating point numbers) are usually in the form
`Sign * Significand * 2 ^ exp`
and decimal floating point numbers change this to
`Sign * Significand * 10 ^ exp`
This eliminates the decimal fraction problem, as the base is in 10.


# Why fixed precision
Most implementations of a decimal floating point datatype implement an 'arbitrary precision' type, which often uses an underlying big int. This gives flexibility in that as the number grows, the number of bits assigned to the number grows ( and thus 'arbitrary precision').
This library is different as it specifies a 64 bit decimal datatype as specified in the ieee-754R standard. This gives the sacrifice of being able to represent arbitrarily large numbers, but is faster than other arbitrary precision libraries.
