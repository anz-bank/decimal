package decimal

import (
	"github.com/stretchr/testify/require"
	"testing"
)

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

func TestDecimal64Abs(t *testing.T) {
	require := require.New(t)

	require.Equal(Zero64, Zero64.Abs())
	require.Equal(Zero64, NegZero64.Abs())
	require.Equal(Infinity64, Infinity64.Abs())
	require.Equal(Infinity64, NegInfinity64.Abs())

	fortyTwo := NewDecimal64FromInt64(42)
	require.Equal(fortyTwo, fortyTwo.Abs())
	require.Equal(fortyTwo, NewDecimal64FromInt64(-42).Abs())
}

func TestDecimal64Add(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestDecimal64Add in short mode.")
	}
	checkDecimal64BinOp(t,
		func(a, b int64) int64 { return a + b },
		func(a, b Decimal64) Decimal64 { return a.Add(b) },
		"+",
	)
}

func TestDecimal64AddNaN(t *testing.T) {
	require := require.New(t)

	fortyTwo := NewDecimal64FromInt64(42)

	require.Equal(QNaN64, fortyTwo.Add(QNaN64))
	require.Equal(QNaN64, QNaN64.Add(fortyTwo))

	require.Panics(func() { fortyTwo.Add(SNaN64) })
	require.Panics(func() { SNaN64.Add(fortyTwo) })
}

func TestDecimal64AddInf(t *testing.T) {
	require := require.New(t)

	fortyTwo := NewDecimal64FromInt64(42)

	require.Equal(Infinity64, fortyTwo.Add(Infinity64))
	require.Equal(Infinity64, Infinity64.Add(fortyTwo))

	require.Equal(NegInfinity64, fortyTwo.Add(NegInfinity64))
	require.Equal(NegInfinity64, NegInfinity64.Add(fortyTwo))

	require.Equal(Infinity64, Infinity64.Add(Infinity64))
	require.Equal(NegInfinity64, NegInfinity64.Add(NegInfinity64))

	require.Equal(QNaN64, Infinity64.Add(NegInfinity64))
	require.Equal(QNaN64, NegInfinity64.Add(Infinity64))
}

func TestDecimal64Cmp(t *testing.T) {
	require := require.New(t)

	require.Equal(0, NegOne64.Cmp(NegOne64))

	require.Equal(0, Zero64.Cmp(Zero64))
	require.Equal(0, Zero64.Cmp(NegZero64))
	require.Equal(0, NegZero64.Cmp(Zero64))
	require.Equal(0, NegZero64.Cmp(NegZero64))

	require.Equal(0, One64.Cmp(One64))
	require.Equal(-1, NegOne64.Cmp(Zero64))
	require.Equal(-1, NegOne64.Cmp(NegZero64))
	require.Equal(-1, NegOne64.Cmp(One64))
	require.Equal(-1, Zero64.Cmp(One64))
	require.Equal(-1, NegZero64.Cmp(One64))
	require.Equal(1, Zero64.Cmp(NegOne64))
	require.Equal(1, NegZero64.Cmp(NegOne64))
	require.Equal(1, One64.Cmp(NegOne64))
	require.Equal(1, One64.Cmp(Zero64))
	require.Equal(1, One64.Cmp(NegZero64))
}

func TestDecimal64CmpNaN(t *testing.T) {
	require := require.New(t)

	require.Equal(-2, QNaN64.Cmp(QNaN64))
	require.Equal(-2, Zero64.Cmp(QNaN64))
	require.Equal(-2, QNaN64.Cmp(Zero64))

	require.Panics(func() { SNaN64.Cmp(SNaN64) })
	require.Panics(func() { SNaN64.Cmp(QNaN64) })
	require.Panics(func() { QNaN64.Cmp(SNaN64) })
	require.Panics(func() { SNaN64.Cmp(Zero64) })
	require.Panics(func() { Zero64.Cmp(SNaN64) })
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

func TestDecimal64Mul(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestDecimal64Mul in short mode.")
	}
	checkDecimal64BinOp(t,
		func(a, b int64) int64 { return a * b },
		func(a, b Decimal64) Decimal64 { return a.Mul(b) },
		"*",
	)
}

func TestDecimal64MulNaN(t *testing.T) {
	require := require.New(t)

	fortyTwo := NewDecimal64FromInt64(42)

	require.Equal(QNaN64, fortyTwo.Mul(QNaN64))
	require.Equal(QNaN64, QNaN64.Mul(fortyTwo))

	require.Panics(func() { fortyTwo.Mul(SNaN64) })
	require.Panics(func() { SNaN64.Mul(fortyTwo) })
}

