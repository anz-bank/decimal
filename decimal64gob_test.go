package decimal

import "testing"

func TestDecimal64Gob(t *testing.T) {
	t.Parallel()

	gob, err := New64FromInt64(23456).GobEncode()
	isnil(t, err)

	var d Decimal64
	isnil(t, d.GobDecode(gob))
	equal(t, New64FromInt64(23456), d)
}
