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
	buf  []byte
	wid  int
	prec int
	flag func(int) bool
}

func newAppender(buf []byte, wid, prec int, flag func(int) bool) *appender {
	if flag == nil {
		flag = func(i int) bool { return false }
	}
	return &appender{
		buf:  buf,
		wid:  wid,
		prec: prec,
		flag: flag,
	}
}

func (a *appender) Bytes() []byte {
	return a.buf
}

func (a *appender) Write(b []byte) (n int, err error) {
	a.buf = append(a.buf, b...)
	return len(b), nil
}

func (a *appender) Width() (wid int, ok bool) {
	return a.wid, a.wid >= 0
}

func (a *appender) Precision() (prec int, ok bool) {
	return a.prec, a.prec >= 0
}

func (a *appender) Flag(c int) bool {
	return a.flag(c)
}

func (a *appender) Append(data ...byte) *appender {
	a.buf = append(a.buf, data...)
	return a
}

func (a *appender) Zeros(n int) *appender {
	if n <= 0 {
		return a
	}
	l := len(zeros)
	for ; n > l; n -= l {
		a.Append(zeros[:]...)
	}
	return a.Append(zeros[:n]...)
}

func (a *appender) DotZeros(n int) *appender {
	if n > 0 {
		a.Append('.').Zeros(n)
	}
	return a
}

func (a *appender) Uint64(n, limit uint64) *appender {
	zeroPrefix := false
	for limit > 0 {
		msd := n / limit
		if msd > 0 || zeroPrefix {
			a.Append(byte('0' + msd))
			zeroPrefix = true
		}
		n -= limit * msd
		limit /= 10
	}
	return a
}

func (a *appender) Uint64New(n uint64) *appender {
	a.buf = strconv.AppendUint(a.buf, n, 10)
	return a
}

func (a *appender) Digits(n uint64, width, prec int) *appender {
	if width > prec {
		n /= tenToThe[width-prec]
		width = prec
	}
	start := len(a.buf)
	a.Append(make([]byte, width)...)
	slot := (a.buf)[start:]
	for i := width - 1; i >= 0; i-- {
		d := n / 10
		slot[i] = byte('0' + n - 10*d)
		n = d
	}
	a.Zeros(prec - width)
	return a
}

func (a *appender) Frac64(n, limit uint64) *appender {
	for n > 0 {
		msd := n / limit
		a.Append(byte('0' + msd))
		n -= limit * msd
		limit /= 10
	}
	return a
}

func (a *appender) TrimTrailingZeros() *appender {
	a.buf = bytes.TrimRight(a.buf, "0")
	return a
}

// Append appends the text representation of d to buf.
func (d Decimal64) Append(buf []byte, format byte, prec int) []byte {
	a := newAppender(buf, -1, prec, nil)
	return DefaultFormatContext64.append(d, a, rune(format)).Bytes()
}

func precScale(prec int) Decimal64 {
	return newFromPartsRaw(0, -15-max(0, prec), decimal64Base)
}

// Append appends the text representation of d to buf.
func (ctx Context64) append(d Decimal64, a *appender, format rune) *appender {
	flav, sign, exp, significand := d.parts()
	if sign == 1 {
		a.Append('-')
	}
	switch flav {
	case flQNaN, flSNaN:
		a.Append([]byte("NaN")...)
		if significand != 0 {
			return a.Uint64(significand, 10000)
		}
		return a
	case flInf:
		return a.Append([]byte("inf")...)
	}

formatBlock:
	switch format {
	case 'e', 'E':
		exp, significand = unsubnormal(exp, significand)

		whole := significand / decimal64Base
		a.Append(byte('0' + whole))
		frac := significand - decimal64Base*whole
		if frac > 0 {
			a.Append('.').Frac64(frac, decimal64Base/10)
		}

		if significand == 0 {
			return a
		}

		exp += 15
		if exp != 0 {
			a.Append(byte(format))
			if exp < 0 {
				a.Append('-')
				exp = -exp
			} else {
				a.Append('+')
			}
			a = a.Uint64(uint64(exp), 1000)
		}
		return a
	case 'f', 'F':
		if 0 <= a.prec && a.prec <= 16 {
			_, _, exp, significand = ctx.roundRaw(d, precScale(a.prec)).parts()
		}

		if significand == 0 {
			return a.Append('0').DotZeros(a.prec)
		}

		exp, significand = unsubnormal(exp, significand)

		// pure integer
		if exp >= 0 {
			return a.Uint64New(significand).Zeros(exp).DotZeros(a.prec)
		}

		// integer part
		fracDigits := min(-exp, 16)
		unit := tenToThe[fracDigits]
		a = a.Uint64New(significand / unit).Zeros(exp)

		// empty fractional part
		if significand%unit == 0 {
			return a.DotZeros(a.prec)
		}

		a.Append('.')

		// fractional part
		prefix := max(0, -exp-16)
		if a.prec < 0 {
			return a.Zeros(prefix).Digits(significand, fracDigits, 16).TrimTrailingZeros()
		}
		return a.Zeros(min(a.prec, prefix)).Digits(significand, fracDigits, a.prec-prefix)
	case 'g', 'G':
		if exp < -decimal64Digits-3 ||
			a.prec >= 0 && exp > a.prec ||
			a.prec < 0 && exp > -decimal64Digits+6 {
			format -= 'g' - 'e'
		} else {
			format -= 'g' - 'f'
		}
		goto formatBlock
	default:
		return a.Append('%', byte(format))
	}
}

// Format implements fmt.Formatter.
func (d Decimal64) Format(s fmt.State, format rune) {
	DefaultFormatContext64.format(d, s, format)
}

// format implements fmt.Formatter.
func (ctx Context64) format(d Decimal64, s fmt.State, format rune) {
	prec := optInt(s.Precision())

	switch format {
	case 'e', 'E', 'f', 'F':
		if prec < 0 {
			prec = 6
		}
	case 'g', 'G':
	case 'v':
		format = 'g'
	default:
		fmt.Fprintf(s, "%%!%c(decimal.Decimal64=%s)", format, d.String())
		return
	}

	wid := optInt(s.Width())
	a := newAppender(make([]byte, 0, 16), wid, prec, nil)
	s.Write(ctx.append(d, a, format).Bytes()) //nolint:errcheck
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
	return ctx.text(d, 'g', -1, -1)
}

// Text converts the floating-point number x to a string according to the given
// format and precision prec.
func (d Decimal64) Text(format byte, prec int) string {
	return DefaultFormatContext64.text(d, rune(format), -1, prec)
}

func (ctx Context64) text(d Decimal64, format rune, width, prec int) string {
	return string(ctx.append(d, newAppender(make([]byte, 0, 16), width, prec, nil), format).Bytes())
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
