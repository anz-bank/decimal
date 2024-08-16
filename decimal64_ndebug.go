//go:build !decimal_debug
// +build !decimal_debug

package decimal

// Decimal64 represents an IEEE 754 64-bit floating point decimal number.
// It uses the binary representation method.
// Decimal64 is intentionally a struct to ensure users don't accidentally cast it to uint64.
type Decimal64 struct {
	bits uint64
}

// This should be the only point at which Decimal64 instances are constructed raw.
// The verbose construction below makes it easy to audit accidental raw cosntruction.
// A search for (?<!\[\])Decimal64\{ must come up empty.
func new64(bits uint64) Decimal64 {
	return new64Raw(bits)
}

func checkSignificandIsNormal(significand uint64) {}
