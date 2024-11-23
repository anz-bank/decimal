package decimal

import (
	"math"
	"math/bits"
	"sync"
)

type uint128T struct {
	lo, hi uint64
}

func (a *uint128T) numDecimalDigits() int16 {
	if a.hi == 0 {
		return numDecimalDigitsU64(a.lo)
	}
	bitSize := 65 + bits.Len64(a.hi)
	numDigitsEst := int16(bitSize * 77 / 256)
	if !a.lt(&tenToThe128[uint(numDigitsEst)%uint(len(tenToThe128))]) {
		numDigitsEst++
	}
	return numDigitsEst
}

var tenToThe128 = func() [64]uint128T {
	var ans [64]uint128T
	for i := range ans[:39] {
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

func (a *uint128T) divrem64(x *uint128T, d uint64) uint64 {
	var r uint64
	if a.hi == 0 {
		a.lo, r = x.lo/d, x.lo%d
	} else {
		a.hi, r = bits.Div64(0, x.hi, d)
		a.lo, r = bits.Div64(r, x.lo, d)
	}
	return r
}

// divbase divides a by [decimal64Base].
func (a *uint128T) divbase(x *uint128T) *uint128T { return a.divc(x, 113, 0x901d7cf73ab0acd9) }

// div10base divides a by 10*[decimal64Base].
func (a *uint128T) div10base(x *uint128T) *uint128T { return a.divc(x, 117, 0xe69594bec44de15b) }

// divc divides a by n where 1<<po2/rdenom ~ n and po2 is chosen such that rdenom is the largest possible value < 1<<64.
func (a *uint128T) divc(x *uint128T, po2 int, rdenom uint64) *uint128T {
	m := x.hi<<(64-43) + x.lo>>43 // (hi, lo) >> 43
	q, _ := bits.Mul64(m, rdenom)
	return a.set(0, q>>(po2-(43+64))+1)
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
	t = uint128T{0, t.lo} // t <<= 64
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

func sqrtu64(n uint64) uint64 {
	const maxu32 = 1<<32 - 1
	switch {
	case n < 1<<16:
		return uint64(sqrtu16(uint16(n)) >> 8)
	case n >= maxu32*maxu32:
		return maxu32
	}

	// Shift up as far as possible, but must be an even amount.
	halfshift := bits.LeadingZeros64(n) / 2
	n <<= 2 * halfshift

	s := sqrtu16(uint16(n >> (64 - 16)))
	x := uint64(s) << (32 - 16)

	// Two iterations suffice.
	x = (x + n/x) >> 1
	// Undo shift in second iteration. Only need 1/2-shift because it's the √.
	return (x + n/x) >> (1 + halfshift)
}

const (
	sqrtSlotSize = 64 / 2 // 1 cache line
	sqrtElts     = 1 << 16
)

var (
	sqrtTable [sqrtElts]uint16
	sqrtOnce  [sqrtElts / sqrtSlotSize]sync.Once
)

// sqrtu16 returns √n << 8.
func sqrtu16(n uint16) uint16 {
	slot := n / sqrtSlotSize
	sqrtOnce[slot].Do(func() {
		a := slot * sqrtSlotSize
		b := a + sqrtSlotSize
		for i := a; i != b; i++ { // != because b wraps
			sqrtTable[i] = uint16(math.Sqrt(float64(uint64(i) << 16)))
		}
	})
	return sqrtTable[n]
}
