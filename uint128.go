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
	if a.lt(&tenToThe128[numDigitsEst]) {
		return numDigitsEst
	}
	return numDigitsEst + 1
}

var tenToThe128 = func() [39]uint128T {
	var ans [39]uint128T
	for i := range ans {
		ans[i].umul64(tenToThe[i/2], tenToThe[(i+1)/2])
	}
	return ans
}()

func (a *uint128T) umul64(x, y uint64) *uint128T {
	a.hi, a.lo = bits.Mul64(x, y)
	return a
}

func (a *uint128T) add(x, y *uint128T) *uint128T {
	var carry uint64
	a.lo, carry = bits.Add64(x.lo, y.lo, 0)
	a.hi, _ = bits.Add64(x.hi, y.hi, carry)
	return a
}

func (a *uint128T) sub(x, y *uint128T) *uint128T {
	var borrow uint64
	a.lo, borrow = bits.Sub64(x.lo, y.lo, 0)
	a.hi, _ = bits.Sub64(x.hi, y.hi, borrow)
	return a
}

func (a *uint128T) bitLen() uint {
	return 128 - a.leadingZeros()
}

func (a *uint128T) div64(x *uint128T, d uint64) *uint128T {
	a.divrem64(x, d)
	return a
}

func (a *uint128T) divrem64(x *uint128T, d uint64) uint64 {
	var r uint64
	a.hi, r = bits.Div64(0, x.hi, d)
	a.lo, r = bits.Div64(r, x.lo, d)
	return r
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

func (a *uint128T) lt(b *uint128T) bool {
	if a.hi != b.hi {
		return a.hi < b.hi
	}
	return a.lo < b.lo
}

func (a *uint128T) mul(x, y *uint128T) *uint128T {
	var t, u uint128T
	t.umul64(x.hi, y.lo)
	u.umul64(x.lo, y.hi)
	t.add(&t, &u)
	t = uint128T{0, t.lo}
	u.umul64(x.lo, y.lo)
	return a.add(&t, &u)
}

func (a *uint128T) mul64(x *uint128T, b uint64) *uint128T {
	var t uint128T
	y := uint128T{0, t.umul64(x.hi, b).lo}
	return a.add(&y, t.umul64(x.lo, b))
}

func (a *uint128T) shl(b *uint128T, s uint) *uint128T {
	if s < 64 {
		return a.set(b.lo>>(64-s)|b.hi<<s, b.lo<<s)
	}
	return a.set(b.lo<<(s-64), 0)
}

func (a *uint128T) set(hi, lo uint64) *uint128T {
	*a = uint128T{lo, hi}
	return a
}

// Assumes a < 1<<125
func (a *uint128T) sqrt() uint64 {
	if a.hi == 0 && a.lo < 2 {
		return a.lo
	}
	for x := uint64(1) << (a.bitLen()/2 + 1); ; {
		var t uint128T
		y := (t.div64(a, x).lo + x) >> 1
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
