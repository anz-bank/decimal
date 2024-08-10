package decimal

import "encoding/json"

var _ json.Marshaler = Zero64
var _ json.Unmarshaler = (*Decimal64)(nil)

// MarshalText implements the encoding.TextMarshaler interface.
func (d Decimal64) MarshalJSON() ([]byte, error) {
	return d.MarshalText()
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (d *Decimal64) UnmarshalJSON(data []byte) error {
	return d.UnmarshalText(data)
}
