package decimal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecimal64Marshal(t *testing.T) {
	require.Equal(t, []byte("23456"), New64FromInt64(23456).MarshalText())
}

func TestDecimal64Unmarshal(t *testing.T) {
	var d Decimal64
	require.NoError(t, d.UnmarshalText([]byte("23456")))
	require.Equal(t, New64FromInt64(23456), d)
}

func TestDecimal64UnmarshalBadInput(t *testing.T) {
	var d Decimal64
	require.Error(t, d.UnmarshalText([]byte("omg")))
}
