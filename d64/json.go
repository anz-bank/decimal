package d64

import "encoding/json"

var _ json.Marshaler = Zero
var _ json.Unmarshaler = (*Decimal)(nil)

// MarshalText implements the encoding.TextMarshaler interface.
func (d Decimal) MarshalJSON() ([]byte, error) {
	return d.MarshalText()
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (d *Decimal) UnmarshalJSON(data []byte) error {
	return d.UnmarshalText(data)
}
