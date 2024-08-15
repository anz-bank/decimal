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

var zeros = []byte("0000000000000000")

type appender []byte

func newAppender() *appender {
	a := make(appender, 0, 16)
	return &a
}

func (buf *appender) Append(data ...byte) *appender {
	*buf = append(*buf, data...)
	return buf
}

func (buf *appender) Zeros(n int) *appender {
	if n <= 0 {
		return buf
	}
	l := len(zeros)
	for ; n > l; n -= l {
		buf.Append(zeros...)
	}
	return buf.Append(zeros[:n]...)
}

func (buf *appender) DotZeros(n int) *appender {
	if n > 0 {
		buf.Append('.').Zeros(n)
	}
	return buf
}

func (buf *appender) Uint64(n, limit uint64) *appender {
	zeroPrefix := false
	for limit > 0 {
		msd := n / limit
		if msd > 0 || zeroPrefix {
			buf.Append(byte('0' + msd))
			zeroPrefix = true
		}
		n -= limit * msd
		limit /= 10
	}
	return buf
}

func (buf *appender) Uint64New(n uint64) *appender {
	*buf = strconv.AppendUint(*buf, n, 10)
	return buf
}

func (buf *appender) Digits(n uint64, width, prec int) *appender {
	if width > prec {
		n /= powersOf10[width-prec]
		width = prec
	}
	start := len(*buf)
	buf.Append(make([]byte, width)...)
	slot := (*buf)[start:]
	for i := width - 1; i >= 0; i-- {
		d := n / 10
		slot[i] = byte('0' + n - 10*d)
		n = d
	}
	buf.Zeros(prec - width)
	return buf
}

func (buf *appender) Frac64(n, limit uint64) *appender {
	for n > 0 {
		msd := n / limit
		buf.Append(byte('0' + msd))
		n -= limit * msd
		limit /= 10
	}
	return buf
}

func (buf *appender) TrimTrailingZeros() *appender {
	*buf = bytes.TrimRight(*buf, "0")
	return buf
}

// Append appends the text representation of d to buf.
func (d Decimal64) Append(buf []byte, format byte, prec int) []byte {
	a := appender(buf)
	return *DefaultFormatContext64.append(d, &a, rune(format), -1, prec)
}

func precScale(prec int) Decimal64 {
	return newFromPartsRaw(0, -15-max(0, prec), decimal64Base)
}

// Append appends the text representation of d to buf.
func (ctx Context64) append(d Decimal64, buf *appender, format rune, width, prec int) *appender {
	if buf == nil {
		buf = newAppender()
	}
	flav, sign, exp, significand := d.parts()
	if sign == 1 {
		buf.Append('-')
	}
	switch flav {
	case flQNaN, flSNaN:
		buf.Append([]byte("NaN")...)
		if significand != 0 {
			return buf.Uint64(significand, 10000)
		}
		return buf
	case flInf:
		return buf.Append([]byte("inf")...)
	}

formatBlock:
	switch format {
	case 'e', 'E':
		exp, significand = unsubnormal(exp, significand)

		whole := significand / decimal64Base
		buf.Append(byte('0' + whole))
		frac := significand - decimal64Base*whole
		if frac > 0 {
			buf.Append('.').Frac64(frac, decimal64Base/10)
		}

		if significand == 0 {
			return buf
		}

		exp += 15
		if exp != 0 {
			buf.Append(byte(format))
			if exp < 0 {
				buf.Append('-')
				exp = -exp
			} else {
				buf.Append('+')
			}
			buf = buf.Uint64(uint64(exp), 1000)
		}
		return buf
	case 'f', 'F':
		if 0 <= prec && prec <= 16 {
			_, _, exp, significand = ctx.roundRaw(d, precScale(prec)).parts()
		}

		if significand == 0 {
			return buf.Append('0').DotZeros(prec)
		}

		exp, significand = unsubnormal(exp, significand)

		// pure integer
		if exp >= 0 {
			return buf.Uint64New(significand).Zeros(exp).DotZeros(prec)
		}

		// integer part
		fracDigits := min(-exp, 16)
		unit := powersOf10[fracDigits]
		buf = buf.Uint64New(significand / unit).Zeros(exp)

		// empty fractional part
		if significand%unit == 0 {
			return buf.DotZeros(prec)
		}

		buf.Append('.')

		// fractional part
		prefix := max(0, -exp-16)
		if prec == -1 {
			return buf.Zeros(prefix).Digits(significand, fracDigits, 16).TrimTrailingZeros()
		}
		return buf.Zeros(min(prec, prefix)).Digits(significand, fracDigits, prec-prefix)
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
		return buf.Append('%', byte(format))
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

	s.Write(*ctx.append(d, newAppender(), format, width, prec)) //nolint:errcheck
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
	return string(*ctx.append(d, newAppender(), format, width, prec))
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
