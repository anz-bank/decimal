package decimal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew64FromInt64(t *testing.T) {
	for i := int64(-1000); i <= 1000; i++ {
		d := New64FromInt64(i)
		j := d.Int64()
		require.EqualValues(t, i, j)
	}
}

func testBinOp64(
	t *testing.T,
	expected func(a, b int64) int64,
	actual func(a, b Decimal64) Decimal64,
	op string,
) {
	for i := int64(-100); i <= 100; i++ {
		a := New64FromInt64(i)
		for j := int64(-100); j <= 100; j++ {
			b := New64FromInt64(j)
			c := actual(a, b)
			k := c.Int64()
			e := expected(i, j)
			require.EqualValues(t, e, k, "%d %s %d ≠ %d (expecting %d)", i, op, j, k, e)
		}
	}
}

func TestAdd64(t *testing.T) {
	testBinOp64(t,
		func(a, b int64) int64 { return a + b },
		func(a, b Decimal64) Decimal64 { return a.Add(b) },
		"+",
	)
}

func TestSub64(t *testing.T) {
	testBinOp64(t,
		func(a, b int64) int64 { return a - b },
		func(a, b Decimal64) Decimal64 { return a.Sub(b) },
		"-",
	)
}

func TestMul64(t *testing.T) {
	testBinOp64(t,
		func(a, b int64) int64 { return a * b },
		func(a, b Decimal64) Decimal64 { return a.Mul(b) },
		"*",
	)
}

func TestMul64_po10(t *testing.T) {
	for i, u := range powersOf10 {
		for j, v := range powersOf10 {
			k := i + j
			if !(k < len(powersOf10)) {
				continue
			}
			w := powersOf10[k]
			if !(w.hi == 0 && w.lo < 0x8000000000000000) {
				continue
			}
			e := New64FromInt64(int64(w.lo))
			a := New64FromInt64(int64(u.lo)).Mul(New64FromInt64(int64(v.lo)))
			require.EqualValues(t, e, a, "%v * %v ≠ %v (expecting %v)", u, v, a, e)
		}
	}
}
