package decimal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecimal64Marshal(t *testing.T) {
	data, err := New64FromInt64(23456).MarshalText()
	require.NoError(t, err)
	assert.Equal(t, []byte("23456"), data)
}

func TestDecimal64Unmarshal(t *testing.T) {
	var d Decimal64
	require.NoError(t, d.UnmarshalText([]byte("23456")))
	assert.Equal(t, New64FromInt64(23456), d)
}

func TestDecimal64UnmarshalBadInput(t *testing.T) {
	var d Decimal64
	assert.Error(t, d.UnmarshalText([]byte("omg")))
}
