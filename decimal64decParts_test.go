package decimal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPartsInf(t *testing.T) {
	a := decParts{}
	a.unpack(Infinity64)
	require.True(t, a.isInf())

	a.unpack(NegInfinity64)
	require.True(t, a.isInf())
}

func TestIsNaN(t *testing.T) {
	require := require.New(t)

	a := decParts{}
	a.unpack(Zero64)
	require.Equal(false, a.isNaN())

	a.unpack(SNaN64)
	require.Equal(true, a.isSNaN())

	a.unpack(QNaN64)
	require.Equal(true, a.isQNaN())
}

func TestPartsSubnormal(t *testing.T) {
	require := require.New(t)

	d := MustParseDecimal64("0.1E-383")
	subnormal64Parts := decParts{}
	subnormal64Parts.unpack(d)
	require.Equal(true, subnormal64Parts.isSubnormal())

	e := NewDecimal64FromInt64(42)
	fortyTwoParts := decParts{}
	fortyTwoParts.unpack(e)
	require.Equal(false, fortyTwoParts.isSubnormal())

}
