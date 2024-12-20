//go:build !decimal_debug
// +build !decimal_debug

package decimal

// Decimal64 represents an IEEE 754 64-bit floating point decimal number.
// It uses the binary representation method.
// Decimal64 is intentionally a struct to ensure users don't accidentally cast it to uint64.
type Decimal64 struct {
	bits uint64
}

func new64(bits uint64) Decimal64 {
	return Decimal64{bits: bits}
}

func new64str(bits uint64, _ string) Decimal64 {
	return Decimal64{bits: bits}
}

func new64nostr(bits uint64) Decimal64 {
	return Decimal64{bits: bits}
}

func checkSignificandIsNormal(significand uint64) {}
