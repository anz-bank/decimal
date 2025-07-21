package d64

import (
	"fmt"
	"log"
	"testing"
)

var sink any

func checkDecimalBinOp(
	t *testing.T,
	expected func(a, b int64) int64,
	actual func(a, b Decimal) Decimal,
) {
	t.Helper()

	for i := int64(-100); i <= 100; i++ {
		a := NewFromInt64(i)
		for j := int64(-100); j <= 100; j++ {
			b := NewFromInt64(j)
			c := actual(a, b)
			k := c.Int64()
			e := expected(i, j)
			equal(t, e, k)
		}
	}
}

func TestDecimalAbs(t *testing.T) {
	t.Parallel()

	equal(t, Zero, Zero.Abs())
	equal(t, Zero, NegZero.Abs())
	equal(t, Inf, Inf.Abs())
	equal(t, Inf, NegInf.Abs())

	fortyTwo := NewFromInt64(42)
	equal(t, fortyTwo, fortyTwo.Abs())
	equal(t, fortyTwo, NewFromInt64(-42).Abs())
}

func TestDecimalAdd(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping TestDecimalAdd in short mode.")
	}
	checkDecimalBinOp(t,
		func(a, b int64) int64 { return a + b },
		func(a, b Decimal) Decimal { return a.Add(b) },
	)

	add := func(a, b, expected string, ctx *Context) func(*testing.T) {
		return func(*testing.T) {
			t.Helper()

			e := MustParse(expected)
			x := MustParse(a)
			y := MustParse(b)
			if ctx == nil {
				ctx = &DefaultContext
			}
			replayOnFail(t, func() {
				z := ctx.Add(x, y)
				equalD64(t, e, z)
			})
		}
	}

	t.Run("tiny-neg", add("1E-383", "-1E-398", "9.99999999999999E-384", nil))

	he := Context{Rounding: HalfEven}
	t.Run("round-even", add("12345678", "0.123456785", "12345678.12345678", &he))
}

func TestDecimalAddNaN(t *testing.T) {
	t.Parallel()

	fortyTwo := NewFromInt64(42)

	equal(t, QNaN, fortyTwo.Add(QNaN))
	equal(t, QNaN, QNaN.Add(fortyTwo))
}

func TestDecimalAddInf(t *testing.T) {
	t.Parallel()

	fortyTwo := NewFromInt64(42)

	equal(t, Inf, fortyTwo.Add(Inf))
	equal(t, Inf, Inf.Add(fortyTwo))

	equal(t, NegInf, fortyTwo.Add(NegInf))
	equal(t, NegInf, NegInf.Add(fortyTwo))

	equal(t, Inf, Inf.Add(Inf))
	equal(t, NegInf, NegInf.Add(NegInf))

	equal(t, QNaN, Inf.Add(NegInf))
	equal(t, QNaN, NegInf.Add(Inf))
}

func TestDecimalCmp(t *testing.T) {
	t.Parallel()

	equal(t, 0, NegOne.Cmp(NegOne))

	equal(t, 0, Zero.Cmp(Zero))
	equal(t, 0, Zero.Cmp(NegZero))
	equal(t, 0, NegZero.Cmp(Zero))
	equal(t, 0, NegZero.Cmp(NegZero))

	equal(t, 0, One.Cmp(One))
	equal(t, -1, NegOne.Cmp(Zero))
	equal(t, -1, NegOne.Cmp(NegZero))
	equal(t, -1, NegOne.Cmp(One))
	equal(t, -1, Zero.Cmp(One))
	equal(t, -1, NegZero.Cmp(One))
	equal(t, 1, Zero.Cmp(NegOne))
	equal(t, 1, NegZero.Cmp(NegOne))
	equal(t, 1, One.Cmp(NegOne))
	equal(t, 1, One.Cmp(Zero))
	equal(t, 1, One.Cmp(NegZero))
}

func TestDecimalCmpNaN(t *testing.T) {
	t.Parallel()

	equal(t, -2, QNaN.Cmp(QNaN))
	equal(t, -2, Zero.Cmp(QNaN))
	equal(t, -2, QNaN.Cmp(Zero))
}

func TestDecimalMulThreeByOneTenthByTen(t *testing.T) {
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
	decThree := NewFromInt64(3)
	decTen := NewFromInt64(10)
	decOne := NewFromInt64(1)
	decOneTenth := decOne.Quo(decTen)
	decProduct := decThree.Mul(decOneTenth).Mul(decTen)
	equalD64(t, decTen.Mul(decOneTenth), decOne)
	equalD64(t, decThree, decProduct)
}

func TestDecimalMul(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping TestDecimalMul in short mode.")
	}
	checkDecimalBinOp(t,
		func(a, b int64) int64 { return a * b },
		func(a, b Decimal) Decimal { return a.Mul(b) },
	)
}

func TestDecimalMulNaN(t *testing.T) {
	t.Parallel()

	fortyTwo := NewFromInt64(42)

	equal(t, QNaN, fortyTwo.Mul(QNaN))
	equal(t, QNaN, QNaN.Mul(fortyTwo))
}

