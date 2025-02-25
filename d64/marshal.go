package d64

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"fmt"
)

var _ encoding.TextMarshaler = Zero
var _ encoding.TextUnmarshaler = (*Decimal)(nil)

// MarshalText implements the encoding.TextMarshaler interface.
func (d Decimal) MarshalText() ([]byte, error) {
	return d.Append(nil, 'g', -1), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (d *Decimal) UnmarshalText(text []byte) error {
	state := &scanner{reader: bytes.NewReader(text)}
	var e Decimal
	if err := DefaultContext.Scan(&e, state, 'e'); err != nil {
		return err
	}

	r, _, err := state.ReadRune()
	if err == nil {
		return fmt.Errorf("expected end of text, found %c", r)
	}

	*d = e
	return nil
}

var _ encoding.BinaryMarshaler = Zero
var _ encoding.BinaryUnmarshaler = (*Decimal)(nil)

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (d Decimal) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, d.bits)
	return buf, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (d *Decimal) UnmarshalBinary(data []byte) error {
	bits := binary.BigEndian.Uint64(data)
	// TODO: Check for out of bounds significand.
	*d = newDec(bits)
	return nil
}
