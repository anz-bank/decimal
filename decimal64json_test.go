package decimal

import (
	"encoding/json"
	"testing"
)

func TestDecimal64MarshalJSON(t *testing.T) {
	t.Parallel()

	j, err := json.Marshal(MustParse64("123.432"))
	isnil(t, err)
	equal(t, "123.432", string(j))
}

func TestDecimal64UnmarshalJSON(t *testing.T) {
	t.Parallel()

	var d Decimal64
	isnil(t, json.Unmarshal([]byte("23456"), &d))
	equal(t, New64FromInt64(23456), d)
}

func TestDecimal64UnmarshalBadInputJSON(t *testing.T) {
	t.Parallel()

	var d Decimal64
	notnil(t, json.Unmarshal([]byte("omg"), &d))
}
