package decimal

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkDecimal64BinOp(
	t *testing.T,
	expected func(a, b int64) int64,
	actual func(a, b Decimal64) Decimal64,
	op string,
) {
	r := require.New(t)
	for i := int64(-100); i <= 100; i++ {
		a := New64FromInt64(i)
		for j := int64(-100); j <= 100; j++ {
			b := New64FromInt64(j)
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

	fortyTwo := New64FromInt64(42)
	require.Equal(fortyTwo, fortyTwo.Abs())
	require.Equal(fortyTwo, New64FromInt64(-42).Abs())
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

	fortyTwo := New64FromInt64(42)

	require.Equal(QNaN64, fortyTwo.Add(QNaN64))
	require.Equal(QNaN64, QNaN64.Add(fortyTwo))
}

func TestDecimal64AddInf(t *testing.T) {
	require := require.New(t)

	fortyTwo := New64FromInt64(42)

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
	decThree := New64FromInt64(3)
	decTen := New64FromInt64(10)
	decOne := New64FromInt64(1)
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

	fortyTwo := New64FromInt64(42)

	require.Equal(QNaN64, fortyTwo.Mul(QNaN64))
	require.Equal(QNaN64, QNaN64.Mul(fortyTwo))
}

func TestDecimal64MulInf(t *testing.T) {
	require := require.New(t)

	fortyTwo := New64FromInt64(42)
	negFortyTwo := New64FromInt64(-42)

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
			q := n.Quo(d)
			if q != e {
				eFlavor, eSign, eExp, eSignificand := q.parts()
				qFlavor, qSign, qExp, qSignificand := q.parts()
				t.Log("e", e.bits, eFlavor, eSign, eExp, eSignificand)
				t.Log("q", q.bits, qFlavor, qSign, qExp, qSignificand)
			}
			if !assert.Equal(t, e, q, "%d / %d ≠ %v (expecting %v)", k, j, q, e) {
				runtime.Breakpoint()
				n.Quo(d)
				t.FailNow()
			}
		}
	}
}

func TestDecimal64Quo(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping TestDecimal64Quo in short mode.")
	}

	checkDecimal64QuoByF(t, 1)
	checkDecimal64QuoByF(t, 7)
	checkDecimal64QuoByF(t, 13)
}

func TestDecimal64Scale(t *testing.T) {
	t.Parallel()

	const limit = 380

	for i := -limit; i <= limit; i += 3 {
		x := Pi64.ScaleBInt(i)
		for j := -limit; j <= limit; j += 5 {
			y := E64.ScaleBInt(j)
			expected := "1.155727349790922"
			exp := i - j
			switch {
			case exp == 0:
			case -383 <= exp && exp <= 384:
				expected += fmt.Sprintf("e%+d", exp)
			default:
				// TODO: subnormals and infinities
				continue
			}
			actual := x.Quo(y).Text('e', -1)
			require.Equal(t, expected, actual, "i = %v, j = %v", i, j)
		}
	}
}

func TestDecimal64QuoNaN(t *testing.T) {
	require := require.New(t)

	fortyTwo := New64FromInt64(42)

	require.Equal(QNaN64, fortyTwo.Quo(QNaN64))
	require.Equal(QNaN64, QNaN64.Quo(fortyTwo))

}

func TestDecimal64QuoInf(t *testing.T) {
	require := require.New(t)

	fortyTwo := New64FromInt64(42)
	negFortyTwo := New64FromInt64(-42)

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
	for i, u := range tenToThe128 {
		for j, v := range tenToThe128 {
			k := i + j
			if !(k < len(tenToThe128)) {
				continue
			}
			w := tenToThe128[k]
			if !(w.hi == 0 && w.lo < decimal64Base) {
				continue
			}
			e := New64FromInt64(int64(w.lo))
			a := New64FromInt64(int64(u.lo)).Mul(New64FromInt64(int64(v.lo)))
			r.EqualValues(e, a, "%v * %v ≠ %v (expecting %v)", u, v, a, e)
		}
	}
}

func TestDecimal64Sqrt(t *testing.T) {
	r := require.New(t)
	for i := int64(0); i < 100000000; i = i*19/17 + 1 {
		i2 := i * i
		e := New64FromInt64(i)
		n := New64FromInt64(i2)
		a := n.Sqrt()
		r.EqualValues(e, a, "√%v != %v (expected %v)", n, a, e)
	}
}

func TestDecimal64SqrtNeg(t *testing.T) {
	require.EqualValues(t, QNaN64, New64FromInt64(-1).Sqrt())
}

