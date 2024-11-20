package decimal

import (
	"math/bits"
)

type uint128T struct {
	lo, hi uint64
}

func (a *uint128T) numDecimalDigits() int {
	if a.hi == 0 {
		return numDecimalDigits(a.lo)
	}
	bitSize := 129 - uint(bits.LeadingZeros64(a.hi))
	numDigitsEst := int(bitSize * 3 / 10)
	if a.lt(tenToThe128[numDigitsEst]) {
		return numDigitsEst
	}
	return numDigitsEst + 1
}

var tenToThe128 = func() [39]uint128T {
	var ans [39]uint128T
	for i := range ans {
		ans[i] = umul64(tenToThe[i/2], tenToThe[(i+1)/2])
	}
	return ans
}()

func umul64(a, b uint64) uint128T {
	var n uint128T
	n.hi, n.lo = bits.Mul64(a, b)
	return n
}

func (a *uint128T) add(x, y *uint128T) *uint128T {
	var carry uint64
	a.lo, carry = bits.Add64(x.lo, y.lo, 0)
	a.hi, _ = bits.Add64(x.hi, y.hi, carry)
	return a
}

func (a *uint128T) subV2(x, b *uint128T) *uint128T {
	var borrow uint64
	a.lo, borrow = bits.Sub64(x.lo, b.lo, 0)
	a.hi, _ = bits.Sub64(x.hi, b.hi, borrow)
	return a
}

func (a *uint128T) bitLen() uint {
	return 128 - a.leadingZeros()
}

func (a uint128T) div64(d uint64) uint128T {
	q, _ := a.divrem64(d)
	return q
}

func (a uint128T) divrem64(d uint64) (q uint128T, r uint64) {
	r = 0
	q.hi, r = bits.Div64(r, a.hi, d)
	q.lo, r = bits.Div64(r, a.lo, d)
	return
}

func (a *uint128T) divbase(x *uint128T) *uint128T {
	*a = uint128T{divbase(x.hi, x.lo), 0}
	return a
}

func divbase(hi, lo uint64) uint64 {
	// divbase is only called by Context64.Mul with (hi, lo) ≤ (10*base - 1)^2
	//                            = 0x0000_04ee_2d6d_415b__8565_e19c_207e_0001
	//                            = up to 43 bits of hi
	m := hi<<(64-43) + lo>>43                 // (hi, lo) >> 43
	a, _ := bits.Mul64(m, 0x901d7cf73ab0acd9) // (2^113)/decimal64Base
	// a := mul64(m)
	// (a, _) ~= (hi, lo) >> 43 << 113 / base = (hi, lo) << 70 / base
	// rbase is a shade under 2^113/base; add 1 here to correct it.
	return a>>(70-64) + 1
}

func (a *uint128T) leadingZeros() uint {
	if a.hi > 0 {
		return uint(bits.LeadingZeros64(a.hi))
	}
	return uint(64 + bits.LeadingZeros64(a.lo))
}

func (a *uint128T) lt(b uint128T) bool {
	if a.hi != b.hi {
		return a.hi < b.hi
	}
	return a.lo < b.lo
}

func (a uint128T) mul(b uint128T) uint128T {
	x := umul64(a.hi, b.lo)
	y := umul64(a.lo, b.hi)
	x.add(&x, &y)
	x = x.shl(64)
	y = umul64(a.lo, b.lo)
	return *x.add(&x, &y)
}

func (a uint128T) mul64(b uint64) uint128T {
	x := uint128T{0, umul64(a.hi, b).lo}
	y := umul64(a.lo, b)
	return *x.add(&x, &y)
}

// 2's-complement negation, used to implement sub.
func (a *uint128T) neg(b *uint128T) *uint128T {
	// return ^a + 1
	a0 := ^b.lo + 1
	a1 := ^b.hi
	if a0 == 0 {
		a1++
	}
	*a = uint128T{a0, a1}
	return a
}

func (a *uint128T) sub(x, y *uint128T) *uint128T {
	var n uint128T
	return a.add(x, n.neg(y))
}

func (a uint128T) shl(s uint) uint128T {
	if s < 64 {
		return uint128T{a.lo << s, a.lo>>(64-s) | a.hi<<s}
	}
	return uint128T{0, a.lo << (s - 64)}
}

// Assumes a < 1<<125
func (a *uint128T) sqrt() uint64 {
	if a.hi == 0 && a.lo < 2 {
		return a.lo
	}
	for x := uint64(1) << (a.bitLen()/2 + 1); ; {
		y := (a.div64(x).lo + x) >> 1
		if y >= x {
			return x
		}
		x = y
	}
}

// func (a uint128T) trailingZeros() uint {
// 	if a.lo > 0 {
// 		return uint(bits.TrailingZeros64(a.lo))
// 	}
// 	return uint(bits.TrailingZeros64(a.hi) + 64)
// }
