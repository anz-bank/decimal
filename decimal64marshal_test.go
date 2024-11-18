package decimal

import "testing"

func TestDecimal64Marshal(t *testing.T) {
	t.Parallel()

	data, err := New64FromInt64(23456).MarshalText()
	isnil(t, err)
	equal(t, "23456", string(data))
}

func TestDecimal64Unmarshal(t *testing.T) {
	t.Parallel()

	var d Decimal64
	isnil(t, d.UnmarshalText([]byte("23456")))
	equal(t, New64FromInt64(23456), d)
}

func TestDecimal64UnmarshalBadInput(t *testing.T) {
	t.Parallel()

	var d Decimal64
	notnil(t, d.UnmarshalText([]byte("omg")))
}
