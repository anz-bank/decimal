//go:build decimal_debug
// +build decimal_debug

package d64

import "fmt"

// Decimal represents an IEEE 754 64-bit floating point decimal number.
// It uses the binary representation method.
// Decimal is intentionally a struct to ensure users don't accidentally cast it to uint64.
type Decimal struct {
	s           string
	fl          flavor
	sign        int8
	exp         int16
	significand uint64
	bits        uint64
}

func newDec(bits uint64) Decimal {
	d := newNostr(bits)
	d.s = d.String()
	return d
}

func newStr(bits uint64, s string) Decimal {
	d := newNostr(bits)
	d.s = s
	return d
}

func newNostr(bits uint64) Decimal {
	d := Decimal{bits: bits}

	fl, sign, exp, significand := d.parts()

	d.fl = fl
	d.sign = sign
	d.exp = exp
	d.significand = significand
	d.bits = d.bits

	return d
}

func checkSignificandIsNormal(significand uint64) {
	if decimalBase > significand {
		panic(fmt.Errorf("Failed logic check: %d <= %d", decimalBase, significand))
	}

	if significand >= 10*decimalBase {
		panic(fmt.Errorf("Failed logic check: %d < %d", significand, 10*decimalBase))
	}
}
