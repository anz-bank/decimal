package decimal

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew64FromInt64(t *testing.T) {
	for i := int64(-1000); i <= 1000; i++ {
		d := NewDecimal64FromInt64(i)
		j := d.Int64()
		require.EqualValues(t, i, j)
	}
}

func TestParseDecimal64(t *testing.T) {
	for i := int64(0); i <= 1000; i++ {
		s := strconv.Itoa(int(i))
		di := NewDecimal64FromInt64(i)
		ds, err := ParseDecimal64(s)
		require.NoError(t, err)
		require.Equal(t, di, ds)
	}
}

func checkDecimal64BinOp(
	t *testing.T,
	expected func(a, b int64) int64,
	actual func(a, b Decimal64) Decimal64,
	op string,
) {
	r := require.New(t)
	for i := int64(-100); i <= 100; i++ {
		a := NewDecimal64FromInt64(i)
		for j := int64(-100); j <= 100; j++ {
			b := NewDecimal64FromInt64(j)
			c := actual(a, b)
			k := c.Int64()
			e := expected(i, j)
			r.EqualValues(e, k, "%d %s %d ≠ %d (expecting %d)", i, op, j, k, e)
		}
	}
}

func TestDecimal64Add(t *testing.T) {
	checkDecimal64BinOp(t,
		func(a, b int64) int64 { return a + b },
		func(a, b Decimal64) Decimal64 { return a.Add(b) },
		"+",
	)
}

func TestDecimal64Cmp(t *testing.T) {
	r := require.New(t)
	negOne := NegOne64
	zero := Zero64
	one := One64
	r.True(negOne.Cmp(negOne) == 0)
	r.True(zero.Cmp(zero) == 0)
	r.True(one.Cmp(one) == 0)
	r.True(negOne.Cmp(zero) < 0)
	r.True(negOne.Cmp(one) < 0)
	r.True(zero.Cmp(one) < 0)
	r.True(zero.Cmp(negOne) > 0)
	r.True(one.Cmp(negOne) > 0)
	r.True(one.Cmp(zero) > 0)
}

func TestDecimal64Float64(t *testing.T) {
	r := require.New(t)
	r.Equal(NegOne64.Float64(), -1.0)
	r.Equal(Zero64.Float64(), 0.0)
	r.Equal(NegZero64.Float64(), -0.0)
	r.Equal(One64.Float64(), 1.0)
	r.Equal(NewDecimal64FromInt64(10).Float64(), 10.0)

	oneThird := One64.Quo(NewDecimal64FromInt64(3))
	one := oneThird.Add(oneThird).Add(oneThird)
	r.InEpsilon(oneThird.Float64(), 1.0/3.0, 0.00000001)
	r.Equal(one.Float64(), 1.0)
}

func TestDecimal64MulThreeByOneTenthByTen(t *testing.T) {
	r := require.New(t)

	// float 3*0.1*10 ≠ 3
	fltThree := 3.0
	fltTen := 10.0
	fltOne := 1.0
	fltOneTenth := fltOne / fltTen
	fltProduct := fltThree * fltOneTenth * fltTen
	r.Equal(fltTen*fltOneTenth, fltOne)
	r.NotEqual(fltThree, fltProduct)

	// decimal 3*0.1*10 = 3
	decThree := NewDecimal64FromInt64(3)
	decTen := NewDecimal64FromInt64(10)
	decOne := NewDecimal64FromInt64(1)
	decOneTenth := decOne.Quo(decTen)
	decProduct := decThree.Mul(decOneTenth).Mul(decTen)
	r.Equal(decTen.Mul(decOneTenth), decOne)
	r.Equal(decThree, decProduct)
}

func TestDecimal64Sub(t *testing.T) {
	checkDecimal64BinOp(t,
		func(a, b int64) int64 { return a - b },
		func(a, b Decimal64) Decimal64 { return a.Sub(b) },
		"-",
	)
}

func TestDecimal64Mul(t *testing.T) {
	checkDecimal64BinOp(t,
		func(a, b int64) int64 { return a * b },
		func(a, b Decimal64) Decimal64 { return a.Mul(b) },
		"*",
	)
}

func checkDecimal64QuoByF(t *testing.T, f int64) {
	r := require.New(t)
	for i := int64(-1000 * f); i <= 1000*f; i += f {
		for j := int64(-100); j <= 100; j++ {
			var e Decimal64
			if j == 0 {
				e = QNaN64
			} else {
				e = NewDecimal64FromInt64(i)
				if i == 0 && j < 0 {
					e = e.Neg()
				}
			}
			k := i * j
			n := NewDecimal64FromInt64(k)
			d := NewDecimal64FromInt64(j)
			q := n.Quo(d)
			r.EqualValues(e, q, "%d / %d ≠ %v (expecting %v)", k, j, q, e)
		}
	}
}

func TestDecimal64Quo(t *testing.T) {
	checkDecimal64QuoByF(t, 1)
	checkDecimal64QuoByF(t, 7)
	checkDecimal64QuoByF(t, 13)
}

func TestDecimal64MulPo10(t *testing.T) {
	r := require.New(t)
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
			e := NewDecimal64FromInt64(int64(w.lo))
			a := NewDecimal64FromInt64(int64(u.lo)).Mul(NewDecimal64FromInt64(int64(v.lo)))
			r.EqualValues(e, a, "%v * %v ≠ %v (expecting %v)", u, v, a, e)
		}
	}
}

func TestDecimal64Sqrt(t *testing.T) {
	r := require.New(t)
	for i := int64(0); i < 100000000; i = i*19/17 + 1 {
		i2 := i * i
		e := NewDecimal64FromInt64(i)
		n := NewDecimal64FromInt64(i2)
		a := n.Sqrt()
		r.EqualValues(e, a, "√%v != %v (expected %v)", n, a, e)
	}
}

func TestDecimal64SqrtNeg(t *testing.T) {
	require.EqualValues(t, QNaN64, NewDecimal64FromInt64(-1).Sqrt())
}

func TestDecimal64String(t *testing.T) {
	for i := int64(-1000); i <= 1000; i++ {
		expected := strconv.Itoa(int(i))
		d := NewDecimal64FromInt64(i)
		actual := d.String()
		require.Equal(t, expected, actual)
	}

	for f := 1; f < 100; f += 2 {
		fraction := NewDecimal64FromInt64(int64(f)).Quo(NewDecimal64FromInt64(100))
		for i := int64(0); i <= 100; i++ {
			expected := strconv.Itoa(int(i)) + "." + strconv.Itoa(100 + f)[1:3]
			d := NewDecimal64FromInt64(i).Add(fraction)
			actual := d.String()
			require.Equal(t, expected, actual)
		}
		for i := int64(-100); i < 0; i++ {
			expected := strconv.Itoa(int(i)) + "." + strconv.Itoa(100 + f)[1:3]
			d := NewDecimal64FromInt64(i).Sub(fraction)
			actual := d.String()
			require.Equal(t, expected, actual)
		}
	}
}

func TestDecimal64Format(t *testing.T) {
	for i := int64(0); i <= 1000; i++ {
		expected := strconv.Itoa(int(i))
		d := NewDecimal64FromInt64(i)
		actual := fmt.Sprintf("%v", d)
		require.Equal(t, expected, actual)
	}
}
