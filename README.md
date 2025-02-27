# decimal

This library implements fixed-precision decimal numbers based on the [IEEE 754-2019 standard](https://ieeexplore.ieee.org/document/8766229).
More info can be found at <http://speleotrove.com/decimal/>.

## Features

- Decimal, partial implementation of the ieee-754R standard
- Rounding modes: half up, half even, down (towards zero)
- Up to 3 times faster than arbitrary precision decimal libraries in Go

## Goals

- Implement 128 bit decimal.
- Implement as much of <https://speleotrove.com/decimal/decarith.pdf> as possible.

## Installation and use

Run `go get github.com/anz-bank/decimal/d64`

```go
package main

import (
	"fmt"

	"github.com/anz-bank/decimal/d64"
)

func main() {
	var a d64.Decimal
	b := d64.MustParse("0.1")
	c := d64.MustParse("0.3")
	d := d64.NewFromInt64(123456)

	fmt.Println(a, b, c, d)
}
```

## Usage notes

The d128 package doesn't exist yet, so d64 is assumed below.

### Formatting

`Decimal` provides numerous ways to present numbers for human and machine consumption.

`Decimal` implements the following conventional interfaces:

- `fmt`: `Formatter`, `Scanner` and `Stringer`
  - It currently supports specifying a precision argument, e.g., `%.10f` for the `f` and `F` verbs, while support for `g` and `G` is [planned](https://github.com/anz-bank/decimal/issues/72), as is [support for width specifiers](https://github.com/anz-bank/decimal/issues/72).
- `json`: `Marshaller` and `Unmarshaller`
- `encoding`: `BinaryMarshaler`, `BinaryUnmarshaler`, `TextMarshaler` and `TextUnmarshaler`
- `encoding/gob`: `GobEncoder` and `GobDecoder`

The following methods provide more direct access to the internal methods used to implement `fmt.Formatter`.
For maximum control, use `fmt.Printf` &co or invoke the `fmt.Formatter` interface directly.

- `Decimal.Append` formats straight into a `[]byte` buffer.
- `Decimal.Text` formats in the same way, but returns a `string`.

### Debugging

tl;dr: Use the `decimal_debug` compiler tag during debugging to greatly ease runtime inspection of `Decimal` values.

Debugging with the `d64` package can be challenging because a decimal number is encoded in a `uint64` and the values it holds are inscrutable even to the trained eye.
For example, the number one `One` is represented internally as the number `3450757314565799936` (`2fe38d7ea4c68000` in hexadecimal).

To ease debugging, `Decimal` holds an optional `debugInfo` structure that contains a string representation and unpacked components of the `uint64` representation for every `Decimal` value.

This feature is enabled through the `decimal_debug` compiler tag.
This is done at compile time instead of through runtime flags because having the structure there, even if not used, would double the size of each number and greatly increase the cost of using it.
The size and runtime cost of this feature is zero when the compiler tag is not present.

## Docs

<https://godoc.org/github.com/anz-bank/decimal/d64>

## Why decimal?

Binary floating point numbers are fundamentally flawed when it comes to representing exact numbers in a decimal world.
Just like 1/3 can't be represented in base 10 (it evaluates to 0.3333333333 repeating), 1/10 can't be represented in binary.
The solution is to use a decimal floating point number.
Binary floating point numbers (often just called floating point numbers) are usually in the form `Sign * Significand * 2 ^ exp` and decimal floating point numbers change this to `Sign * Significand * 10 ^ exp`.
The use of base 10 eliminates the decimal fraction problem.

## Why fixed precision?

Most implementations of a decimal floating point datatype implement an *arbitrary precision* type, which often uses an underlying big int.
This gives flexibility in that as the number grows, the number of bits assigned to the number grows (hence the term "arbitrary precision").
This library is different.
It uses a 64-bit decimal datatype as specified in the IEEE-754R standard.
This sacrifices the ability to represent arbitrarily large numbers, but is much faster than arbitrary precision libraries.
There are two main reasons for this:

1. The fixed-size data type is a `uint64` under the hood and never requires heap allocation.
2. All the algorithms can hard-code assumptions about the number of bits to work with.
   In fact, many of the operations work on the entire number as a single unit
   using 64-bit integer arithmetic and, on the occasions it needs to use more,
   128 bits always suffices.
