//go:build !decimal_debug
// +build !decimal_debug

package decimal

// Decimal64 represents an IEEE 754 64-bit floating point decimal number.
// It uses the binary representation method.
// Decimal64 is intentionally a struct to ensure users don't accidentally cast it to uint64
type debugInfo struct{} //nolint:unused

func (d Decimal64) debug() Decimal64 {
	return d
}
