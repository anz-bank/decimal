package decimal

import "testing"

func TestPartsInf(t *testing.T) {
	t.Parallel()

	var a decParts
	a.unpack(Infinity64)
	check(t, a.fl == flInf)

	a.unpack(NegInfinity64)
	check(t, a.fl == flInf)
}

func TestIsNaN(t *testing.T) {
	t.Parallel()

	var a decParts
	a.unpack(Zero64)
	check(t, !a.fl.nan())

	a.unpack(SNaN64)
	check(t, a.fl == flSNaN)
}

func TestPartsSubnormal(t *testing.T) {
	t.Parallel()

	d := MustParse64("0.1E-383")
	var subnormal64Parts decParts
	subnormal64Parts.unpack(d)
	check(t, subnormal64Parts.isSubnormal())

	e := New64FromInt64(42)
	var fortyTwoParts decParts
	fortyTwoParts.unpack(e)
	check(t, !fortyTwoParts.isSubnormal())

}
