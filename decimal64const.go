package decimal

// Zero64 represents 0 as a Decimal64.
var Zero64 = newFromParts(0, 0, 0)

// NegZero64 represents -0 as a Decimal64.
var NegZero64 = newFromParts(1, 0, 0)

// One64 represents 1 as a Decimal64.
var One64 = newFromParts(0, -15, decimal64Base)

// NegOne64 represents -1 as a Decimal64.
var NegOne64 = newFromParts(1, -15, decimal64Base)

// Infinity64 represents ∞ as a Decimal64.
var Infinity64 = Decimal64{inf64}

// NegInfinity64 represents -∞ as a Decimal64.
var NegInfinity64 = Decimal64{neg64 | inf64}

// QNaN64 represents a quiet NaN as a Decimal64.
var QNaN64 = Decimal64{0x7c << 56}

// SNaN64 represents a signalling NaN as a Decimal64.
var SNaN64 = Decimal64{0x7e << 56}

// Pi64 represents π.
var Pi64 = newFromParts(0, -15, 3141592653589793)

// E64 represents e (lim[n→∞](1+1/n)ⁿ).
var E64 = newFromParts(0, -15, 2718281828459045)

var neg64 uint64 = 0x80 << 56
var inf64 uint64 = 0x78 << 56

// 1E15
const decimal64Base = 1000 * 1000 * 1000 * 1000 * 1000

// maxSig is the maximum significand possible that fits in 16 decimal places.
const maxSig = 10*decimal64Base - 1

const expOffset = 398
const minExp = -398
const expMax = 369

// Max64  is the maximum number representable with a Decimal64.
var Max64 = newFromParts(0, expMax, maxSig)

// NegMax64  is the minimum finite number (most negative) possible with Decimal64 (negative).
var NegMax64 = newFromParts(1, expMax, maxSig)

// Min64 is the smallest number that is subnormal possible with Decimal64.
var Min64 = newFromParts(0, -398, 1)

var zeroes = []Decimal64{Zero64, NegZero64}
var infinities = []Decimal64{Infinity64, NegInfinity64}

// DefaultContext is the context that Arithmetic functions will use in order to do calculations
var DefaultContext = Context64{roundHalfUp}
