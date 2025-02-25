package d64

import (
	"fmt"
	"io"
	"strings"
)

var DefaultScanContext = DefaultFormatContext

// Parse parses a string representation of a number as a [Decimal].
// It uses [DefaultScanContext].
func Parse(s string) (Decimal, error) {
	return DefaultScanContext.Parse(s)
}

// Parse parses a string representation of a number as a [Decimal].
func (ctx Context) Parse(s string) (Decimal, error) {
	state := &scanner{reader: strings.NewReader(s)}
	var d Decimal
	if err := ctx.Scan(&d, state, 'e'); err != nil {
		return d, err
	}

	// entire string must have been consumed
	r, _, err := state.ReadRune()
	if err == nil {
		return QNaN, fmt.Errorf("expected end of string, found %c", r)
	}
	return d, nil
}

// MustParse parses a string as a [Decimal] and returns the value or
// panics if the string doesn't represent a valid [Decimal].
// It uses [DefaultScanContext].
func MustParse(s string) Decimal {
	return DefaultScanContext.MustParse(s)
}

// MustParse parses a string as a [Decimal] and returns the value or
// panics if the string doesn't represent a valid [Decimal].
func (ctx Context) MustParse(s string) Decimal {
	d, err := ctx.Parse(s)
	if err != nil {
		panic(err)
	}
	return d
}

// Scan implements [fmt.Scanner].
// It uses [DefaultScanContext].
func (d *Decimal) Scan(state fmt.ScanState, verb rune) error {
	return DefaultScanContext.Scan(d, state, verb)
}

// Scan scans a string into a [Decimal], applying context rounding.
func (ctx Context) Scan(d *Decimal, state fmt.ScanState, verb rune) error {
	*d = SNaN
	sign, err := eatRune(state, '+', '-')
	if err != nil {
		return err
	}
	if sign < 0 {
		sign = 0
	}
	// Word-number: [Ii]nf(inity)?|∞|[qs]?(nan|NaN)
	kw, err := keywords.Match(state)
	if err != nil {
		return err
	}
	switch kw {
	case 0:
	case 1:
		if sign == 0 {
			*d = Inf
		} else {
			*d = NegInf
		}
		return nil
	case 3:
		payload, _ := eatBytes(state, isDigit)
		payloadInt, _, err := ctx.parseUint(payload)
		if err != nil {
			return err
		}
		*d = newPayloadNan(sign, flQNaN, uint64(payloadInt))
		return nil
	case 2:
		payload, _ := eatBytes(state, isDigit)
		payloadInt, _, err := ctx.parseUint(payload)
		if err != nil {
			return err
		}
		*d = newPayloadNan(sign, flSNaN, uint64(payloadInt))
		return nil
	default:
		return errNotDecimal
	}

	whole, err := eatBytes(state, isDigit)
	if err != nil {
		return err
	}
	var buf [64]byte
	mantissa := append(buf[:0], whole...)

	if _, err := eatRune(state, '.', -1); err != nil {
		return err
	}

	frac, err := eatBytes(state, isDigit)
	if err != nil {
		return err
	}

	mantissa = append(mantissa, frac...)
	if len(mantissa) == 0 {
		return fmt.Errorf("mantissa missing")
	}

	e, err := eatRune(state, 'e', 'E')
	if err != nil {
		return err
	}

	var expSign int
	var exp []byte
	if e != -1 {
		expSign, err = eatRune(state, '+', '-')
		if err != nil {
			return err
		}
		if expSign < 0 {
			expSign = 0
		}
		exp, err = eatBytes(state, isDigit)
		if err != nil {
			return err
		}
		if len(exp) == 0 {
			return fmt.Errorf("exponent value missing")
		}
	}

	significand, sExp, err := ctx.parseUint(mantissa)
	if err != nil {
		return err
	}
	if significand == 0 {
		*d = zeroes[sign]
		return nil
	}

	uexponent, _, err := ctx.parseUint(exp)
	if err != nil {
		return err
	}
	exponent := int64(uexponent)
	exponent *= int64(1 - 2*expSign)
	if exponent > 1000 {
		*d = infinities[sign]
		return nil
	} else if exponent < -1000 {
		*d = zeroes[sign]
		return nil
	}
	exponent += int64(sExp - len(frac))

	partExp, partSignificand := renormalize(int16(exponent), uint64(significand))
	*d = newFromParts(int8(sign), partExp, partSignificand)
	return nil
}

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

var errNotDecimal error = Error("not a valid Decimal")

func allZeros(s []byte) bool {
	for _, c := range s {
		if c != '0' {
			return false
		}
	}
	return true
}

func (ctx Context) parseUint(s []byte) (uint64, int, error) {
	var a uint64
	var exp int
	for i, c := range s {
		if a >= decimalBase {
			switch ctx.Rounding {
			case HalfUp:
				if c >= '5' {
					a++
				}
			case HalfEven:
				if c > '5' || c == '5' && (a%2 == 1 || !allZeros(s[i+1:])) {
					a++
				}
			case Down:
			default:
				return 0, 0, fmt.Errorf("unsupported rounding mode: %v", ctx.Rounding)
			}
			exp = len(s) - i
			break
		}
		a = 10*a + uint64(c-'0')
	}
	return a, exp, nil
}

type trie struct {
	heads  []string
	tails  tries
	result int
}

func trieBranch(heads string, tails ...trie) trie {
	return trie{heads: strings.Split(heads, "|"), tails: tails}
}

func trieLeaf(heads string, result int) trie {
	return trie{heads: strings.Split(heads, "|"), result: result}
}

type tries []trie

func (tt tries) Match(state fmt.ScanState) (int, error) {
	for _, t := range tt {
	heads:
		for _, head := range t.heads {
			for i, c := range head {
				// Try to eat the head.
				r, _, err := state.ReadRune()
				if err != nil {
					if err != io.EOF {
						return 0, err
					}
					continue heads
				}
				if r != c {
					if i > 0 {
						return 0, errNotDecimal
					}
					if err := state.UnreadRune(); err != nil {
						return 0, err
					}
					continue heads
				}
			}
			if len(t.tails) == 0 {
				return t.result, nil
			}
			return t.tails.Match(state)
		}
	}
	return 0, nil
}

var keywords = tries{
	trieBranch("inf|Inf", trieLeaf("inity|", 1)),
	trieLeaf("∞", 1),
	trieBranch("s", trieLeaf("nan|NaN", 2)),
	trieBranch("q|", trieLeaf("nan|NaN", 3)),
}

func eatBytes(state fmt.ScanState, f func(r rune) bool) ([]byte, error) {
	token, err := state.Token(false, f)
	if err != nil {
		return nil, err
	}
	return token, err
}

// eatRune returns 0 if it reads a, 1 if it reads b, -1 otherwise.
func eatRune(state fmt.ScanState, a, b rune) (int, error) {
	r, _, err := state.ReadRune()
	if err != nil {
		if err != io.EOF {
			return 0, err
		}
		return -1, nil
	}
	if r == a {
		return 0, nil
	}
	if r == b {
		return 1, nil
	}
	return -1, state.UnreadRune()
}

func newPayloadNan(sign int, fl flavor, weight uint64) Decimal {
	s := uint64(sign) << 63
	switch fl {
	case flQNaN:
		return newDec(s | QNaN.bits | weight)
	case flSNaN:
		return newDec(s | SNaN.bits | weight)
	default:
		return QNaN
	}
}