func TestDecimal64MulInf(t *testing.T) {
	require := require.New(t)

	fortyTwo := NewDecimal64FromInt64(42)
	negFortyTwo := NewDecimal64FromInt64(-42)

	require.Equal(Infinity64, fortyTwo.Mul(Infinity64))
	require.Equal(Infinity64, Infinity64.Mul(fortyTwo))
	require.Equal(NegInfinity64, negFortyTwo.Mul(Infinity64))
	require.Equal(NegInfinity64, Infinity64.Mul(negFortyTwo))

	require.Equal(NegInfinity64, fortyTwo.Mul(NegInfinity64))
	require.Equal(NegInfinity64, NegInfinity64.Mul(fortyTwo))
	require.Equal(Infinity64, negFortyTwo.Mul(NegInfinity64))
	require.Equal(Infinity64, NegInfinity64.Mul(negFortyTwo))

	require.Equal(Infinity64, Infinity64.Mul(Infinity64))
	require.Equal(Infinity64, NegInfinity64.Mul(NegInfinity64))
	require.Equal(NegInfinity64, Infinity64.Mul(NegInfinity64))
	require.Equal(NegInfinity64, NegInfinity64.Mul(Infinity64))
}

func checkDecimal64QuoByF(t *testing.T, f int64) {
	require := require.New(t)
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
			if q != e {
				eFlavor, eSign, eExp, eSignificand := q.parts()
				qFlavor, qSign, qExp, qSignificand := q.parts()
				t.Log("e", e.bits, eFlavor, eSign, eExp, eSignificand)
				t.Log("q", q.bits, qFlavor, qSign, qExp, qSignificand)
			}
			require.Equal(e, q, "%d / %d ≠ %v (expecting %v)", k, j, q, e)
		}
	}
}

func TestDecimal64Quo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestDecimal64Quo in short mode.")
	}
	checkDecimal64QuoByF(t, 1)
	checkDecimal64QuoByF(t, 7)
	checkDecimal64QuoByF(t, 13)
}

func TestDecimal64QuoNaN(t *testing.T) {
	require := require.New(t)

	fortyTwo := NewDecimal64FromInt64(42)

	require.Equal(QNaN64, fortyTwo.Quo(QNaN64))
	require.Equal(QNaN64, QNaN64.Quo(fortyTwo))

	require.Panics(func() { fortyTwo.Quo(SNaN64) })
	require.Panics(func() { SNaN64.Quo(fortyTwo) })
}

func TestDecimal64QuoInf(t *testing.T) {
	require := require.New(t)

	fortyTwo := NewDecimal64FromInt64(42)
	negFortyTwo := NewDecimal64FromInt64(-42)

	require.Equal(Zero64, fortyTwo.Quo(Infinity64))
	require.Equal(Infinity64, Infinity64.Quo(fortyTwo))
	require.Equal(NegZero64, negFortyTwo.Quo(Infinity64))
	require.Equal(NegInfinity64, Infinity64.Quo(negFortyTwo))

	require.Equal(NegZero64, fortyTwo.Quo(NegInfinity64))
	require.Equal(NegInfinity64, NegInfinity64.Quo(fortyTwo))
	require.Equal(Zero64, negFortyTwo.Quo(NegInfinity64))
	require.Equal(Infinity64, NegInfinity64.Quo(negFortyTwo))

	require.Equal(QNaN64, Infinity64.Quo(Infinity64))
	require.Equal(QNaN64, NegInfinity64.Quo(NegInfinity64))
	require.Equal(QNaN64, Infinity64.Quo(NegInfinity64))
	require.Equal(QNaN64, NegInfinity64.Quo(Infinity64))
}

