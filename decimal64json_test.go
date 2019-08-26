package decimal

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecimal64MarshalJSON(t *testing.T) {
	j, err := json.Marshal(MustParseDecimal64("123.432"))
	require.NoError(t, err)
	require.Equal(t, []byte("123.432"), j)
}

func TestDecimal64UnmarshalJSON(t *testing.T) {
	var d Decimal64
	require.NoError(t, json.Unmarshal([]byte("23456"), &d))
	require.Equal(t, NewDecimal64FromInt64(23456), d)
}

func TestDecimal64UnmarshalBadInputJSON(t *testing.T) {
	var d Decimal64
	require.Error(t, json.Unmarshal([]byte("omg"), &d))
}
