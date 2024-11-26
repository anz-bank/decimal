package decimal

// Zero64 is 0 represented as a Decimal64.
var Zero64 = newFromParts(0, 0, 0)

// NegZero64 is -0 represented as a Decimal64.
// Note that Zero64 != NegZero64, but Zero64.Equal(NegZero64) returns true.
var NegZero64 = newFromParts(1, 0, 0)

// One64 is 1 represented as a Decimal64.
var One64 = newFromParts(0, -15, decimal64Base)

// NegOne64 is -1 represented as a Decimal64.
var NegOne64 = newFromParts(1, -15, decimal64Base)

// Infinity64 is ∞ represented as a Decimal64.
var Infinity64 = new64(inf64)

// NegInfinity64 is -∞ represented as a Decimal64.
var NegInfinity64 = new64(neg64 | inf64)

// QNaN64 a quiet NaN represented as a Decimal64.
var QNaN64 = new64(0x7c << 56)

// SNaN64 a signalling NaN represented as a Decimal64.
// Note that the decimal never signals on NaNs but some operations treat sNaN
// differently to NaN.
var SNaN64 = new64(0x7e << 56)

// Pi64 represents the transcendental number π.
var Pi64 = newFromParts(0, -15, 3_141592653589793)

// E64 represents the transcendental number e (lim[n→∞](1+1/n)ⁿ).
var E64 = newFromParts(0, -15, 2_718281828459045)

const neg64 uint64 = 0x80 << 56
const inf64 uint64 = 0x78 << 56

const decimal64Base uint64 = 1_000_000_000_000_000 // 1E15
const decimal64Digits = 16

// maxSig is the maximum significand possible that fits in 16 decimal places.
const maxSig = 10*decimal64Base - 1

const expOffset = 398
const expMax = 369

// Max64 is the highest finite number representable as a Decimal64.
// It has the value 9.999999999999999E+384.
var Max64 = newFromParts(0, expMax, maxSig)

// NegMax64 is the lowest finite number representable as a Decimal64.
// It has the value -9.999999999999999E+384.
var NegMax64 = newFromParts(1, expMax, maxSig)

// Min64 is the closest positive number to zero.
// It has the value 1E-398.
var Min64 = newFromParts(0, -398, 1)

// Min64 is the closest negative number to zero.
// It has the value -1E-398.
var NegMin64 = newFromParts(1, -398, 1)

var zeroes64 = [2]Decimal64{Zero64, NegZero64}
var infinities64 = [2]Decimal64{Infinity64, NegInfinity64}

// DefaultContext64 is the context that arithmetic functions will use in order to
// do calculations.
// Setting this context to a different value will globally affect all
// [Decimal64] methods whose behavior depends on context.
// Note that all such methods are also available as direct methods of Context64.
// It uses [HalfUp] rounding.
var DefaultContext64 = Context64{Rounding: HalfUp}
