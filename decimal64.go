package decimal

import (
	"math"
	"math/bits"
)

type discardedDigit int

const (
	eq0 discardedDigit = 1 << iota
	lt5
	eq5
	gt5
)

type flavor int

const (
	flNormal flavor = 1 << iota
	flInf
	flQNaN
	flSNaN
)

type roundingMode int

const (
	roundHalfUp roundingMode = iota
	roundHalfEven
	roundDown
)

// Decimal64 represents an IEEE 754 64-bit floating point decimal number.
// It uses the binary representation method.
// Decimal64 is intentionally a struct to ensure users don't accidentally cast it to uint64
type Decimal64 struct {
	bits      uint64
	debugInfo //nolint:unused
}

// Context64 stores the rounding type for arithmetic operations.
type Context64 struct {
	roundingMode roundingMode
}

var powersOf10 = []uint64{
	1,
	10,
	100,
	1000,
	10000,
	100000,
	1000000,
	10000000,
	100000000,
	1000000000,
	10000000000,
	100000000000,
	1000000000000,
	10000000000000,
	100000000000000,
	1000000000000000,
	10000000000000000,
	100000000000000000,
	1000000000000000000,
	10000000000000000000,
}

func (context roundingMode) round(significand uint64, rndStatus discardedDigit) uint64 {
	switch context {
	case roundHalfUp:
		if rndStatus&(gt5|eq5) != 0 {
			return significand + 1
		}
	case roundHalfEven:
		if (rndStatus == eq5 && significand%2 == 1) || rndStatus == gt5 {
			return significand + 1
		}
	case roundDown: // TODO: implement proper down behaviour
		return significand
		// case roundFloor:// TODO: implement proper Floor behaviour
		// 	return significand
		// case roundCeiling: //TODO: fine tune ceiling,
	}
	return significand
}

func signalNaN64() {
	panic("sNaN64")
}

func checkSignificandIsNormal(significand uint64) {
	logicCheck(decimal64Base <= significand, "%d <= %d", decimal64Base, significand)
	logicCheck(significand < 10*decimal64Base, "%d < %d", significand, 10*decimal64Base)
}

// New64FromInt64 returns a new Decimal64 with the given value.
func New64FromInt64(value int64) Decimal64 {
	if value == 0 {
		return Zero64
	}
	sign := 0
	if value < 0 {
		sign = 1
		value = -value
	}
	// TODO: handle abs(value) > 9 999 999 999 999 999
	// lz := bits.LeadingZeros64(uint64(value))
	exp, significand := renormalize(0, uint64(value))
	checkSignificandIsNormal(significand)
	return newFromParts(sign, exp, significand)
}

func renormalize(exp int, significand uint64) (int, uint64) {

	numBits := 64 - bits.LeadingZeros64(significand)
	numDigits := numBits * 3 / 10
	normExp := 15 - numDigits
	if normExp > 0 {
		if exp-normExp < -expOffset {
			normExp = exp + expOffset
		}
		exp -= normExp
		significand *= powersOf10[normExp]
	} else if normExp < -1 {
		normExp++
		if exp-normExp > expMax {
			normExp = exp - expMax
		}
		exp -= normExp
		significand /= powersOf10[-normExp]
	}
	for significand < decimal64Base && exp > -expOffset {
		exp--
		significand *= 10
	}
	for significand >= 10*decimal64Base && exp < expMax {
		exp++
		significand /= 10
	}
	return exp, significand
}

// roundStatus gives info about the truncated part of the significand that can't be fully stored in 16 decimal digits.
func roundStatus(significand uint64, exp int, targetExp int) discardedDigit {
	expDiff := targetExp - exp
	if expDiff > 19 && significand != 0 {
		return lt5
	}
	remainder := significand % powersOf10[expDiff]
	midpoint := 5 * powersOf10[expDiff-1]
	if remainder == 0 {
		return eq0
	} else if remainder < midpoint {
		return lt5
	} else if remainder == midpoint {
		return eq5
	}
	return gt5
}

