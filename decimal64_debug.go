//go:build decimal_debug
// +build decimal_debug

package decimal

import "fmt"

// Decimal64 represents an IEEE 754 64-bit floating point decimal number.
// It uses the binary representation method.
// Decimal64 is intentionally a struct to ensure users don't accidentally cast it to uint64.
type Decimal64 struct {
	s           string
	fl          flavor
	sign        int
	exp         int
	significand uint64
	bits        uint64
}

// This should be the only point at which Decimal64 instances are constructed raw.
// The verbose construction below makes it easy to audit accidental raw cosntruction.
// A search for (?<!\[\])Decimal64\{ must come up empty.
func new64(bits uint64) Decimal64 {
	d := new64Raw(bits)

	fl, sign, exp, significand := d.parts()

	d.s = d.String()
	d.fl = fl
	d.sign = sign
	d.exp = exp
	d.significand = significand
	d.bits = d.bits

	return d
}

func checkSignificandIsNormal(significand uint64) {
	if decimal64Base > significand {
		panic(fmt.Errorf("Failed logic check: %d <= %d", decimal64Base, significand))
	}

	if significand >= 10*decimal64Base {
		panic(fmt.Errorf("Failed logic check: %d < %d", significand, 10*decimal64Base))
	}
}
