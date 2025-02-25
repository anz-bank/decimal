package d64

import "testing"

func TestPartsInf(t *testing.T) {
	t.Parallel()

	var a decParts
	a.unpack(Inf)
	check(t, a.fl == flInf)

	a.unpack(NegInf)
	check(t, a.fl == flInf)
}

func TestIsNaN(t *testing.T) {
	t.Parallel()

	var a decParts
	a.unpack(Zero)
	check(t, !a.fl.nan())

	a.unpack(SNaN)
	check(t, a.fl == flSNaN)
}

func TestPartsSubnormal(t *testing.T) {
	t.Parallel()

	d := MustParse("0.1E-383")
	var subnormalParts decParts
	subnormalParts.unpack(d)
	check(t, subnormalParts.isSubnormal())

	e := NewFromInt64(42)
	var fortyTwoParts decParts
	fortyTwoParts.unpack(e)
	check(t, !fortyTwoParts.isSubnormal())

}