func TestDecimal64SqrtNaN(t *testing.T) {
	require := require.New(t)
	require.Equal(QNaN64, QNaN64.Sqrt())
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

func rnd(ctx Context64, x, y uint64) uint64 {
	ans, _ := ctx.round(x, y)
	return ans
}

func TestRoundHalfUp(t *testing.T) {
	t.Parallel()

	ctx := Context64{Rounding: HalfUp}
	assert.Equal(t, uint64(10), rnd(ctx, 10, 1))
	assert.Equal(t, uint64(10), rnd(ctx, 11, 1))
	assert.Equal(t, uint64(20), rnd(ctx, 15, 1))
	assert.Equal(t, uint64(20), rnd(ctx, 19, 1))
	assert.Equal(t, uint64(200), rnd(ctx, 249, 10))
	assert.Equal(t, uint64(300), rnd(ctx, 250, 10))
	assert.Equal(t, uint64(300), rnd(ctx, 251, 10))
	assert.Equal(t, uint64(300), rnd(ctx, 299, 10))
	assert.Equal(t, uint64(300), rnd(ctx, 300, 10))
	assert.Equal(t, uint64(1000000000000000), rnd(ctx, 1000000000000000, 100000000000000))
	assert.Equal(t, uint64(1000000000000000), rnd(ctx, 1100000000000000, 100000000000000))
	assert.Equal(t, uint64(1000000000000000), rnd(ctx, 1499999999999999, 100000000000000))
	assert.Equal(t, uint64(2000000000000000), rnd(ctx, 1500000000000000, 100000000000000))
	assert.Equal(t, uint64(2000000000000000), rnd(ctx, 1500000000000001, 100000000000000))
	assert.Equal(t, uint64(2000000000000000), rnd(ctx, 1900000000000000, 100000000000000))
	assert.Equal(t, uint64(2000000000000000), rnd(ctx, 1999999999999999, 100000000000000))
	assert.Equal(t, uint64(2000000000000000), rnd(ctx, 2000000000000000, 100000000000000))
	assert.Equal(t, uint64(2000000000000000), rnd(ctx, 2499999999999999, 100000000000000))
	assert.Equal(t, uint64(3000000000000000), rnd(ctx, 2500000000000000, 100000000000000))
	assert.Equal(t, uint64(3000000000000000), rnd(ctx, 2500000000000001, 100000000000000))
	assert.Equal(t, uint64(3000000000000000), rnd(ctx, 2999999999999999, 100000000000000))
	assert.Equal(t, uint64(3000000000000000), rnd(ctx, 3000000000000000, 100000000000000))
}

func TestRoundHalfEven(t *testing.T) {
	t.Parallel()

	ctx := Context64{Rounding: HalfEven}
	assert.Equal(t, uint64(10), rnd(ctx, 10, 1))
	assert.Equal(t, uint64(10), rnd(ctx, 11, 1))
	assert.Equal(t, uint64(20), rnd(ctx, 15, 1))
	assert.Equal(t, uint64(20), rnd(ctx, 19, 1))
	assert.Equal(t, uint64(200), rnd(ctx, 249, 10))
	assert.Equal(t, uint64(200), rnd(ctx, 250, 10))
	assert.Equal(t, uint64(300), rnd(ctx, 251, 10))
	assert.Equal(t, uint64(300), rnd(ctx, 299, 10))
	assert.Equal(t, uint64(300), rnd(ctx, 300, 10))
	assert.Equal(t, uint64(1000000000000000), rnd(ctx, 1000000000000000, 100000000000000))
	assert.Equal(t, uint64(1000000000000000), rnd(ctx, 1100000000000000, 100000000000000))
	assert.Equal(t, uint64(1000000000000000), rnd(ctx, 1499999999999999, 100000000000000))
	assert.Equal(t, uint64(2000000000000000), rnd(ctx, 1500000000000000, 100000000000000))
	assert.Equal(t, uint64(2000000000000000), rnd(ctx, 1500000000000001, 100000000000000))
	assert.Equal(t, uint64(2000000000000000), rnd(ctx, 1900000000000000, 100000000000000))
	assert.Equal(t, uint64(2000000000000000), rnd(ctx, 1999999999999999, 100000000000000))
	assert.Equal(t, uint64(2000000000000000), rnd(ctx, 2000000000000000, 100000000000000))
	assert.Equal(t, uint64(2000000000000000), rnd(ctx, 2499999999999999, 100000000000000))
	assert.Equal(t, uint64(2000000000000000), rnd(ctx, 2500000000000000, 100000000000000))
	assert.Equal(t, uint64(3000000000000000), rnd(ctx, 2500000000000001, 100000000000000))
	assert.Equal(t, uint64(3000000000000000), rnd(ctx, 2999999999999999, 100000000000000))
	assert.Equal(t, uint64(3000000000000000), rnd(ctx, 3000000000000000, 100000000000000))
}

func TestRoundHDown(t *testing.T) {
	t.Parallel()

	ctx := Context64{Rounding: Down}
	assert.Equal(t, uint64(10), rnd(ctx, 10, 1))
	assert.Equal(t, uint64(10), rnd(ctx, 11, 1))
	assert.Equal(t, uint64(10), rnd(ctx, 15, 1))
	assert.Equal(t, uint64(10), rnd(ctx, 19, 1))
	assert.Equal(t, uint64(200), rnd(ctx, 249, 10))
	assert.Equal(t, uint64(200), rnd(ctx, 250, 10))
	assert.Equal(t, uint64(200), rnd(ctx, 251, 10))
	assert.Equal(t, uint64(200), rnd(ctx, 299, 10))
	assert.Equal(t, uint64(300), rnd(ctx, 300, 10))
	assert.Equal(t, uint64(1000000000000000), rnd(ctx, 1000000000000000, 100000000000000))
	assert.Equal(t, uint64(1000000000000000), rnd(ctx, 1100000000000000, 100000000000000))
	assert.Equal(t, uint64(1000000000000000), rnd(ctx, 1499999999999999, 100000000000000))
	assert.Equal(t, uint64(1000000000000000), rnd(ctx, 1500000000000000, 100000000000000))
	assert.Equal(t, uint64(1000000000000000), rnd(ctx, 1500000000000001, 100000000000000))
	assert.Equal(t, uint64(1000000000000000), rnd(ctx, 1900000000000000, 100000000000000))
	assert.Equal(t, uint64(1000000000000000), rnd(ctx, 1999999999999999, 100000000000000))
	assert.Equal(t, uint64(2000000000000000), rnd(ctx, 2000000000000000, 100000000000000))
	assert.Equal(t, uint64(2000000000000000), rnd(ctx, 2499999999999999, 100000000000000))
	assert.Equal(t, uint64(2000000000000000), rnd(ctx, 2500000000000000, 100000000000000))
	assert.Equal(t, uint64(2000000000000000), rnd(ctx, 2500000000000001, 100000000000000))
	assert.Equal(t, uint64(2000000000000000), rnd(ctx, 2999999999999999, 100000000000000))
	assert.Equal(t, uint64(3000000000000000), rnd(ctx, 3000000000000000, 100000000000000))
}

func TestToIntegral(t *testing.T) {
	t.Parallel()

	ctx := Context64{Rounding: HalfUp}
	assert.Equal(t, "0", ctx.ToIntegral(MustParse64("0")).String())
	assert.Equal(t, "0", ctx.ToIntegral(MustParse64("0.499999999999999")).String())
	assert.Equal(t, "1", ctx.ToIntegral(MustParse64("1")).String())
	assert.Equal(t, "1", ctx.ToIntegral(MustParse64("1.49999999999999")).String())
	assert.Equal(t, "2", ctx.ToIntegral(MustParse64("1.5")).String())
	assert.Equal(t, "9", ctx.ToIntegral(MustParse64("9.49999999999999")).String())
	assert.Equal(t, "10", ctx.ToIntegral(MustParse64("9.5")).String())
	assert.Equal(t, "99", ctx.ToIntegral(MustParse64("99.499999999999")).String())
	assert.Equal(t, "100", ctx.ToIntegral(MustParse64("99.5")).String())
}

func benchmarkDecimal64Data() []Decimal64 {
	return []Decimal64{
		One64,
		QNaN64,
		Infinity64,
		NegInfinity64,
		Pi64,
		E64,
		New64FromInt64(42),
		MustParse64("2345678e100"),
		New64FromInt64(1234567),
		New64FromInt64(-42),
		MustParse64("3456789e-120"),
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
	require.Equal(NegInfinity64, NegMax64.Sub(MustParse64("0.00000000000001e384")))
	require.Equal(Infinity64, Max64.Add(MustParse64("0.000000000000001e384")))
	require.Equal(Max64, Max64.Add(MustParse64("1")))
	require.Equal(Max64, Zero64.Add(Max64))
}

func TestQuoOverflow(t *testing.T) {
	test := func(expected Decimal64, num, denom string) {
		n := MustParse64(num)
		d := MustParse64(denom)
		if !assert.Equal(t, expected, n.Quo(d)) {
			runtime.Breakpoint()
			n.Quo(d)
		}
	}
	test(Infinity64, "1e384", ".01")
	test(NegInfinity64, "1e384", "-.01")
	test(NegInfinity64, "-1e384", ".01")
	test(NegInfinity64, "-1e384", "0")
	test(QNaN64, "0", "0")
	test(Zero64, "0", "100")
}

func TestMul(t *testing.T) {
	require := require.New(t)
	require.Equal(Infinity64, MustParse64("1e384").Mul(MustParse64("10")))
	require.Equal(NegInfinity64, MustParse64("1e384").Mul(MustParse64("-10")))
	require.Equal(NegInfinity64, MustParse64("-1e384").Mul(MustParse64("10")))
	require.Equal(NegZero64, MustParse64("-1e384").Mul(Zero64))
	require.Equal(Zero64, Zero64.Mul(Zero64))
	require.Equal(Zero64, Zero64.Mul(MustParse64("100")))
}
