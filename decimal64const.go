package decimal

// Zero64 represents 0 as a Decimal64.
var Zero64 = newFromParts(0, 0, 0)

// var Zero64 = Decimal64{0}

// NegZero64 represents -0 as a Decimal64.
var NegZero64 = newFromParts(1, 0, 0)

// var NegZero64 = Decimal64{1 << 63}

// One64 represents 1 as a Decimal64.
var One64 = NewDecimal64FromInt64(1)

// NegOne64 represents -1 as a Decimal64.
var NegOne64 = NewDecimal64FromInt64(1).Neg()

// Infinity64 represents ∞ as a Decimal64.
var Infinity64 = Decimal64{inf64}

// NegInfinity64 represents -∞ as a Decimal64.
var NegInfinity64 = Decimal64{neg64 | inf64}

// QNaN64 represents a quiet NaN as a Decimal64.
var QNaN64 = Decimal64{0x7c00000000000000}

// SNaN64 represents a signalling NaN as a Decimal64.
var SNaN64 = Decimal64{0x7e00000000000000}

var neg64 uint64 = 0x8000000000000000
var inf64 uint64 = 0x7800000000000000

var zeroes = []Decimal64{Zero64, NegZero64}
var infinities = []Decimal64{Infinity64, NegInfinity64}

// 10E15
const decimal64Base = 10 * 1000 * 1000 * 1000 * 1000 * 1000
