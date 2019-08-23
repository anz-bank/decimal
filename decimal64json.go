package decimal

// MarshalText implements the encoding.TextMarshaler interface.
func (d Decimal64) MarshalJSON() ([]byte, error) {
	return d.MarshalText(), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (d *Decimal64) UnmarshalJSON(data []byte) error {
	return d.UnmarshalText(data)
}