// func from stack overflow: samgak
// TODO: make this more efficent
func countTrailingZeros(n uint64) int {
	zeros := 0
	if n%10000000000000000 == 0 {
		zeros += 16
		n /= 10000000000000000
	}
	if n%100000000 == 0 {
		zeros += 8
		n /= 100000000
	}
	if n%10000 == 0 {
		zeros += 4
		n /= 10000
	}
	if n%100 == 0 {
		zeros += 2
		n /= 100
	}
	if n%10 == 0 {
		zeros++
	}
	return zeros
}

func newFromParts(sign int, exp int, significand uint64) Decimal64 {
	s := uint64(sign) << 63

	if significand < 0x8<<50 {
		// s EEeeeeeeee   (0)ttt tttttttttt tttttttttt tttttttttt tttttttttt tttttttttt
		//   EE ∈ {00, 01, 10}
		return Decimal64{bits: s | uint64(exp+expOffset)<<(63-10) | significand}.debug()
	}
	// s 11EEeeeeeeee (100)t tttttttttt tttttttttt tttttttttt tttttttttt tttttttttt
	//     EE ∈ {00, 01, 10}
	significand &= 0x8<<50 - 1
	return Decimal64{bits: s | uint64(0xc00|(exp+expOffset))<<(63-12) | significand}.debug()
}

func (d Decimal64) parts() (fl flavor, sign int, exp int, significand uint64) {
	sign = int(d.bits >> 63)
	switch (d.bits >> (63 - 4)) & 0xf {
	case 15:
		switch (d.bits >> (63 - 6)) & 3 {
		case 0, 1:
			fl = flInf
		case 2:
			fl = flQNaN
			significand = d.bits & (1<<53 - 1)
		case 3:
			fl = flSNaN
			significand = d.bits & (1<<53 - 1)
		}
	case 12, 13, 14:
		// s 11EEeeeeeeee (100)t tttttttttt tttttttttt tttttttttt tttttttttt tttttttttt
		//     EE ∈ {00, 01, 10}
		fl = flNormal
		exp = int((d.bits>>(63-12))&(1<<10-1)) - expOffset
		significand = d.bits&(1<<51-1) | (1 << 53)
	default:
		// s EEeeeeeeee   (0)ttt tttttttttt tttttttttt tttttttttt tttttttttt tttttttttt
		//   EE ∈ {00, 01, 10}
		fl = flNormal
		exp = int((d.bits>>(63-10))&(1<<10-1)) - expOffset
		significand = d.bits & (1<<53 - 1)
		if significand == 0 {
			exp = 0
		}
	}
	return
}

func expWholeFrac(exp int, significand uint64) (exp2 int, whole uint64, frac uint64) {
	if significand == 0 {
		return 0, 0, 0
	}
	if exp >= 0 {
		return exp, significand, 0
	}
	n := uint128T{significand, 0}
	exp += 16
	if exp > 0 {
		n = n.mul64(powersOf10[exp])
		exp = 0
	} else {
		// exp++ till it hits 0 or continuing would throw away digits.
		for step := 3; step >= 0; step-- {
			expStep := 1 << uint(step)
			powerOf10 := powersOf10[expStep]
			for ; n.lo >= powerOf10 && exp <= -expStep; exp += expStep {
				quo := n.lo / powerOf10
				rem := n.lo - quo*powerOf10
				if rem > 0 {
					break
				}
				n.lo = quo
			}
		}
	}
	whole128 := n.div64(10 * decimal64Base)
	frac128 := n.sub(whole128.mul64(10 * decimal64Base))
	return exp, whole128.lo, frac128.lo
}

// Float64 returns a float64 representation of d.
func (d Decimal64) Float64() float64 {
	flav, sign, exp, significand := d.parts()
	switch flav {
	case flNormal:
		if significand == 0 {
			return 0.0 * float64(1-2*sign)
		}
		if exp&1 == 1 {
			exp--
			significand *= 10
		}
		return float64(1-2*sign) * float64(significand) * math.Pow10(exp)
	case flInf:
		return math.Inf(1 - 2*sign)
	case flQNaN:
		return math.NaN()
	}
	signalNaN64()
	return 0
}

