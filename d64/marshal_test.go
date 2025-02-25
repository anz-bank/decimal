package d64

import "testing"

func TestDecimalMarshal(t *testing.T) {
	t.Parallel()

	data, err := NewFromInt64(23456).MarshalText()
	isnil(t, err)
	equal(t, "23456", string(data))
}

func TestDecimalUnmarshal(t *testing.T) {
	t.Parallel()

	var d Decimal
	isnil(t, d.UnmarshalText([]byte("23456")))
	equal(t, NewFromInt64(23456), d)
}

func TestDecimalUnmarshalBadInput(t *testing.T) {
	t.Parallel()

	var d Decimal
	notnil(t, d.UnmarshalText([]byte("omg")))
}
