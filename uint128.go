package decimal

import (
	"math/bits"
)

type uint128T struct {
	lo, hi uint64
}

func (a uint128T) numDecimalDigits() int {
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

func (a uint128T) add(b uint128T) uint128T {
	lo, carry := bits.Add64(a.lo, b.lo, 0)
	hi, _ := bits.Add64(a.hi, b.hi, carry)
	return uint128T{lo, hi}
}

func (a uint128T) bitLen() uint {
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

// See http://www.hackersdelight.org/divcMore.pdf for div-by-const tricks.

func (a uint128T) divBy10() uint128T {
	q := a.shr(1).add(a.shr(2))
	q = q.add(q.shr(4))
	q = q.add(q.shr(8))
	q = q.add(q.shr(16))
	q = q.add(q.shr(32))
	q = q.add(q.shr(64))
	q = q.shr(3)
	r := a.sub(q.mulBy10())
	return q.add(uint128T{(r.lo + 6) >> 4, 0})
}

func (a uint128T) leadingZeros() uint {
	if a.hi > 0 {
		return uint(bits.LeadingZeros64(a.hi))
	}
	return uint(64 + bits.LeadingZeros64(a.lo))
}

func (a uint128T) lt(b uint128T) bool {
	if a.hi != b.hi {
		return a.hi < b.hi
	}
	return a.lo < b.lo
}

func (a uint128T) mulBy10() uint128T {
	// a*10 = a*8 + a*2
	a8 := a.shl(3)
	a2 := a.shl(1)
	return a8.add(a2)
}

func (a uint128T) mul(b uint128T) uint128T {
	return umul64(a.hi, b.lo).add(umul64(a.lo, b.hi)).shl(64).add(umul64(a.lo, b.lo))
}

func (a uint128T) mul64(b uint64) uint128T {
	return uint128T{0, umul64(a.hi, b).lo}.add(umul64(a.lo, b))
}

// 2's-complement negation, used to implement sub.
func (a uint128T) neg() uint128T {
	// return ^a + 1
	a0 := ^a.lo + 1
	a1 := ^a.hi
	if a0 == 0 {
		a1++
	}
	return uint128T{a0, a1}
}

func (a uint128T) sub(b uint128T) uint128T {
	return a.add(b.neg())
}

func (a uint128T) shl(s uint) uint128T {
	if s < 64 {
		return uint128T{a.lo << s, a.lo>>(64-s) | a.hi<<s}
	}
	return uint128T{0, a.lo << (s - 64)}
}

func (a uint128T) shr(s uint) uint128T {
	if s < 64 {
		return uint128T{a.lo>>s | a.hi<<(64-s), a.hi >> s}
	}
	return uint128T{a.hi >> (s - 64), 0}
}

// Assumes a < 1<<125
func (a uint128T) sqrt() uint64 {
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
