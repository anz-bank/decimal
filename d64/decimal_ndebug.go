//go:build !decimal_debug
// +build !decimal_debug

package d64

// Decimal represents an IEEE 754 64-bit floating point decimal number.
// It uses the binary representation method.
// Decimal is intentionally a struct to ensure users don't accidentally cast it to uint64.
type Decimal struct {
	bits uint64
}

func newDec(bits uint64) Decimal {
	return Decimal{bits: bits}
}

func newStr(bits uint64, _ string) Decimal {
	return Decimal{bits: bits}
}

func newNostr(bits uint64) Decimal {
	return Decimal{bits: bits}
}

func checkSignificandIsNormal(significand uint64) {}
