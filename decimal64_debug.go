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
	sign        int8
	exp         int16
	significand uint64
	bits        uint64
}

func new64(bits uint64) Decimal64 {
	d := new64nostr(bits)
	d.s = d.String()
	return d
}

func new64str(bits uint64, s string) Decimal64 {
	d := new64nostr(bits)
	d.s = s
	return d
}

func new64nostr(bits uint64) Decimal64 {
	d := Decimal64{bits: bits}

	fl, sign, exp, significand := d.parts()

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
