package decimal

import (
	"cmp"
	"fmt"
	"log"
	"testing"
)

var sink any

func checkDecimal64BinOp(
	t *testing.T,
	expected func(a, b int64) int64,
	actual func(a, b Decimal64) Decimal64,
) {
	for i := int64(-100); i <= 100; i++ {
		a := New64FromInt64(i)
		for j := int64(-100); j <= 100; j++ {
			b := New64FromInt64(j)
			c := actual(a, b)
			k := c.Int64()
			e := expected(i, j)
			equal(t, e, k)
		}
	}
}

func TestDecimal64Abs(t *testing.T) {
	t.Parallel()

	equal(t, Zero64, Zero64.Abs())
	equal(t, Zero64, NegZero64.Abs())
	equal(t, Infinity64, Infinity64.Abs())
	equal(t, Infinity64, NegInfinity64.Abs())

	fortyTwo := New64FromInt64(42)
	equal(t, fortyTwo, fortyTwo.Abs())
	equal(t, fortyTwo, New64FromInt64(-42).Abs())
}

func TestDecimal64Add(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping TestDecimal64Add in short mode.")
	}
	checkDecimal64BinOp(t,
		func(a, b int64) int64 { return a + b },
		func(a, b Decimal64) Decimal64 { return a.Add(b) },
	)

	add := func(a, b, expected string, ctx *Context64) func(*testing.T) {
		return func(*testing.T) {
			t.Helper()

			e := MustParse64(expected)
			x := MustParse64(a)
			y := MustParse64(b)
			replayOnFail(t, func() {
				z := cmp.Or(ctx, &DefaultContext64).Add(x, y)
				equalD64(t, e, z)
			})
		}
	}

	t.Run("tiny-neg", add("1E-383", "-1E-398", "9.99999999999999E-384", nil))

	he := Context64{Rounding: HalfEven}
	t.Run("round-even", add("12345678", "0.123456785", "12345678.12345678", &he))
}

func TestDecimal64AddNaN(t *testing.T) {
	t.Parallel()

	fortyTwo := New64FromInt64(42)

	equal(t, QNaN64, fortyTwo.Add(QNaN64))
	equal(t, QNaN64, QNaN64.Add(fortyTwo))
}

func TestDecimal64AddInf(t *testing.T) {
	t.Parallel()

	fortyTwo := New64FromInt64(42)

	equal(t, Infinity64, fortyTwo.Add(Infinity64))
	equal(t, Infinity64, Infinity64.Add(fortyTwo))

	equal(t, NegInfinity64, fortyTwo.Add(NegInfinity64))
	equal(t, NegInfinity64, NegInfinity64.Add(fortyTwo))

	equal(t, Infinity64, Infinity64.Add(Infinity64))
	equal(t, NegInfinity64, NegInfinity64.Add(NegInfinity64))

	equal(t, QNaN64, Infinity64.Add(NegInfinity64))
	equal(t, QNaN64, NegInfinity64.Add(Infinity64))
}

func TestDecimal64Cmp(t *testing.T) {
	t.Parallel()

	equal(t, 0, NegOne64.Cmp(NegOne64))

	equal(t, 0, Zero64.Cmp(Zero64))
	equal(t, 0, Zero64.Cmp(NegZero64))
	equal(t, 0, NegZero64.Cmp(Zero64))
	equal(t, 0, NegZero64.Cmp(NegZero64))

	equal(t, 0, One64.Cmp(One64))
	equal(t, -1, NegOne64.Cmp(Zero64))
	equal(t, -1, NegOne64.Cmp(NegZero64))
	equal(t, -1, NegOne64.Cmp(One64))
	equal(t, -1, Zero64.Cmp(One64))
	equal(t, -1, NegZero64.Cmp(One64))
	equal(t, 1, Zero64.Cmp(NegOne64))
	equal(t, 1, NegZero64.Cmp(NegOne64))
	equal(t, 1, One64.Cmp(NegOne64))
	equal(t, 1, One64.Cmp(Zero64))
	equal(t, 1, One64.Cmp(NegZero64))
}

