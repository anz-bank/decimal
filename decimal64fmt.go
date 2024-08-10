package decimal

import (
	"bytes"
	"fmt"
	"strconv"
)

var _ fmt.Formatter = Zero64
var _ fmt.Scanner = (*Decimal64)(nil)
var _ fmt.Stringer = Zero64

func appendFrac64(buf []byte, n, limit uint64) []byte {
	for n > 0 {
		msd := n / limit
		buf = append(buf, byte('0'+msd))
		n -= limit * msd
		limit /= 10
	}
	return buf
}

var zeros = []byte("0000000000000000")

func appendZeros(buf []byte, n int) []byte {
	if n <= 0 {
		return buf
	}
	l := len(zeros)
	for ; n > l; n -= l {
		buf = append(buf, zeros...)
	}
	return append(buf, zeros[:n]...)
}

func appendFrac64Prec(buf []byte, n uint64, prec int) []byte {
	if prec < 0 {
		return buf
	}
	// Add a digit in front so strconv.AppendUint doesn't trim leading zeros.
	n += 10 * decimal64Base
	if prec < 16 {
		unit := powersOf10[max(0, 16-prec)]
		rem := n % unit
		n /= unit
		if rem > unit/2 || rem == unit/2 && n%2 == 1 {
			n++
		}
	}

	// p/2 adds 5 to the digit past the desired precision in order to round up.
	buflen := len(buf)
	prefix := buf[buflen-1]
	buf = strconv.AppendUint(buf[:buflen-1], n, 10)
	buf[buflen-1] = prefix

	return appendZeros(buf, prec-decimal64Digits)
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

func appendUint64New(buf []byte, n, limit uint64) []byte {
	return strconv.AppendUint(buf, n/(decimal64Base/limit), 10)
}

// Append appends the text representation of d to buf.
func (d Decimal64) Append(buf []byte, format byte, prec int) []byte {
	return d.append(buf, format, -1, prec)
}

var dotSuffix = []byte{'.'}

// Append appends the text representation of d to buf.
func (d Decimal64) append(buf []byte, format byte, width, prec int) []byte {
	flav, sign, exp, significand := d.parts()
	if sign == 1 {
		buf = append(buf, '-')
	}
	switch flav {
	case flQNaN, flSNaN:
		buf = append(buf, []byte("NaN")...)
		if significand != 0 {
			return appendUint64(buf, significand, 10000)
		}
		return buf
	case flInf:
		return append(buf, []byte("inf")...)
	}

formatBlock:
	switch format {
	case 'e', 'E':
		// normalise subnormals
		exp, significand = unsubnormal(exp, significand)

		whole := significand / decimal64Base
		buf = append(buf, byte('0'+whole))
		frac := significand - decimal64Base*whole
		if frac > 0 {
			buf = appendFrac64(append(buf, '.'), frac, decimal64Base/10)
		}

		if significand == 0 {
			return buf
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
		exponent, whole, frac := expWholeFrac(exp, significand)
		if whole > 0 {
			buf = appendUint64New(buf, whole, decimal64Base)
			for ; exponent > 0; exponent-- {
				buf = append(buf, '0')
			}
		} else {
			buf = append(buf, '0')
		}
		if frac > 0 || prec != 0 {
			p := prec
			if prec == -1 {
				p = decimal64Digits
			}
			buf = append(buf, '.')
			if exponent < 0 {
				p += exponent
				buf = appendZeros(buf, min(-exponent, prec))
			}
			buf = appendFrac64Prec(buf, frac, p)
			if prec == -1 {
				buf = bytes.TrimRight(buf, "0")
			}
			buf = bytes.TrimSuffix(buf, dotSuffix)
		}
		return buf
	case 'g', 'G':
		if exp < -decimal64Digits-3 ||
			prec >= 0 && exp > prec ||
			prec < 0 && exp > -decimal64Digits+6 {
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
	width, hasWidth := s.Width()
	if !hasWidth {
		width = -1
	}
	prec, hasPrec := s.Precision()
	if !hasPrec {
		prec = -1
	}
	switch format {
	case 'e', 'E', 'f', 'F':
		if !hasPrec {
			prec = 6
		}
	case 'g', 'G':
	case 'v':
		format = 'g'
	default:
		fmt.Fprintf(s, "%%!%c(*decimal.Decimal64=%s)", format, d.String())
		return
	}
	s.Write(d.append(make([]byte, 0, 16), byte(format), width, prec)) //nolint:errcheck
}

// String returns a string representation of d.
func (d Decimal64) String() string {
	return d.Text('g', -1)
}

// Text converts the floating-point number x to a string according to the given
// format and precision prec.
func (d Decimal64) Text(format byte, prec int) string {
	return string(d.Append(make([]byte, 0, 16), format, prec))
}
