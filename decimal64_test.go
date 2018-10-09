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

func TestMulThreeByOneTenthByTen(t *testing.T) {
	// float 3*0.1*10 ≠ 3
	fltThree := 3.0
	fltTen := 10.0
	fltOne := 1.0
	fltOneTenth := fltOne / fltTen
	fltProduct := fltThree * fltOneTenth * fltTen
	require.Equal(t, fltTen*fltOneTenth, fltOne)
	require.NotEqual(t, fltThree, fltProduct)

	// decimal 3*0.1*10 = 3
	decThree := New64FromInt64(3)
	decTen := New64FromInt64(10)
	decOne := New64FromInt64(1)
	decOneTenth := decOne.Div(decTen)
	decProduct := decThree.Mul(decOneTenth).Mul(decTen)
	require.Equal(t, decTen.Mul(decOneTenth), decOne)
	require.Equal(t, decThree, decProduct)
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

func requireDiv64ByF(t *testing.T, f int64) {
	for i := int64(-1000 * f); i <= 1000*f; i += f {
		for j := int64(-100); j <= 100; j++ {
			var e Decimal64
			if j == 0 {
				e = QNaN64
			} else {
				e = New64FromInt64(i)
				if i == 0 && j < 0 {
					e = e.Neg()
				}
			}
			k := i * j
			n := New64FromInt64(k)
			d := New64FromInt64(j)
			q := n.Div(d)
			require.EqualValues(t, e, q, "%d / %d ≠ %v (expecting %v)", k, j, q, e)
		}
	}
}

func TestDiv64(t *testing.T) {
	requireDiv64ByF(t, 1)
	requireDiv64ByF(t, 7)
	requireDiv64ByF(t, 13)
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
