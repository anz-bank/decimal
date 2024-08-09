//go:build decimal_debug
// +build decimal_debug

package decimal

// Decimal64 represents an IEEE 754 64-bit floating point decimal number.
// It uses the binary representation method.
// Decimal64 is intentionally a struct to ensure users don't accidentally cast it to uint64
type debugInfo struct {
	s           string
	fl          flavor
	sign        int
	exp         int
	significand uint64
}

func (d Decimal64) debug() Decimal64 {
	s := d.s
	if s == "" {
		s = d.String()
	}
	fl, sign, exp, significand := d.parts()
	return Decimal64{
		bits: d.bits,
		debugInfo: debugInfo{
			s:           s,
			fl:          fl,
			sign:        sign,
			exp:         exp,
			significand: significand,
		},
	}
}
