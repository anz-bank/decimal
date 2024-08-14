package decimal

import (
	"bytes"
	"fmt"
	"strconv"
)

var _ fmt.Formatter = Zero64
var _ fmt.Scanner = (*Decimal64)(nil)
var _ fmt.Stringer = Zero64

// DefaultFormatContext64 is the default context use for formatting Decimal64.
// Unlike [DefaultContext64], it uses HalfEven rounding to conform to standard
// Go formatting for float types.
var DefaultFormatContext64 = Context64{Rounding: HalfEven}

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
		n /= unit
	}

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
	return DefaultFormatContext64.append(d, buf, rune(format), -1, prec)
}

var dotSuffix = []byte{'.'}

func precScale(prec int) Decimal64 {
	return newFromPartsRaw(0, -15-max(0, prec), decimal64Base)
}

// Append appends the text representation of d to buf.
func (ctx Context64) append(d Decimal64, buf []byte, format rune, width, prec int) []byte {
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
			buf = append(buf, byte(format))
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
		if 0 <= prec && prec <= 16 {
			_, _, _, significand = ctx.roundRaw(d, precScale(prec)).parts()
		}

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
		return append(buf, '%', byte(format))
	}
}

// Format implements fmt.Formatter.
func (d Decimal64) Format(s fmt.State, format rune) {
	DefaultFormatContext64.format(d, s, format)
}

// format implements fmt.Formatter.
func (ctx Context64) format(d Decimal64, s fmt.State, format rune) {
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
		fmt.Fprintf(s, "%%!%c(decimal.Decimal64=%s)", format, d.String())
		return
	}

	s.Write(ctx.append(d, make([]byte, 0, 16), format, width, prec)) //nolint:errcheck
}

// String returns a string representation of d.
func (d Decimal64) String() string {
	return DefaultFormatContext64.str(d)
}

func (ctx Context64) str(d Decimal64) string {
	return ctx.text(d, 'g', -1, -1)
}

// Text converts the floating-point number x to a string according to the given
// format and precision prec.
func (d Decimal64) Text(format byte, prec int) string {
	return DefaultFormatContext64.text(d, rune(format), -1, prec)
}

func (ctx Context64) text(d Decimal64, format rune, width, prec int) string {
	return string(ctx.append(d, make([]byte, 0, 16), format, width, prec))
}

// Contextual binds a [Decimal64] to a [Context64] for greater control of formatting.
// It implements [fmt.Stringer] and [fmt.Formatter] on behalf of the number,
// using the context to control formatting.
type Contextual struct {
	ctx Context64
	d   Decimal64
}

func (c Contextual) String() string {
	return c.ctx.str(c.d)
}

func (c Contextual) Format(s fmt.State, format rune) {
	c.ctx.format(c.d, s, format)
}

func (c Contextual) Text(format rune, width, prec int) string {
	return c.ctx.text(c.d, format, width, prec)
}

func (ctx Context64) With(d Decimal64) Contextual {
	return Contextual{ctx, d}
}