func TestDecimal64CmpNaN(t *testing.T) {
	t.Parallel()

	equal(t, -2, QNaN64.Cmp(QNaN64))
	equal(t, -2, Zero64.Cmp(QNaN64))
	equal(t, -2, QNaN64.Cmp(Zero64))
}

func TestDecimal64MulThreeByOneTenthByTen(t *testing.T) {
	t.Parallel()

	// float 3*0.1*10 ≠ 3
	fltThree := 3.0
	fltTen := 10.0
	fltOne := 1.0
	fltOneTenth := fltOne / fltTen
	fltProduct := fltThree * fltOneTenth * fltTen
	equal(t, fltTen*fltOneTenth, fltOne)
	notequal(t, fltThree, fltProduct)

	// decimal 3*0.1*10 = 3
	decThree := New64FromInt64(3)
	decTen := New64FromInt64(10)
	decOne := New64FromInt64(1)
	decOneTenth := decOne.Quo(decTen)
	decProduct := decThree.Mul(decOneTenth).Mul(decTen)
	equalD64(t, decTen.Mul(decOneTenth), decOne)
	equalD64(t, decThree, decProduct)
}

func TestDecimal64Mul(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping TestDecimal64Mul in short mode.")
	}
	checkDecimal64BinOp(t,
		func(a, b int64) int64 { return a * b },
		func(a, b Decimal64) Decimal64 { return a.Mul(b) },
	)
}

func TestDecimal64MulNaN(t *testing.T) {
	t.Parallel()

	fortyTwo := New64FromInt64(42)

	equal(t, QNaN64, fortyTwo.Mul(QNaN64))
	equal(t, QNaN64, QNaN64.Mul(fortyTwo))
}

func TestDecimal64MulInf(t *testing.T) {
	t.Parallel()

	fortyTwo := New64FromInt64(42)
	negFortyTwo := New64FromInt64(-42)

	equal(t, Infinity64, fortyTwo.Mul(Infinity64))
	equal(t, Infinity64, Infinity64.Mul(fortyTwo))
	equal(t, NegInfinity64, negFortyTwo.Mul(Infinity64))
	equal(t, NegInfinity64, Infinity64.Mul(negFortyTwo))

	equal(t, NegInfinity64, fortyTwo.Mul(NegInfinity64))
	equal(t, NegInfinity64, NegInfinity64.Mul(fortyTwo))
	equal(t, Infinity64, negFortyTwo.Mul(NegInfinity64))
	equal(t, Infinity64, NegInfinity64.Mul(negFortyTwo))

	equal(t, Infinity64, Infinity64.Mul(Infinity64))
	equal(t, Infinity64, NegInfinity64.Mul(NegInfinity64))
	equal(t, NegInfinity64, Infinity64.Mul(NegInfinity64))
	equal(t, NegInfinity64, NegInfinity64.Mul(Infinity64))
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
			if !equal(t, e, q) {
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
			equal(t, expected, actual)
		}
	}
}

func TestDecimal64QuoNaN(t *testing.T) {
	t.Parallel()

	fortyTwo := New64FromInt64(42)

	equal(t, QNaN64, fortyTwo.Quo(QNaN64))
	equal(t, QNaN64, QNaN64.Quo(fortyTwo))

}

func TestDecimal64QuoInf(t *testing.T) {
	t.Parallel()

	fortyTwo := New64FromInt64(42)
	negFortyTwo := New64FromInt64(-42)

	equal(t, Zero64, fortyTwo.Quo(Infinity64))
	equal(t, Infinity64, Infinity64.Quo(fortyTwo))
	equal(t, NegZero64, negFortyTwo.Quo(Infinity64))
	equal(t, NegInfinity64, Infinity64.Quo(negFortyTwo))

	equal(t, NegZero64, fortyTwo.Quo(NegInfinity64))
	equal(t, NegInfinity64, NegInfinity64.Quo(fortyTwo))
	equal(t, Zero64, negFortyTwo.Quo(NegInfinity64))
	equal(t, Infinity64, NegInfinity64.Quo(negFortyTwo))

	equal(t, QNaN64, Infinity64.Quo(Infinity64))
	equal(t, QNaN64, NegInfinity64.Quo(NegInfinity64))
	equal(t, QNaN64, Infinity64.Quo(NegInfinity64))
	equal(t, QNaN64, NegInfinity64.Quo(Infinity64))
}

