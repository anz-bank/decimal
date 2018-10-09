package decimal

import (
	"math/bits"
)

type uint128T struct {
	lo, hi uint64
}

func umul64(a, b uint64) uint128T {
	a0 := a & 0xffffffff
	a1 := a >> 32
	b0 := b & 0xffffffff
	b1 := b >> 32

	r0 := a0 * b0
	r1 := a1*b0 + a0*b1 + r0>>32
	r2 := a1*b1 + r1>>32
	return uint128T{r0&0xffffffff | r1<<32, r2}
}

func (a uint128T) add(b uint128T) uint128T {
	carry1 := (a.lo&1 + b.lo&1) >> 1
	carry64 := (a.lo>>1 + b.lo>>1 + carry1) >> 63
	return uint128T{a.lo + b.lo, a.hi + b.hi + carry64}
}

const base32 = 1 << 32

// Compute q, r such that q*d + r = a.
// Assumes a < d<<64.
// http://www.hackersdelight.org/hdcodetxt/divlu.c.txt: divlu1()
func (a uint128T) divPart64(d uint64) (q, r uint64) {
	shift := uint(bits.LeadingZeros64(d))
	d <<= shift

	d0, d1 := d%base32, d/base32
	a = a.shl(shift)
	a0, a1 := a.lo%base32, a.lo/base32
	q1 := partialQuotient32(a1, a.hi, d0, d1)
	a12 := a1 + a.hi*base32 - q1*d
	q0 := partialQuotient32(a0, a12, d0, d1)
	a01 := a0 + a12*base32 - q0*d

	return q0 + q1*base32, a01 >> shift
}

// Helper for divPart64
func partialQuotient32(a0, a12, d0, d1 uint64) uint64 {
	q := a12 / d1
	for r := a12 - q*d1; r < base32 && !(q < base32 && q*d0 <= r<<32+a0); r += d1 {
		q--
	}
	return q
}

func (a uint128T) div64(d uint64) (q uint128T, r uint64) {
	r = 0
	q.hi, r = uint128T{a.hi, r}.divPart64(d)
	q.lo, r = uint128T{a.lo, r}.divPart64(d)
	return
}

func (a uint128T) divBy10() uint128T {
	// http://www.hackersdelight.org/divcMore.pdf
	q := a.shr(1).add(a.shr(2))
	q = q.add(q.shr(4))
	q = q.add(q.shr(8))
	q = q.add(q.shr(16))
	q = q.add(q.shr(32))
	q = q.add(q.shr(64))
	q = q.shr(3)
	r := a.sub(q.mulBy10())
	return q.add(r.add(uint128T{6, 0}).shr(4))
}

func (a uint128T) ge(b uint128T) bool {
	return !a.lt(b)
}

func (a uint128T) gt(b uint128T) bool {
	return b.lt(a)
}

func (a uint128T) le(b uint128T) bool {
	return !b.lt(a)
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
	return a.shl(1).add(a.shl(3))
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

func (a uint128T) trailingZeros() uint {
	if a.lo > 0 {
		return uint(bits.TrailingZeros64(a.lo))
	}
	return uint(bits.TrailingZeros64(a.hi) + 64)
}
