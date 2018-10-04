package decimal

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

func add128(a, b uint128T) uint128T {
	carry1 := (a.lo&1 + b.lo&1) >> 1
	carry64 := (a.lo>>1 + b.lo>>1 + carry1) >> 63
	return uint128T{a.lo + b.lo, a.hi + b.hi + carry64}
}

// 2's-complement negation, used to implement sub128.
func neg128(a uint128T) uint128T {
	return add128(uint128T{^a.lo, ^a.hi}, uint128T{1, 0})
}

func sub128(a, b uint128T) uint128T {
	return add128(a, neg128(b))
}

func shl128(a uint128T, s uint) uint128T {
	return uint128T{a.lo << s, a.lo>>(64-s) | a.hi<<s}
}

func shr128(a uint128T, s uint) uint128T {
	return uint128T{a.lo>>s | a.hi<<(64-s), a.hi >> s}
}

func div10_64(a uint128T) uint128T {
	// http://www.hackersdelight.org/divcMore.pdf
	q := add128(shr128(a, 1), shr128(a, 2))
	q = add128(q, shr128(q, 4))
	q = add128(q, shr128(q, 8))
	q = add128(q, shr128(q, 16))
	q = add128(q, shr128(q, 32))
	q = add128(q, shr128(q, 64))
	q = shr128(q, 3)
	r := sub128(a, mul10_64(q))
	return add128(q, shr128(add128(r, uint128T{6, 0}), 4))
}

func mul10_64(a uint128T) uint128T {
	return add128(shl128(a, 1), shl128(a, 3))
}
