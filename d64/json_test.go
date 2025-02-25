package d64

import (
	"encoding/json"
	"testing"
)

func TestDecimalMarshalJSON(t *testing.T) {
	t.Parallel()

	j, err := json.Marshal(MustParse("123.432"))
	isnil(t, err)
	equal(t, "123.432", string(j))
}

func TestDecimalUnmarshalJSON(t *testing.T) {
	t.Parallel()

	var d Decimal
	isnil(t, json.Unmarshal([]byte("23456"), &d))
	equal(t, NewFromInt64(23456), d)
}

func TestDecimalUnmarshalBadInputJSON(t *testing.T) {
	t.Parallel()

	var d Decimal
	notnil(t, json.Unmarshal([]byte("omg"), &d))
}