func TestDecimal64MulPo10(t *testing.T) {
	t.Parallel()

	for i, u := range tenToThe128[:39] {
		for j, v := range tenToThe128[:39] {
			k := i + j
			if !(k < 39) {
				continue
			}
			w := tenToThe128[k]
			if !(w.hi == 0 && w.lo < decimal64Base) {
				continue
			}
			e := New64FromInt64(int64(w.lo))
			a := New64FromInt64(int64(u.lo)).Mul(New64FromInt64(int64(v.lo)))
			equalD64(t, e, a)
		}
	}
}

func TestDecimal64Sqrt(t *testing.T) {
	t.Parallel()

	for i := int64(0); i < 100000000; i = i*19/17 + 1 {
		i2 := i * i
		e := New64FromInt64(i)
		n := New64FromInt64(i2)
		replayOnFail(t, func() {
			a := n.Sqrt()
			equalD64(t, e, a).Or(t.FailNow)
		})
	}
}

func TestDecimal64SqrtNeg(t *testing.T) {
	t.Parallel()

	equal(t, QNaN64, New64FromInt64(-1).Sqrt())
}

func TestDecimal64SqrtNaN(t *testing.T) {
	t.Parallel()

	equal(t, QNaN64, QNaN64.Sqrt())
}

func TestDecimal64SqrtInf(t *testing.T) {
	t.Parallel()

	equal(t, Infinity64, Infinity64.Sqrt())
	equal(t, QNaN64, NegInfinity64.Sqrt())
}

func TestDecimal64Sub(t *testing.T) {
	t.Parallel()

	checkDecimal64BinOp(t,
		func(a, b int64) int64 { return a - b },
		func(a, b Decimal64) Decimal64 { return a.Sub(b) },
	)
}

func rnd(ctx Context64, x, y uint64) uint64 {
	ans, _ := ctx.round(x, y)
	return ans
}

func TestRoundHalfUp(t *testing.T) {
	t.Parallel()

	ctx := Context64{Rounding: HalfUp}
	equal(t, uint64(10), rnd(ctx, 10, 1))
	equal(t, uint64(10), rnd(ctx, 11, 1))
	equal(t, uint64(20), rnd(ctx, 15, 1))
	equal(t, uint64(20), rnd(ctx, 19, 1))
	equal(t, uint64(200), rnd(ctx, 249, 10))
	equal(t, uint64(300), rnd(ctx, 250, 10))
	equal(t, uint64(300), rnd(ctx, 251, 10))
	equal(t, uint64(300), rnd(ctx, 299, 10))
	equal(t, uint64(300), rnd(ctx, 300, 10))
	equal(t, uint64(1000000000000000), rnd(ctx, 1000000000000000, 100000000000000))
	equal(t, uint64(1000000000000000), rnd(ctx, 1100000000000000, 100000000000000))
	equal(t, uint64(1000000000000000), rnd(ctx, 1499999999999999, 100000000000000))
	equal(t, uint64(2000000000000000), rnd(ctx, 1500000000000000, 100000000000000))
	equal(t, uint64(2000000000000000), rnd(ctx, 1500000000000001, 100000000000000))
	equal(t, uint64(2000000000000000), rnd(ctx, 1900000000000000, 100000000000000))
	equal(t, uint64(2000000000000000), rnd(ctx, 1999999999999999, 100000000000000))
	equal(t, uint64(2000000000000000), rnd(ctx, 2000000000000000, 100000000000000))
	equal(t, uint64(2000000000000000), rnd(ctx, 2499999999999999, 100000000000000))
	equal(t, uint64(3000000000000000), rnd(ctx, 2500000000000000, 100000000000000))
	equal(t, uint64(3000000000000000), rnd(ctx, 2500000000000001, 100000000000000))
	equal(t, uint64(3000000000000000), rnd(ctx, 2999999999999999, 100000000000000))
	equal(t, uint64(3000000000000000), rnd(ctx, 3000000000000000, 100000000000000))
}

