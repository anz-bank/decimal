package d64

// Zero is 0 represented as a [Decimal].
var Zero = newFromParts(0, 0, 0)

// NegZero is -0 represented as a [Decimal].
// Note that [Zero] != NegZero, but [Zero].Equal(NegZero) returns true.
var NegZero = newFromParts(1, 0, 0)

// One is 1 represented as a [Decimal].
var One = newFromParts(0, -15, decimalBase)

// NegOne is -1 represented as a [Decimal].
var NegOne = newFromParts(1, -15, decimalBase)

// Inf is ∞ represented as a [Decimal].
var Inf = newDec(inf)

// NegInf is -∞ represented as a [Decimal].
var NegInf = newDec(neg | inf)

// QNaN is a quiet NaN represented as a [Decimal].
var QNaN = newDec(0x7c << 56)

// SNaN is a signalling NaN represented as a [Decimal].
// Note that the decimal never signals on NaNs but some operations treat sNaN
// differently to NaN.
var SNaN = newDec(0x7e << 56)

// Pi represents the transcendental number π.
var Pi = newFromParts(0, -15, 3_141592653589793)

// E represents the transcendental number e (lim[n→∞](1+1/n)ⁿ).
var E = newFromParts(0, -15, 2_718281828459045)

const neg uint64 = 0x80 << 56
const inf uint64 = 0x78 << 56

const decimalBase uint64 = 1_000_000_000_000_000 // 1E15
const decimalDigits = 16

// maxSig is the maximum significand possible that fits in 16 decimal places.
const maxSig = 10*decimalBase - 1

const expOffset = 398
const expMax = 369

// Max is the highest finite number representable as a [Decimal].
// It has the value 9.999999999999999E+384.
var Max = newFromParts(0, expMax, maxSig)

// NegMax is the lowest finite number representable as a [Decimal].
// It has the value -9.999999999999999E+384.
var NegMax = newFromParts(1, expMax, maxSig)

// Min is the closest positive number to zero.
// It has the value 1E-398.
var Min = newFromParts(0, -398, 1)

// NegMin is the closest negative number to zero.
// It has the value -1E-398.
var NegMin = newFromParts(1, -398, 1)

var zeroes = [2]Decimal{Zero, NegZero}
var infinities = [2]Decimal{Inf, NegInf}

// DefaultContext is the context that arithmetic functions will use in order to
// do calculations.
// Setting this context to a different value will globally affect all
// [Decimal] methods whose behavior depends on context.
// Note that all such methods are also available as direct methods of [Context].
// It uses [HalfUp] rounding.
var DefaultContext = Context{Rounding: HalfUp}
