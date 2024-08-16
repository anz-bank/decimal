package decimal

import (
	"fmt"
	"strings"
	"unicode"
)

var DefaultScanContext64 = DefaultFormatContext64

// Parse64 parses a string representation of a number as a Decimal64.
// It uses [DefaultScanContext64].
func Parse64(s string) (Decimal64, error) {
	return DefaultScanContext64.Parse(s)
}

// Parse64 parses a string representation of a number as a Decimal64.
func (ctx Context64) Parse(s string) (Decimal64, error) {
	state := stringScanner{reader: strings.NewReader(s)}
	var d Decimal64
	if err := ctx.Scan(&d, &state, 'e'); err != nil {
		return d, err
	}

	// entire string must have been consumed
	r, _, err := state.ReadRune()
	if err == nil {
		return d, fmt.Errorf("expected end of string, found %c", r)
	}
	return d, nil
}

// MustParse64 parses a string as a Decimal64 and returns the value or
// panics if the string doesn't represent a valid Decimal64.
// It uses [DefaultScanContext64].
func MustParse64(s string) Decimal64 {
	return DefaultScanContext64.MustParse(s)
}

// MustParse64 parses a string as a Decimal64 and returns the value or
// panics if the string doesn't represent a valid Decimal64.
func (ctx Context64) MustParse(s string) Decimal64 {
	d, err := ctx.Parse(s)
	if err != nil {
		panic(err)
	}
	return d
}

// Scan implements fmt.Scanner.
// It uses [DefaultScanContext64].
func (d *Decimal64) Scan(state fmt.ScanState, verb rune) error {
	return DefaultScanContext64.Scan(d, state, verb)
}

// Scan scans a string into a Decimal64, applying context rounding.
func (ctx Context64) Scan(d *Decimal64, state fmt.ScanState, verb rune) error {
	*d = SNaN64
	sign, err := scanSign(state)
	if err != nil {
		return err
	}
	// Word-number ([Ii]nf|∞|nan|NaN)
	word, err := tokenString(state, isLetterOrInf)
	if err != nil {
		return err
	}
	switch strings.ToLower(word) {
	case "":
	case "inf", "infinity", "∞":
		if sign == 0 {
			*d = Infinity64
		} else {
			*d = NegInfinity64
		}
		return nil
	case "nan", "qnan":
		payload, _ := tokenString(state, unicode.IsDigit)
		payloadInt, _, err := ctx.parseUint(payload)
		if err != nil {
			return err
		}
		*d = newPayloadNan(sign, flQNaN, uint64(payloadInt))
		return nil
	case "snan":
		payload, _ := tokenString(state, unicode.IsDigit)
		payloadInt, _, err := ctx.parseUint(payload)
		if err != nil {
			return err
		}
		*d = newPayloadNan(sign, flSNaN, uint64(payloadInt))
		return nil
	default:
		return notDecimal64()
	}

	whole, err := tokenString(state, unicode.IsDigit)
	if err != nil {
		return err
	}

	dot, err := tokenString(state, func(r rune) bool { return r == '.' })
	if err != nil {
		return err
	}
	if len(dot) > 1 {
		return fmt.Errorf("too many dots")
	}

	frac, err := tokenString(state, unicode.IsDigit)
	if err != nil {
		return err
	}

	e, err := tokenString(state, func(r rune) bool { return r == 'e' || r == 'E' })
	if err != nil {
		return err
	}
	if len(e) > 1 {
		return fmt.Errorf("too many 'e's")
	}

	var expSign int
	var exp string
	if len(e) == 1 {
		expSign, err = scanSign(state)
		if err != nil {
			return err
		}
		exp, err = tokenString(state, unicode.IsDigit)
		if err != nil {
			return err
		}
		if exp == "" {
			return fmt.Errorf("exponent value missing")
		}
	}

	mantissa := whole + frac
	if mantissa == "" {
		return fmt.Errorf("mantissa missing")
	}

	significand, sExp, err := ctx.parseUint(mantissa)
	if err != nil {
		return err
	}
	if significand == 0 {
		*d = zeroes64[sign]
		return nil
	}

	uexponent, _, err := ctx.parseUint(exp)
	if err != nil {
		return err
	}
	exponent := int64(uexponent)
	exponent *= int64(1 - 2*expSign)
	if exponent > 1000 {
		*d = infinities64[sign]
		return nil
	} else if exponent < -1000 {
		*d = zeroes64[sign]
		return nil
	}
	exponent += int64(sExp - len(frac))

	partExp, partSignificand := renormalize(int(exponent), uint64(significand))
	*d = newFromParts(sign, partExp, partSignificand)
	return nil
}

func notDecimal64() error {
	return fmt.Errorf("not a valid Decimal64")
}

func allZeros(s string) bool {
	for _, c := range s {
		if c != '0' {
			return false
		}
	}
	return true
}

func (ctx Context64) parseUint(s string) (uint64, int, error) {
	var a uint64
	var exp int
	for i, c := range s {
		if a >= decimal64Base {
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

func tokenString(state fmt.ScanState, f func(r rune) bool) (string, error) {
	token, err := state.Token(false, f)
	if err != nil {
		return "", err
	}
	return string(token), err
}

func scanSign(state fmt.ScanState) (int, error) {
	s, err := state.Token(false, func(r rune) bool { return r == '-' || r == '+' })
	if err != nil {
		return 0, err
	}
	switch len(s) {
	case 0:
		// Implied '+'
	case 1:
		if s[0] == '-' {
			return 1, nil
		}
	default:
		return 0, fmt.Errorf("too many +/- characters: %s", string(s))
	}
	return 0, nil
}

func newPayloadNan(sign int, fl flavor, weight uint64) Decimal64 {
	s := uint64(sign) << 63
	switch fl {
	case flQNaN:
		return new64(s | QNaN64.bits | weight)
	case flSNaN:
		return new64(s | SNaN64.bits | weight)
	default:
		return QNaN64
	}
}
