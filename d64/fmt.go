package d64

import (
	"bytes"
	"fmt"
	"strconv"
)

var _ fmt.Formatter = Zero
var _ fmt.Scanner = (*Decimal)(nil)
var _ fmt.Stringer = Zero

// DefaultFormatContext is the default context use for formatting [Decimal].
// Unlike [DefaultContext], it uses HalfEven rounding to conform to standard
// Go formatting for float types.
var DefaultFormatContext = Context{Rounding: HalfEven}

var zeros = [16]byte{
	'0', '0', '0', '0', '0', '0', '0', '0',
	'0', '0', '0', '0', '0', '0', '0', '0',
}

type appender struct {
	wid     int
	prec    int
	flagger interface{ Flag(int) bool }
}

type noFlagsInt int

var noFlags interface{ Flag(int) bool } = noFlagsInt(0)

func (noFlagsInt) Flag(c int) bool {
	return false
}

func appendZeros(buf []byte, n int) []byte {
	if n <= 0 {
		return buf
	}
	l := len(zeros)
	for ; n > l; n -= l {
		buf = append(buf, zeros[:]...)
	}
	return append(buf, zeros[:n]...)
}

func dotZeros(buf []byte, n int) []byte {
	if n > 0 {
		buf = append(buf, '.')
		buf = appendZeros(buf, n)
	}
	return buf
}

func appendUint64(buf []byte, n, limit uint64) []byte {
	// TODO: avoid dividing by a variable.
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

func (a *appender) uint64New(buf []byte, n uint64) []byte {
	return strconv.AppendUint(buf, n, 10)
}

func appendFracF(buf []byte, n uint64, width, prec int) []byte {
	if width > prec {
		n /= tenToThe[width-prec]
		width = prec
	}
	buf = formatBits10(buf, n%tenToThe[width], width)
	return appendZeros(buf, prec-width)
}

func appendFracE(buf []byte, n uint64) []byte {
	w := 15
	for n%10 == 0 {
		n /= 10
		w--
	}
	return formatBits10(buf, n, w)
}

// Append appends the text representation of d to buf.
func (d Decimal) Append(buf []byte, format byte, prec int) []byte {
	return DefaultFormatContext.append(d, buf, -1, prec, noFlags, rune(format))
}

func precScale(prec int) Decimal {
	return newDec(newFromPartsRaw(0, -15-max(0, int16(prec)), decimalBase).bits)
}

// Append appends the text representation of d to buf.
func (ctx Context) append(
	d Decimal,
	buf []byte,
	wid int,
	prec int,
	flagger interface{ Flag(int) bool },
	verb rune,
) []byte {
	if buf == nil {
		buf = make([]byte, 0, 32)
	}
	a := appender{wid, prec, flagger}

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
	switch verb {
	case 'e', 'E':
		exp, significand = unsubnormal(exp, significand)

		whole := significand / decimalBase
		buf = append(buf, byte('0'+whole))
		frac := significand - decimalBase*whole
		if frac > 0 {
			buf = append(buf, '.')
			buf = appendFracE(buf, frac)
		}

		if significand == 0 {
			return buf
		}

		exp += 15
		if exp != 0 {
			buf = append(buf, byte(verb))
			if exp < 0 {
				buf = append(buf, '-')
				exp = -exp
			} else {
				buf = append(buf, '+')
			}
			return appendUint64(buf, uint64(exp), 1000)
		}
		return buf
	case 'f', 'F':
		if 0 <= a.prec && a.prec <= 16 {
			_, _, exp, significand = ctx.roundRaw(d, precScale(a.prec)).parts()
		}

		if significand == 0 {
			buf = append(buf, '0')
			return dotZeros(buf, a.prec)
		}

		exp, significand = unsubnormal(exp, significand)

		// pure integer
		if exp >= 0 {
			buf = a.uint64New(buf, significand)
			buf = appendZeros(buf, int(exp))
			return dotZeros(buf, a.prec)
		}

		// integer part
		fracDigits := int(min(-exp, 16))
		unit := tenToThe[fracDigits]
		buf = a.uint64New(buf, significand/unit)
		buf = appendZeros(buf, int(exp))

		// empty fractional part
		if significand%unit == 0 {
			return dotZeros(buf, a.prec)
		}

		buf = append(buf, '.')

		// fractional part
		prefix := int(max(0, -exp-16))
		if a.prec < 0 {
			buf = appendZeros(buf, prefix)
			buf = appendFracF(buf, significand, fracDigits, 16)
			buf = bytes.TrimRight(buf, "0")
			return buf
		}
		buf = appendZeros(buf, min(a.prec, prefix))
		return appendFracF(buf, significand, fracDigits, a.prec-prefix)
	case 'g', 'G':
		if exp < -decimalDigits-3 ||
			a.prec >= 0 && int(exp) > a.prec ||
			a.prec < 0 && exp > -decimalDigits+6 {
			verb -= 'g' - 'e'
		} else {
			verb -= 'g' - 'f'
		}
		goto formatBlock
	default:
		return append(buf, '%', byte(verb))
	}
}

// Format implements fmt.Formatter.
func (d Decimal) Format(s fmt.State, verb rune) {
	DefaultFormatContext.format(d, s, verb)
}

// format implements fmt.Formatter.
func (ctx Context) format(d Decimal, s fmt.State, verb rune) {
	prec := optInt(s.Precision())

	switch verb {
	case 'e', 'E', 'f', 'F':
		if prec < 0 {
			prec = 6
		}
	case 'g', 'G':
	case 'v':
		verb = 'g'
	default:
		fmt.Fprintf(s, "%%!%c(d64.Decimal=%s)", verb, d.String())
		return
	}

	wid := optInt(s.Width())

	s.Write(ctx.append(d, nil, wid, prec, s, verb)) //nolint:errcheck
}

func optInt(i int, has bool) int {
	if has {
		return i
	}
	return -1
}

// String returns a string representation of d.
func (d Decimal) String() string {
	return DefaultFormatContext.str(d)
}

func (ctx Context) str(d Decimal) string {
	if s, has := smallStrings[d.bits]; has {
		return s
	}
	return ctx.text(d, 'g', -1, -1)
}

// Text converts the floating-point number x to a string according to the given
// format and precision prec.
func (d Decimal) Text(format byte, prec int) string {
	return DefaultFormatContext.text(d, rune(format), -1, prec)
}

func (ctx Context) text(d Decimal, verb rune, width, prec int) string {
	var buf [32]byte
	return string(ctx.append(d, buf[:0], width, prec, noFlags, verb))
}

// Contextual binds a [Decimal] to a [Context] for greater control of formatting.
// It implements [fmt.Stringer] and [fmt.Formatter] on behalf of the number,
// using the context to control formatting.
type Contextual struct {
	ctx Context
	d   Decimal
}

func (c Contextual) String() string {
	return c.ctx.str(c.d)
}

func (c Contextual) Format(s fmt.State, verb rune) {
	c.ctx.format(c.d, s, verb)
}

func (c Contextual) Text(verb rune, width, prec int) string {
	return c.ctx.text(c.d, verb, width, prec)
}

func (ctx Context) With(d Decimal) Contextual {
	return Contextual{ctx, d}
}
