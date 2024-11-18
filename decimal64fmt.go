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
func (d Decimal64) Append(buf []byte, format byte, prec int) []byte {
	return DefaultFormatContext64.append(d, buf, -1, prec, noFlags, rune(format))
}

func precScale(prec int) Decimal64 {
	return new64(newFromPartsRaw(0, -15-max(0, prec), decimal64Base).bits)
}

// Append appends the text representation of d to buf.
func (ctx Context64) append(
	d Decimal64,
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

		whole := significand / decimal64Base
		buf = append(buf, byte('0'+whole))
		frac := significand - decimal64Base*whole
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
			buf = appendZeros(buf, exp)
			return dotZeros(buf, a.prec)
		}

		// integer part
		fracDigits := min(-exp, 16)
		unit := tenToThe[fracDigits]
		buf = a.uint64New(buf, significand/unit)
		buf = appendZeros(buf, exp)

		// empty fractional part
		if significand%unit == 0 {
			return dotZeros(buf, a.prec)
		}

		buf = append(buf, '.')

		// fractional part
		prefix := max(0, -exp-16)
		if a.prec < 0 {
			buf = appendZeros(buf, prefix)
			buf = appendFracF(buf, significand, fracDigits, 16)
			buf = bytes.TrimRight(buf, "0")
			return buf
		}
		buf = appendZeros(buf, min(a.prec, prefix))
		return appendFracF(buf, significand, fracDigits, a.prec-prefix)
	case 'g', 'G':
		if exp < -decimal64Digits-3 ||
			a.prec >= 0 && exp > a.prec ||
			a.prec < 0 && exp > -decimal64Digits+6 {
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
func (d Decimal64) Format(s fmt.State, verb rune) {
	DefaultFormatContext64.format(d, s, verb)
}

// format implements fmt.Formatter.
func (ctx Context64) format(d Decimal64, s fmt.State, verb rune) {
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
		fmt.Fprintf(s, "%%!%c(decimal.Decimal64=%s)", verb, d.String())
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
func (d Decimal64) String() string {
	return DefaultFormatContext64.str(d)
}

func (ctx Context64) str(d Decimal64) string {
	if s, has := small64Strings[d.bits]; has {
		return s
	}
	return ctx.text(d, 'g', -1, -1)
}

// Text converts the floating-point number x to a string according to the given
// format and precision prec.
func (d Decimal64) Text(format byte, prec int) string {
	return DefaultFormatContext64.text(d, rune(format), -1, prec)
}

func (ctx Context64) text(d Decimal64, verb rune, width, prec int) string {
	var buf [32]byte
	return string(ctx.append(d, buf[:0], width, prec, noFlags, verb))
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

func (c Contextual) Format(s fmt.State, verb rune) {
	c.ctx.format(c.d, s, verb)
}

func (c Contextual) Text(verb rune, width, prec int) string {
	return c.ctx.text(c.d, verb, width, prec)
}

func (ctx Context64) With(d Decimal64) Contextual {
	return Contextual{ctx, d}
}
