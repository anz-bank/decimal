package d64

import "testing"

func TestDecimalGob(t *testing.T) {
	t.Parallel()

	gob, err := NewFromInt64(23456).GobEncode()
	isnil(t, err)

	var d Decimal
	isnil(t, d.GobDecode(gob))
	equal(t, NewFromInt64(23456), d)
}
