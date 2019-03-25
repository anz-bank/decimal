package decimal

import (
	"fmt"
)

func appendFrac64(buf []byte, n, limit uint64) []byte {
	for n > 0 {
		msd := n / limit
		buf = append(buf, byte('0'+msd))
		n -= limit * msd
		limit /= 10
	}
	return buf
}

func appendUint64(buf []byte, n, limit uint64) []byte {
	zeroPrefix := false
	for limit > 0 {
		msd := n / limit
		if msd > 0 || zeroPrefix {
			buf = append(buf, byte('0'+msd))
			zeroPrefix = true
		}
		n -= limit * msd
		limit /= 10
	}
	return buf
}

// Append appends the text representation of d to buf.
func (d Decimal64) Append(buf []byte, format byte, prec int) []byte {
	flavor, sign, exp, significand := d.parts()
	if sign == 1 {
		buf = append(buf, '-')
	}
	switch flavor {
	case flQNaN, flSNaN:
		return appendUint64(append(buf, []byte("NaN")...), significand, 10000)
	case flInf:
		return append(buf, []byte("inf")...)
	}

formatBlock:
	switch format {
	case 'e', 'E':
		whole := significand / decimal64Base
		buf = append(buf, byte('0'+whole))
		frac := significand - decimal64Base*whole
		if frac > 0 {
			buf = appendFrac64(append(buf, '.'), frac, decimal64Base/10)
		}

		exp += 15
		if exp != 0 {
			buf = append(buf, format)
			if exp < 0 {
				buf = append(buf, '-')
				exp = -exp
			} else {
				buf = append(buf, '+')
			}
			buf = appendUint64(buf, uint64(exp), 1000)
		}
		return buf
	case 'f', 'F':
		exp, whole, frac := expWholeFrac(exp, significand)
		if whole > 0 {
			buf = appendUint64(buf, whole, decimal64Base)
			for ; exp > 0; exp-- {
				buf = append(buf, '0')
			}
		} else {
			buf = append(buf, '0')
		}
		if frac > 0 {
			buf = appendFrac64(append(buf, '.'), frac, decimal64Base)
		}
		return buf
	case 'g', 'G':
		if exp < -16-4 || exp > prec {
			format -= 'g' - 'e'
		} else {
			format -= 'g' - 'f'
		}
		goto formatBlock
	default:
		return append(buf, '%', format)
	}
}

// Format implements fmt.Formatter.
func (d Decimal64) Format(s fmt.State, format rune) {
	prec, hasPrec := s.Precision()
	if !hasPrec {
		prec = 6
	}
	switch format {
	case 'e', 'E', 'f', 'F', 'g', 'G':
		// nothing to do
	case 'v':
		format = 'g'
	default:
		fmt.Fprintf(s, "%%!%c(*decimal.Decimal64=%s)", format, d.String())
		return
	}
	s.Write(d.Append(make([]byte, 0, 8), byte(format), prec))
}

// String returns a string representation of d.
func (d Decimal64) String() string {
	return d.Text('g', 10)
}

// Text converts the floating-point number x to a string according to the given
// format and precision prec.
func (d Decimal64) Text(format byte, prec int) string {
	return string(d.Append(make([]byte, 0, 8), format, prec))
}