func TestRoundHalfEven(t *testing.T) {
	t.Parallel()

	ctx := Context64{Rounding: HalfEven}
	equal(t, uint64(10), rnd(ctx, 10, 1))
	equal(t, uint64(10), rnd(ctx, 11, 1))
	equal(t, uint64(20), rnd(ctx, 15, 1))
	equal(t, uint64(20), rnd(ctx, 19, 1))
	equal(t, uint64(200), rnd(ctx, 249, 10))
	equal(t, uint64(200), rnd(ctx, 250, 10))
	equal(t, uint64(300), rnd(ctx, 251, 10))
	equal(t, uint64(300), rnd(ctx, 299, 10))
	equal(t, uint64(300), rnd(ctx, 300, 10))
	equal(t, uint64(1000000000000000), rnd(ctx, 1000000000000000, 100000000000000))
	equal(t, uint64(1000000000000000), rnd(ctx, 1100000000000000, 100000000000000))
	equal(t, uint64(1000000000000000), rnd(ctx, 1499999999999999, 100000000000000))
	equal(t, uint64(2000000000000000), rnd(ctx, 1500000000000000, 100000000000000))
	equal(t, uint64(2000000000000000), rnd(ctx, 1500000000000001, 100000000000000))
	equal(t, uint64(2000000000000000), rnd(ctx, 1900000000000000, 100000000000000))
	equal(t, uint64(2000000000000000), rnd(ctx, 1999999999999999, 100000000000000))
	equal(t, uint64(2000000000000000), rnd(ctx, 2000000000000000, 100000000000000))
	equal(t, uint64(2000000000000000), rnd(ctx, 2499999999999999, 100000000000000))
	equal(t, uint64(2000000000000000), rnd(ctx, 2500000000000000, 100000000000000))
	equal(t, uint64(3000000000000000), rnd(ctx, 2500000000000001, 100000000000000))
	equal(t, uint64(3000000000000000), rnd(ctx, 2999999999999999, 100000000000000))
	equal(t, uint64(3000000000000000), rnd(ctx, 3000000000000000, 100000000000000))
}

func TestRoundHDown(t *testing.T) {
	t.Parallel()

	ctx := Context64{Rounding: Down}
	equal(t, uint64(10), rnd(ctx, 10, 1))
	equal(t, uint64(10), rnd(ctx, 11, 1))
	equal(t, uint64(10), rnd(ctx, 15, 1))
	equal(t, uint64(10), rnd(ctx, 19, 1))
	equal(t, uint64(200), rnd(ctx, 249, 10))
	equal(t, uint64(200), rnd(ctx, 250, 10))
	equal(t, uint64(200), rnd(ctx, 251, 10))
	equal(t, uint64(200), rnd(ctx, 299, 10))
	equal(t, uint64(300), rnd(ctx, 300, 10))
	equal(t, uint64(1000000000000000), rnd(ctx, 1000000000000000, 100000000000000))
	equal(t, uint64(1000000000000000), rnd(ctx, 1100000000000000, 100000000000000))
	equal(t, uint64(1000000000000000), rnd(ctx, 1499999999999999, 100000000000000))
	equal(t, uint64(1000000000000000), rnd(ctx, 1500000000000000, 100000000000000))
	equal(t, uint64(1000000000000000), rnd(ctx, 1500000000000001, 100000000000000))
	equal(t, uint64(1000000000000000), rnd(ctx, 1900000000000000, 100000000000000))
	equal(t, uint64(1000000000000000), rnd(ctx, 1999999999999999, 100000000000000))
	equal(t, uint64(2000000000000000), rnd(ctx, 2000000000000000, 100000000000000))
	equal(t, uint64(2000000000000000), rnd(ctx, 2499999999999999, 100000000000000))
	equal(t, uint64(2000000000000000), rnd(ctx, 2500000000000000, 100000000000000))
	equal(t, uint64(2000000000000000), rnd(ctx, 2500000000000001, 100000000000000))
	equal(t, uint64(2000000000000000), rnd(ctx, 2999999999999999, 100000000000000))
	equal(t, uint64(3000000000000000), rnd(ctx, 3000000000000000, 100000000000000))
}