func TestDecimal64MulPo10(t *testing.T) {
	r := require.New(t)
	for i, u := range powersOf10U128 {
		for j, v := range powersOf10U128 {
			k := i + j
			if !(k < len(powersOf10U128)) {
				continue
			}
			w := powersOf10U128[k]
			if !(w.hi == 0 && w.lo < decimal64Base) {
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

func TestDecimal64SqrtNaN(t *testing.T) {
	require := require.New(t)

	require.Equal(QNaN64, QNaN64.Sqrt())
	require.Panics(func() { SNaN64.Sqrt() })
}

func TestDecimal64SqrtInf(t *testing.T) {
	require := require.New(t)

	require.Equal(Infinity64, Infinity64.Sqrt())
	require.Equal(QNaN64, NegInfinity64.Sqrt())
}

func TestDecimal64Sub(t *testing.T) {
	checkDecimal64BinOp(t,
		func(a, b int64) int64 { return a - b },
		func(a, b Decimal64) Decimal64 { return a.Sub(b) },
		"-",
	)
}

func benchmarkDecimal64Data() []Decimal64 {
	return []Decimal64{
		One64,
		QNaN64,
		Infinity64,
		NegInfinity64,
		Pi64,
		E64,
		NewDecimal64FromInt64(42),
		MustParseDecimal64("2345678e100"),
		NewDecimal64FromInt64(1234567),
		NewDecimal64FromInt64(-42),
		MustParseDecimal64("3456789e-120"),
	}
}

func BenchmarkDecimal64Abs(b *testing.B) {
	x := benchmarkDecimal64Data()
	for i := 0; i <= b.N; i++ {
		_ = x[i%len(x)].Abs()
	}
}

func BenchmarkDecimal64Add(b *testing.B) {
	x := benchmarkDecimal64Data()
	y := x[:len(x)-2]
	for i := 0; i <= b.N; i++ {
		_ = x[i%len(x)].Add(y[i%len(y)])
	}
}

func BenchmarkDecimal64Cmp(b *testing.B) {
	x := benchmarkDecimal64Data()
	y := x[:len(x)-2]
	for i := 0; i <= b.N; i++ {
		_ = x[i%len(x)].Cmp(y[i%len(y)])
	}
}

func BenchmarkDecimal64Mul(b *testing.B) {
	x := benchmarkDecimal64Data()
	y := x[:len(x)-2]
	for i := 0; i <= b.N; i++ {
		_ = x[i%len(x)].Mul(y[i%len(y)])
	}
}

func BenchmarkDecimal64Quo(b *testing.B) {
	x := benchmarkDecimal64Data()
	y := x[:len(x)-2]
	for i := 0; i <= b.N; i++ {
		_ = x[i%len(x)].Quo(y[i%len(y)])
	}
}

func BenchmarkDecimal64Sqrt(b *testing.B) {
	x := benchmarkDecimal64Data()
	for i := 0; i <= b.N; i++ {
		_ = x[i%len(x)].Sqrt()
	}
}

func BenchmarkDecimal64Sub(b *testing.B) {
	x := benchmarkDecimal64Data()
	y := x[:len(x)-2]
	for i := 0; i <= b.N; i++ {
		_ = x[i%len(x)].Sub(y[i%len(y)])
	}
}

func TestAddOverflow(t *testing.T) {
	require := require.New(t)

	max := MustParseDecimal64("-9.999999999999999e384") // largest dec64 number
	small := MustParseDecimal64("0.000000000000001e384")
	require.Equal(NegInfinity64, max.Sub(small))

	max = MustParseDecimal64("9.999999999999999e384") // largest dec64 number
	small = MustParseDecimal64("0.000000000000001e384")
	require.Equal(Infinity64, max.Add(small))

}

func TestQuoOverflow(t *testing.T) {
	require := require.New(t)

	max := MustParseDecimal64("1e384") // largest dec64 number
	small := MustParseDecimal64(".01")
	require.Equal(Infinity64, max.Quo(small))

	max = MustParseDecimal64("1e384") // largest dec64 number
	small = MustParseDecimal64("-.01")
	require.Equal(NegInfinity64, max.Quo(small))

	max = MustParseDecimal64("-1e384") // largest dec64 number
	small = MustParseDecimal64(".01")
	require.Equal(NegInfinity64, max.Quo(small))

	max = MustParseDecimal64("-1e384") // largest dec64 number
	small = MustParseDecimal64("0")
	require.Equal(NegInfinity64, max.Quo(small))

	max = MustParseDecimal64("0") // largest dec64 number
	small = MustParseDecimal64("0")
	require.Equal(QNaN64, max.Quo(small))

	max = MustParseDecimal64("0") // largest dec64 number
	small = MustParseDecimal64("100")
	require.Equal(Zero64, max.Quo(small))

}

func TestMul(t *testing.T) {
	require := require.New(t)

	a := MustParseDecimal64("1e384") // largest dec64 number
	b := MustParseDecimal64("10")
	require.Equal(Infinity64, a.Mul(b))

	a = MustParseDecimal64("1e384") // largest dec64 number
	b = MustParseDecimal64("-10")
	require.Equal(NegInfinity64, a.Mul(b))

	a = MustParseDecimal64("-1e384") // largest dec64 number
	b = MustParseDecimal64("10")
	require.Equal(NegInfinity64, a.Mul(b))

	a = MustParseDecimal64("-1e384") // largest dec64 number
	b = MustParseDecimal64("0")
	require.Equal(NegZero64, a.Mul(b))

	a = MustParseDecimal64("0") // largest dec64 number
	b = MustParseDecimal64("0")
	require.Equal(Zero64, a.Mul(b))

	a = MustParseDecimal64("0") // largest dec64 number
	b = MustParseDecimal64("100")
	require.Equal(Zero64, a.Mul(b))

}
