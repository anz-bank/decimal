package decimal

import (
	"encoding"
	"encoding/binary"
	"fmt"
)

var _ encoding.TextMarshaler = Decimal64{}
var _ encoding.TextUnmarshaler = (*Decimal64)(nil)

// MarshalText implements the encoding.TextMarshaler interface.
func (d Decimal64) MarshalText() ([]byte, error) {
	data := d.Append(make([]byte, 0, 16), 'g', -1)
	return data, nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (d *Decimal64) UnmarshalText(text []byte) error {
	e, err := Parse64(string(text))
	if err != nil {
		err = fmt.Errorf("decimal: cannot unmarshal %q as Decimal64 (%v)", text, err)
	} else {
		*d = e
	}
	return err
}

var _ encoding.BinaryMarshaler = Decimal64{}
var _ encoding.BinaryUnmarshaler = (*Decimal64)(nil)

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (d Decimal64) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, d.bits)
	return buf, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (d *Decimal64) UnmarshalBinary(data []byte) error {
	d.bits = binary.BigEndian.Uint64(data)
	// TODO: Check for out of bounds significand.
	return nil
}