// Int64 converts d to an int64.
func (d Decimal64) Int64() int64 {
	flav, sign, exp, significand := d.parts()
	switch flav {
	case flInf:
		if sign == 0 {
			return math.MaxInt64
		}
		return math.MinInt64
	case flQNaN:
		return 0
	case flSNaN:
		signalNaN64()
		return 0
	}
	exp, whole, _ := expWholeFrac(exp, significand)
	for exp > 0 && whole < math.MaxInt64/10 {
		exp--
		whole *= 10
	}
	if exp > 0 {
		return math.MaxInt64
	}
	return int64(1-2*sign) * int64(whole)
}

// IsZero returns true if the Decimal encodes a zero value.
func (d Decimal64) IsZero() bool {
	flav, _, _, significand := d.parts()
	return significand == 0 && flav == flNormal
}

// IsInf returns true iff d = ±∞.
func (d Decimal64) IsInf() bool {
	flav, _, _, _ := d.parts()
	return flav == flInf
}

// IsNaN returns true iff d is not a number.
func (d Decimal64) IsNaN() bool {
	flav, _, _, _ := d.parts()
	return flav == flQNaN || flav == flSNaN
}

// IsQNaN returns true iff d is a quiet NaN.
func (d Decimal64) IsQNaN() bool {
	flav, _, _, _ := d.parts()
	return flav == flQNaN
}

// IsSNaN returns true iff d is a signalling NaN.
func (d Decimal64) IsSNaN() bool {
	flav, _, _, _ := d.parts()
	return flav == flSNaN
}

// IsInt returns true iff d is an integer.
func (d Decimal64) IsInt() bool {
	flav, _, exp, significand := d.parts()
	switch flav {
	case flNormal:
		_, _, frac := expWholeFrac(exp, significand)
		return frac == 0
	default:
		return false
	}
}

// quiet returns a quiet form of d, which must be a NaN.
func (d Decimal64) quiet() Decimal64 {
	return Decimal64{bits: d.bits &^ (2 << 56)}
}

// IsSubnormal returns true iff d is a subnormal.
func (d Decimal64) IsSubnormal() bool {
	flav, _, _, significand := d.parts()
	return significand != 0 && significand < decimal64Base && flav == flNormal
}

// Sign returns -1/0/1 depending on whether d is </=/> 0.
func (d Decimal64) Sign() int {
	if d == Zero64 || d == NegZero64 {
		return 0
	}
	return 1 - 2*int(d.bits>>63)
}

// Signbit returns true iff d is negative or -0.
func (d Decimal64) Signbit() bool {
	return d.bits>>63 == 1
}

// Class returns a string of the 'type' that the decimal is.
func (d Decimal64) Class() string {
	var dp decParts
	dp.unpack(d)
	if dp.isSNaN() {
		return "sNaN"
	} else if dp.isNaN() {
		return "NaN"
	}

	sign := "+"
	if dp.sign == 1 {
		sign = "-"
	}

	if dp.isInf() {
		return sign + "Infinity"
	}
	if dp.isZero() {
		return sign + "Zero"
	}
	if dp.isSubnormal() {
		return sign + "Subnormal"
	}
	return sign + "Normal"

}

// numDecimalDigits returns the magnitude (number of digits) of a uint64.
func numDecimalDigits(n uint64) int {
	numBits := 64 - bits.LeadingZeros64(n)
	numDigits := numBits * 3 / 10
	if n < powersOf10[numDigits] {
		return numDigits
	}
	return numDigits + 1
}

// checkNan returns the decimal NaN that is to be propogated and true else first decimal and false
func checkNan(d, e *decParts) (Decimal64, bool) {
	if d.fl == flSNaN {
		return d.original, true
	}
	if e.fl == flSNaN {
		return e.original, true
	}
	if d.fl == flQNaN {
		return d.original, true
	}
	if e.fl == flQNaN {
		return e.original, true
	}
	return d.original, false
}

// checkNan3 returns the decimal NaN that is to be propogated and true else first decimal and false
func checkNan3(d, e, f *decParts) (Decimal64, bool) {
	if d.fl == flSNaN {
		return d.original, true
	}
	if e.fl == flSNaN {
		return e.original, true
	}
	if f.fl == flSNaN {
		return f.original, true
	}
	if d.fl == flQNaN {
		return d.original, true
	}
	if e.fl == flQNaN {
		return e.original, true
	}
	if f.fl == flQNaN {
		return f.original, true
	}
	return d.original, false
}
