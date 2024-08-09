# decimal

This library implements fixed-precision decimal numbers based on IEEE 754R standard;
<https://ieeexplore.ieee.org/document/4674342>.
More info can be found at:
<http://speleotrove.com/decimal/>

## Features

- Decimal64, partial implementation of the ieee-754R standard
- Half up and half even rounding
- Up to 3 times faster than arbitrary precision decimal libraries in Go

## Goals

- Implement 128 bit decimal
- Implement as much of <https://speleotrove.com/decimal/decarith.pdf> as possible.

## Installation and use

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

## Usage notes

### Formatting

`Decimal64` provides a range of ways to present numbers, both for human and
machine consumption.

`Decimal64` implements the following conventional interfaces:

- `fmt`: `Formatter`, `Scanner` and `Stringer`
  - It currently supports specifying a precision argument, e.g., `%.10f` for the `f` and `F` verbs, while support `g` and `G` is [on the board](https://github.com/anz-bank/decimal/issues/72), as is [support for a width argument](https://github.com/anz-bank/decimal/issues/72).
- `json`: `Marshaller` and `Unmarshaller`
- `encoding`: `BinaryMarshaler`, `BinaryUnmarshaler`, `TextMarshaler` and `TextUnmarshaler`
- `encoging/gob`: `GobEncoder` and `GobDecoder`
- thus enabling the use of decimal numbers in the `fmt.Printf` family.
- `Decimal64.Append` formats straight into a `[]byte` buffer.
- `Decimal64.Text` formats in the same way, but returns a `string`.

### Debugging

tl;dr: Use the `decimal_debug` compiler tag during debugging to greatly ease
runtime inspection of `Decimal64` values.

Debugging with the `decimal` package can be challeging because a `Decimal64`
number is encoded in a `uint64` and the values it holds are inscrutable even to
the trained eye. For example, the number one `decimal.One64` is represented
internally as the number `3450757314565799936` (`2fe38d7ea4c68000` in
hexadecimal).

To ease debugging, `Decimal64` holds an optional `debugInfo` structure that
contains a string representation and unpacked components of the `uint64`
representation for every `Decimal64` value.

This feature is enabled through the `decimal_debug` compiler tag. This is done
at compile time instead of through runtime flags because having the structure
there even if not used would double the size of each number and greatly increase
the cost of using it. The size and runtime cost of this feature is zero when the
compiler tag is not present.

## Docs

<https://godoc.org/github.com/anz-bank/decimal>

## Why decimal?

Binary floating point numbers are fundamentally flawed when it comes to representing exact numbers in a decimal world. Just like 1/3 can't be represented in base 10 (it evaluates to 0.3333333333 repeating), 1/10 can't be represented in binary.
The solution is to use a decimal floating point number.
Binary floating point numbers (often just called floating point numbers) are usually in the form `Sign * Significand * 2 ^ exp` and decimal floating point numbers change this to `Sign * Significand * 10 ^ exp`.
The use of base 10 eliminates the decimal fraction problem.

## Why fixed precision?

Most implementations of a decimal floating point datatype implement an *arbitrary precision* type, which often uses an underlying big int. This gives flexibility in that as the number grows, the number of bits assigned to the number grows ( hence the term "arbitrary precision").
This library is different. It uses a 64-bit decimal datatype as specified in the IEEE-754R standard. This sacrifices the ability to represent arbitrarily large numbers, but is much faster than arbitrary precision libraries.
There are two main reasons for this:

1. The fixed-size data type is a `uint64` under the hood and never requires
   heap allocation.
2. All the algorithms can hard-code assumptions about the number of bits to work with. In fact, many of the operations work on the entire number as a single unit using 64-bit integer arithmetic and, on the occasions it needs to use more, 128 bits always suffices.
