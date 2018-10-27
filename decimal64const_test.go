package decimal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPi64(t *testing.T) {
	require.Equal(t, "3.141592653589793", Pi64.String())
}

func TestE64(t *testing.T) {
	require.Equal(t, "2.718281828459045", E64.String())
}
