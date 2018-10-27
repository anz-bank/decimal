package decimal

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// ParseDecimal64 parses a string representation of a number as a Decimal64.
func ParseDecimal64(s string) (Decimal64, error) {
	r := strings.NewReader(s)
	d := Zero64
	if err := d.scan(r); err != nil {
		return d, err
	}

	// entire string must have been consumed
	if ch, err := r.ReadByte(); err == nil {
		return d, fmt.Errorf("expected end of string, found %q", ch)
	} else if err != io.EOF {
		return d, err
	}
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
	state.SkipSpace()
	return d.scan(byteReader{state})
}

func notDecimal64() error {
	return fmt.Errorf("Not a valid Decimal64")
}

func scanRule(r io.ByteScanner, match func(b byte) bool) (string, error) {
	var buf bytes.Buffer
	for {
		b, err := r.ReadByte()
		if err != nil {
			if err != io.EOF {
				return "", err
			}
			break
		}
		// A dirty hack to recognise ∞, which UTF-8-encodes as [226, 136, 158]
		if !match(b) {
			if err := r.UnreadByte(); err != nil {
				return "", err
			}
			break
		}
		buf.WriteByte(b)
	}
	return string(buf.Bytes()), nil
}

func scanWord(r io.ByteScanner) (string, error) {
	return scanRule(r, func(b byte) bool {
		return 'A' <= b && b <= 'Z' || 'a' <= b && b <= 'z' || b == 226 || b == 136 || b == 158
	})
}

func scanDigits(r io.ByteScanner) (string, error) {
	return scanRule(r, func(b byte) bool {
		return '0' <= b && b <= '9'
	})
}

func scanByte(r io.ByteScanner, match func(b byte) bool) (byte, error) {
	b, err := r.ReadByte()
	if err != nil {
		if err != io.EOF {
			return 0, err
		}
		return 0, nil
	}
	if match(b) {
		return b, nil
	}
	if err := r.UnreadByte(); err != nil {
		return 0, err
	}
	return 0, nil
}

func scanSign(r io.ByteScanner) (int, error) {
	s, err := scanByte(r, func(b byte) bool { return b == '-' || b == '+' })
	if err != nil {
		return 0, err
	}
	if s == '-' {
		return 1, nil
	}
	return 0, nil
}

func parseUint(s string) (int64, int) {
	var a int64
	var exp int
	for i, c := range s {
		if a >= (1<<63-9)/10 {
			exp = len(s) - i
			break
		}
		a = 10*a + int64(c-'0')
	}
	return a, exp
}

func (d *Decimal64) scan(r io.ByteScanner) error {
	sign, err := scanSign(r)

	// Word-number ([Ii]nf|∞|nan|NaN)
	word, err := scanWord(r)
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
	default:
		return notDecimal64()
	}

	whole, err := scanDigits(r)
	if err != nil {
		return err
	}

	if _, err = scanByte(r, func(b byte) bool { return b == '.' }); err != nil {
		return err
	}

	frac, err := scanDigits(r)
	if err != nil {
		return err
	}

	e, err := scanByte(r, func(b byte) bool { return b == 'e' || b == 'E' })
	if err != nil {
		return err
	}

	var expSign int
	var exp string
	if e != 0 {
		expSign, err = scanSign(r)
		if err != nil {
			return err
		}
		exp, err = scanDigits(r)
		if err != nil {
			return err
		}
	}

	mantissa := strings.TrimLeft(whole+frac, "0")
	if mantissa == "" {
		mantissa = "0"
	}

	significand, sExp := parseUint(mantissa)

	exponent, _ := parseUint(exp)
	if exponent > 1000 {
		exponent = 1000
	}
	exponent *= int64(1 - 2*expSign)
	exponent += int64(sExp - len(frac))

	partExp, partSignificand := renormalize(int(exponent), uint64(significand))
	*d = newFromParts(sign, partExp, partSignificand)
	return nil
}
