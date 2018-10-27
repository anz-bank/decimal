package decimal

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

// ParseDecimal64 parses a string representation of a number as a Decimal64.
func ParseDecimal64(s string) (Decimal64, error) {
	state := stringScanner{reader: strings.NewReader(s)}
	var d Decimal64
	if err := d.Scan(&state, 'e'); err != nil {
		return d, err
	}

	// entire string must have been consumed
	r, _, err := state.ReadRune()
	if err == nil {
		return d, fmt.Errorf("expected end of string, found %c", r)
	}
	logicCheck(err == io.EOF, "%v == io.EOF", err)
	return d, nil
}

// MustParseDecimal64 parses a string as a Decimal64 and returns the value or
// panics if the string doesn't represent a valid Decimal64.
func MustParseDecimal64(s string) Decimal64 {
	d, err := ParseDecimal64(s)
	if err != nil {
		panic(err)
	}
	return d
}

// Scan implements fmt.Scanner.
func (d *Decimal64) Scan(state fmt.ScanState, verb rune) error {
	sign, err := scanSign(state)
	if err != nil {
		return err
	}

	// Word-number ([Ii]nf|∞|nan|NaN)
	word, err := tokenString(state, isLetterOrInf)
	if err != nil {
		return err
	}
	switch word {
	case "":
	case "inf", "Inf", "∞":
		if sign == 0 {
			*d = Infinity64
		} else {
			*d = NegInfinity64
		}
		return nil
	case "nan", "NaN":
		*d = QNaN64
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
		return fmt.Errorf("Too many dots")
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
		return fmt.Errorf("Too many 'e's")
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
			return fmt.Errorf("Exponent value missing")
		}
	}

	mantissa := whole + frac
	if mantissa == "" {
		return fmt.Errorf("Mantissa missing")
	}
	mantissa = strings.TrimLeft(mantissa, "0")
	if mantissa == "" {
		mantissa = "0"
	}

	significand, sExp := parseUint(mantissa)
	if significand == 0 {
		*d = zeroes[sign]
		return nil
	}

	exponent, _ := parseUint(exp)
	exponent *= int64(1 - 2*expSign)
	if exponent > 1000 {
		*d = infinities[sign]
		return nil
	} else if exponent < -1000 {
		*d = zeroes[sign]
		return nil
	}
	exponent += int64(sExp - len(frac))

	partExp, partSignificand := renormalize(int(exponent), uint64(significand))
	*d = newFromParts(sign, partExp, partSignificand)
	return nil
}

func notDecimal64() error {
	return fmt.Errorf("Not a valid Decimal64")
}

func parseUint(s string) (int64, int) {
	var a int64
	var exp int
	for i, c := range s {
		if a >= decimal64Base {
			exp = len(s) - i
			break
		}
		a = 10*a + int64(c-'0')
	}
	return a, exp
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
		return 0, fmt.Errorf("Too many +/- characters: %s", string(s))
	}
	return 0, nil
}