func TestToIntegral(t *testing.T) {
	t.Parallel()

	ctx := Context64{Rounding: HalfUp}
	equal(t, "0", ctx.ToIntegral(MustParse64("0")).String())
	equal(t, "0", ctx.ToIntegral(MustParse64("0.499999999999999")).String())
	equal(t, "1", ctx.ToIntegral(MustParse64("1")).String())
	equal(t, "1", ctx.ToIntegral(MustParse64("1.49999999999999")).String())
	equal(t, "2", ctx.ToIntegral(MustParse64("1.5")).String())
	equal(t, "9", ctx.ToIntegral(MustParse64("9.49999999999999")).String())
	equal(t, "10", ctx.ToIntegral(MustParse64("9.5")).String())
	equal(t, "99", ctx.ToIntegral(MustParse64("99.499999999999")).String())
	equal(t, "100", ctx.ToIntegral(MustParse64("99.5")).String())
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
		MustParse64("9945678e100"),
		New64FromInt64(1234567),
		New64FromInt64(-42),
		MustParse64("3456789e-120"),
	}
}

func BenchmarkDecimal64Abs(b *testing.B) {
	x := benchmarkDecimal64Data()
	for i := 0; i < b.N; i++ {
		_ = x[i%len(x)].Abs()
	}
}

func BenchmarkDecimal64Add(b *testing.B) {
	x := benchmarkDecimal64Data()
	y := x[:len(x)-2]
	for i := 0; i < b.N; i++ {
		_ = x[i%len(x)].Add(y[i%len(y)])
	}
}

func BenchmarkDecimal64Cmp(b *testing.B) {
	x := benchmarkDecimal64Data()
	y := x[:len(x)-2]
	for i := 0; i < b.N; i++ {
		_ = x[i%len(x)].Cmp(y[i%len(y)])
	}
}

func BenchmarkDecimal64Mul(b *testing.B) {
	x := One64
	y, err := Parse64("3.142")
	if err != nil {
		b.Fatal(err)
	}
	z := x.Quo(y)
	for i := 0; i < b.N; i++ {
		x = x.Mul(z)
		y, z = z, y
	}
	sink = x
}

func BenchmarkFloat64Mul(b *testing.B) {
	x := 1.0
	y := 3.142
	z := 1 / y
	for i := 0; i < b.N; i++ {
		x *= z
		y, z = z, y
	}
	sink = x
}

func BenchmarkDecimal64Quo(b *testing.B) {
	x := benchmarkDecimal64Data()
	for i := 0; i < b.N; i++ {
		_ = x[i%len(x)].Mul(x[(2*i)%len(x)])
	}
}

func BenchmarkDecimal64Sqrt(b *testing.B) {
	x := benchmarkDecimal64Data()
	for i := 0; i < b.N; i++ {
		_ = x[i%len(x)].Sqrt()
	}
}

func BenchmarkDecimal64Sub(b *testing.B) {
	x := benchmarkDecimal64Data()
	y := x[:len(x)-2]
	for i := 0; i < b.N; i++ {
		_ = x[i%len(x)].Sub(y[i%len(y)])
	}
}

func TestAddOverflow(t *testing.T) {
	t.Parallel()

	equal(t, NegInfinity64, NegMax64.Sub(MustParse64("0.00000000000001e384")))
	equal(t, Infinity64, Max64.Add(MustParse64("0.000000000000001e384")))
	equal(t, Max64, Max64.Add(MustParse64("1")))
	equal(t, Max64, Zero64.Add(Max64))
}

func TestQuoOverflow(t *testing.T) {
	t.Parallel()

	test := func(expected Decimal64, num, denom string) {
		n := MustParse64(num)
		d := MustParse64(denom)
		if !equal(t, expected, n.Quo(d)) {
			log.Printf("TestQuoOverflow: num = %d", n)
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
	t.Parallel()

	equal(t, Infinity64, MustParse64("1e384").Mul(MustParse64("10")))
	equal(t, NegInfinity64, MustParse64("1e384").Mul(MustParse64("-10")))
	equal(t, NegInfinity64, MustParse64("-1e384").Mul(MustParse64("10")))
	equal(t, NegZero64, MustParse64("-1e384").Mul(Zero64))
	equal(t, Zero64, Zero64.Mul(Zero64))
	equal(t, Zero64, Zero64.Mul(MustParse64("100")))
}