func TestDecimalMulInf(t *testing.T) {
	t.Parallel()

	fortyTwo := NewFromInt64(42)
	negFortyTwo := NewFromInt64(-42)

	equal(t, Inf, fortyTwo.Mul(Inf))
	equal(t, Inf, Inf.Mul(fortyTwo))
	equal(t, NegInf, negFortyTwo.Mul(Inf))
	equal(t, NegInf, Inf.Mul(negFortyTwo))

	equal(t, NegInf, fortyTwo.Mul(NegInf))
	equal(t, NegInf, NegInf.Mul(fortyTwo))
	equal(t, Inf, negFortyTwo.Mul(NegInf))
	equal(t, Inf, NegInf.Mul(negFortyTwo))

	equal(t, Inf, Inf.Mul(Inf))
	equal(t, Inf, NegInf.Mul(NegInf))
	equal(t, NegInf, Inf.Mul(NegInf))
	equal(t, NegInf, NegInf.Mul(Inf))
}

func checkDecimalQuoByF(t *testing.T, f int64) {
	for i := int64(-1000 * f); i <= 1000*f; i += f {
		for j := int64(-100); j <= 100; j++ {
			var e Decimal
			if j == 0 {
				e = QNaN
			} else {
				e = NewFromInt64(i)
				if i == 0 && j < 0 {
					e = e.Neg()
				}
			}
			k := i * j
			n := NewFromInt64(k)
			d := NewFromInt64(j)
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

func TestDecimalQuo(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping TestDecimalQuo in short mode.")
	}

	checkDecimalQuoByF(t, 1)
	checkDecimalQuoByF(t, 7)
	checkDecimalQuoByF(t, 13)
}

func TestDecimalRound(t *testing.T) {
	t.Parallel()

	round := func(x, y, e string) func(t *testing.T) {
		return func(t *testing.T) {
			t.Helper()

			expected := MustParse(e)
			actual := DefaultContext.Round(MustParse(x), MustParse(y))
			equalD64(t, expected, actual)
		}
	}

	t.Run("one", round("2", "1", "2"))
	t.Run("zero", round("2", "0", "0"))
	t.Run("ten", round(("-2"), "10", "-0"))
	t.Run("one-10th", round("2", "0.1", "2"))
	t.Run("one-100th", round("2", "0.01", "2"))
	t.Run("one-100th-lg", round("2000.046", "0.01", "2000.05"))
}

func TestDecimalScale(t *testing.T) {
	t.Parallel()

	const limit = 380

	for i := -limit; i <= limit; i += 3 {
		x := Pi.ScaleBInt(i)
		for j := -limit; j <= limit; j += 5 {
			y := E.ScaleBInt(j)
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

func TestDecimalQuoNaN(t *testing.T) {
	t.Parallel()

	fortyTwo := NewFromInt64(42)

	equal(t, QNaN, fortyTwo.Quo(QNaN))
	equal(t, QNaN, QNaN.Quo(fortyTwo))

}

func TestDecimalQuoInf(t *testing.T) {
	t.Parallel()

	fortyTwo := NewFromInt64(42)
	negFortyTwo := NewFromInt64(-42)

	equal(t, Zero, fortyTwo.Quo(Inf))
	equal(t, Inf, Inf.Quo(fortyTwo))
	equal(t, NegZero, negFortyTwo.Quo(Inf))
	equal(t, NegInf, Inf.Quo(negFortyTwo))

	equal(t, NegZero, fortyTwo.Quo(NegInf))
	equal(t, NegInf, NegInf.Quo(fortyTwo))
	equal(t, Zero, negFortyTwo.Quo(NegInf))
	equal(t, Inf, NegInf.Quo(negFortyTwo))

	equal(t, QNaN, Inf.Quo(Inf))
	equal(t, QNaN, NegInf.Quo(NegInf))
	equal(t, QNaN, Inf.Quo(NegInf))
	equal(t, QNaN, NegInf.Quo(Inf))
}

func TestDecimalMulPo10(t *testing.T) {
	t.Parallel()

	for i, u := range tenToThe128[:39] {
		for j, v := range tenToThe128[:39] {
			k := i + j
			if !(k < 39) {
				continue
			}
			w := tenToThe128[k]
			if !(w.hi == 0 && w.lo < decimalBase) {
				continue
			}
			e := NewFromInt64(int64(w.lo))
			a := NewFromInt64(int64(u.lo)).Mul(NewFromInt64(int64(v.lo)))
			equalD64(t, e, a)
		}
	}
}

func TestDecimalSqrt(t *testing.T) {
	t.Parallel()

	for i := int64(0); i < 100000000; i = i*19/17 + 1 {
		i2 := i * i
		e := NewFromInt64(i)
		n := NewFromInt64(i2)
		replayOnFail(t, func() {
			a := n.Sqrt()
			equalD64(t, e, a).Or(t.FailNow)
		})
	}
}

func TestDecimalSqrtNeg(t *testing.T) {
	t.Parallel()

	equal(t, QNaN, NewFromInt64(-1).Sqrt())
}

func TestDecimalSqrtNaN(t *testing.T) {
	t.Parallel()

	equal(t, QNaN, QNaN.Sqrt())
}

func TestDecimalSqrtInf(t *testing.T) {
	t.Parallel()

	equal(t, Inf, Inf.Sqrt())
	equal(t, QNaN, NegInf.Sqrt())
}

func TestDecimalSub(t *testing.T) {
	t.Parallel()

	checkDecimalBinOp(t,
		func(a, b int64) int64 { return a - b },
		func(a, b Decimal) Decimal { return a.Sub(b) },
	)
}

func rnd(ctx Context, x, y uint64) uint64 {
	ans, _ := ctx.round(x, y)
	return ans
}

func TestRoundHalfUp(t *testing.T) {
	t.Parallel()

	ctx := Context{Rounding: HalfUp}
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

	ctx := Context{Rounding: HalfEven}
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

	ctx := Context{Rounding: Down}
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

	ctx := Context{Rounding: HalfUp}
	equal(t, "0", ctx.ToIntegral(MustParse("0")).String())
	equal(t, "0", ctx.ToIntegral(MustParse("0.499999999999999")).String())
	equal(t, "1", ctx.ToIntegral(MustParse("1")).String())
	equal(t, "1", ctx.ToIntegral(MustParse("1.49999999999999")).String())
	equal(t, "2", ctx.ToIntegral(MustParse("1.5")).String())
	equal(t, "9", ctx.ToIntegral(MustParse("9.49999999999999")).String())
	equal(t, "10", ctx.ToIntegral(MustParse("9.5")).String())
	equal(t, "99", ctx.ToIntegral(MustParse("99.499999999999")).String())
	equal(t, "100", ctx.ToIntegral(MustParse("99.5")).String())
}

func benchmarkDecimalData() []Decimal {
	return []Decimal{
		One,
		QNaN,
		Inf,
		NegInf,
		Pi,
		E,
		NewFromInt64(42),
		MustParse("9945678e100"),
		NewFromInt64(1234567),
		NewFromInt64(-42),
		MustParse("3456789e-120"),
	}
}

func BenchmarkDecimalAbs(b *testing.B) {
	x := benchmarkDecimalData()
	for i := 0; i < b.N; i++ {
		_ = x[i%len(x)].Abs()
	}
}

func BenchmarkDecimalAdd(b *testing.B) {
	x := benchmarkDecimalData()
	y := x[:len(x)-2]
	for i := 0; i < b.N; i++ {
		_ = x[i%len(x)].Add(y[i%len(y)])
	}
}

func BenchmarkDecimalCmp(b *testing.B) {
	x := benchmarkDecimalData()
	y := x[:len(x)-2]
	for i := 0; i < b.N; i++ {
		_ = x[i%len(x)].Cmp(y[i%len(y)])
	}
}

func BenchmarkDecimalMul(b *testing.B) {
	x := One
	y, err := Parse("3.142")
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

func BenchmarkDecimalQuo(b *testing.B) {
	x := benchmarkDecimalData()
	for i := 0; i < b.N; i++ {
		_ = x[i%len(x)].Mul(x[(2*i)%len(x)])
	}
}

func BenchmarkDecimalSqrt(b *testing.B) {
	x := benchmarkDecimalData()
	for i := 0; i < b.N; i++ {
		_ = x[i%len(x)].Sqrt()
	}
}

func BenchmarkDecimalSub(b *testing.B) {
	x := benchmarkDecimalData()
	y := x[:len(x)-2]
	for i := 0; i < b.N; i++ {
		_ = x[i%len(x)].Sub(y[i%len(y)])
	}
}

func TestAddOverflow(t *testing.T) {
	t.Parallel()

	equal(t, NegInf, NegMax.Sub(MustParse("0.00000000000001e384")))
	equal(t, Inf, Max.Add(MustParse("0.000000000000001e384")))
	equal(t, Max, Max.Add(MustParse("1")))
	equal(t, Max, Zero.Add(Max))
}

func TestQuoOverflow(t *testing.T) {
	t.Parallel()

	test := func(expected Decimal, num, denom string) {
		n := MustParse(num)
		d := MustParse(denom)
		if !equal(t, expected, n.Quo(d)) {
			log.Printf("TestQuoOverflow: num = %d", n)
			n.Quo(d)
		}
	}
	test(Inf, "1e384", ".01")
	test(NegInf, "1e384", "-.01")
	test(NegInf, "-1e384", ".01")
	test(NegInf, "-1e384", "0")
	test(QNaN, "0", "0")
	test(Zero, "0", "100")
}

func TestMul(t *testing.T) {
	t.Parallel()

	equal(t, Inf, MustParse("1e384").Mul(MustParse("10")))
	equal(t, NegInf, MustParse("1e384").Mul(MustParse("-10")))
	equal(t, NegInf, MustParse("-1e384").Mul(MustParse("10")))
	equal(t, NegZero, MustParse("-1e384").Mul(Zero))
	equal(t, Zero, Zero.Mul(Zero))
	equal(t, Zero, Zero.Mul(MustParse("100")))
}
