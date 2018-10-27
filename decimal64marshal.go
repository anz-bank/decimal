package decimal

import (
	"fmt"
)

// MarshalText implements the encoding.TextMarshaler interface.
func (d Decimal64) MarshalText() []byte {
	var buf []byte
	return d.Append(buf, 'g', -1)
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (d *Decimal64) UnmarshalText(text []byte) error {
	e, err := ParseDecimal64(string(text))
	if err != nil {
		err = fmt.Errorf("decimal: cannot unmarshal %q as Decimal64 (%v)", text, err)
	} else {
		*d = e
	}
	return err
}
