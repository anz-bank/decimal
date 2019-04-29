package decimal

import (
	"math"
	"math/bits"
)

type flavor int
type roundingMode int
type discardedDigit int
type decErr int

const (
	eq0 discardedDigit = 1 << iota
	lt5
	eq5
	gt5
)

const (
	flNormal flavor = iota
	flInf
	flQNaN
	flSNaN
)

const (
	roundHalfUp roundingMode = iota
	roundHalfEven
	roundDown
)

// // TODO: implement returns of these error types
// const (
// 	noError decErr = iota
// 	invalidOperation
// 	divisionByzero
// 	inexact
// 	overflow
// 	underflow
// )

// Decimal64 represents an IEEE 754 64-bit floating point decimal number.
// It uses the binary representation method.
type Decimal64 struct {
	bits uint64
}

// decParts stores the constituting decParts of a decimal64.
type decParts struct {
	fl          flavor
	sign        int
	exp         int
	significand uint64
	mag         int
	dec         *Decimal64
}

// Context64 stores the rounding type and and exceptions needed to be signalled
type Context64 struct {
	roundingMode roundingMode
	// TODO:use exceptions in order to report errors to calling fucntion
	// exceptions decErr
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

// separation gets the separation in decimal places of the MSD's of two decimal 64s
func (dec *decParts) separation(eDec decParts) int {
	return dec.mag + dec.exp - eDec.mag - eDec.exp
}

// removeZeros removes zeros and increments the exponent to match.
func (dec *decParts) removeZeros() {
	zeros := countTrailingZeros(dec.significand)
	dec.significand /= powersOf10[zeros]
	dec.exp += zeros
}

// updateMag updates the magnitude of the dec object
func (dec *decParts) updateMag() {
	dec.mag = numDecimalDigits(dec.significand)
}

// updateMag updates the magnitude of the dec object
func (dec *decParts) isZero() bool {
	return dec.significand == 0 && dec.fl == flNormal
}
func signalNaN64() {
	panic("sNaN64")
}

func checkSignificandIsNormal(significand uint64) {
	logicCheck(decimal64Base <= significand, "%d <= %d", decimal64Base, significand)
	logicCheck(significand < 10*decimal64Base, "%d < %d", significand, 10*decimal64Base)
}

// NewDecimal64FromInt64 returns a new Decimal64 with the given value.
func NewDecimal64FromInt64(value int64) Decimal64 {
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
	logicCheck(significand != 0, "significand (%d) != 0", significand)

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

func rescale(exp int, significand uint64, targetExp int) (uint64, int) {
	expDiff := targetExp - exp
	mag := numDecimalDigits(significand)
	if expDiff > mag {
		return 0, targetExp
	}
	divisor := powersOf10[expDiff]
	return significand / divisor, targetExp
}

func (dec *decParts) rescale(targetExp int) (rndStatus discardedDigit) {
	expDiff := targetExp - dec.exp
	mag := dec.mag
	rndStatus = roundStatus(dec.significand, dec.exp, targetExp)
	if expDiff > mag {
		dec.significand, dec.exp = 0, targetExp
		return
	}
	divisor := powersOf10[expDiff]
	dec.significand = dec.significand / divisor
	dec.exp = targetExp
	return
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

//func from stack overflow: samgak
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

// match scales matches the exponents of d and e and returns the info about the discarded digit
func matchScales(d, e *decParts) discardedDigit {
	logicCheck(d.significand != 0, "d.significand (%d) != 0", d.significand)
	logicCheck(e.significand != 0, "e.significand (%d) != 0", e.significand)
	if d.exp == e.exp {
		return eq0
	}
	if d.exp < e.exp {
		return d.rescale(e.exp)
	}
	return e.rescale(d.exp)
}

func newFromParts(sign int, exp int, significand uint64) Decimal64 {
	s := uint64(sign) << 63

	if significand < 0x8<<50 {
		// s EEeeeeeeee   (0)ttt tttttttttt tttttttttt tttttttttt tttttttttt tttttttttt
		//   EE ∈ {00, 01, 10}
		return Decimal64{s | uint64(exp+expOffset)<<(63-10) | significand}
	}
	// s 11EEeeeeeeee (100)t tttttttttt tttttttttt tttttttttt tttttttttt tttttttttt
	//     EE ∈ {00, 01, 10}
	significand &= 0x8<<50 - 1
	return Decimal64{s | uint64(0xc00|(exp+expOffset))<<(63-12) | significand}
}

// TODO: merge parts()/ implement version that uses decParts struct
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
func (dec decParts) isNan() bool {
	return dec.fl == flQNaN || dec.fl == flSNaN
}

// decParts gets the parts and returns in decParts stuct, doesn't get the magnitude due to performance issues\
// TODO: rename this to parts when parts is depreciated
func (d *Decimal64) getParts() decParts {
	var fl flavor
	var sign, exp int
	var significand uint64
	sign = int(d.bits >> 63)

	switch (d.bits >> (63 - 4)) & 0xf {
	case 15:
		switch (d.bits >> (63 - 6)) & 3 {
		case 0, 1:
			fl = flInf
		case 2:
			fl = flQNaN
		case 3:
			fl = flSNaN
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
	return decParts{fl, sign, exp, significand, 0, d}
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
	flavor, sign, exp, significand := d.parts()
	switch flavor {
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
	flavor, sign, exp, significand := d.parts()
	switch flavor {
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

// IsInf returns true iff d = ±∞.
func (d Decimal64) IsInf() bool {
	flavor, _, _, _ := d.parts()
	return flavor == flInf
}

// IsNaN returns true iff d is not a number.
func (d Decimal64) IsNaN() bool {
	flavor, _, _, _ := d.parts()
	return flavor == flQNaN || flavor == flSNaN
}

// IsQNaN returns true iff d is a quiet NaN.
func (d Decimal64) IsQNaN() bool {
	flavor, _, _, _ := d.parts()
	return flavor == flQNaN
}

// IsSNaN returns true iff d is a signalling NaN.
func (d Decimal64) IsSNaN() bool {
	flavor, _, _, _ := d.parts()
	return flavor == flSNaN
}

// IsInt returns true iff d is an integer.
func (d Decimal64) IsInt() bool {
	fl, _, exp, significand := d.parts()
	switch fl {
	case flNormal:
		_, _, frac := expWholeFrac(exp, significand)
		return frac == 0
	default:
		return false
	}
}
func (d Decimal64) isZero() bool {
	fl, _, _, significand := d.parts()
	return significand == 0 && fl == flNormal
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

//numDecimalDigits returns the magnitude (number of digits) of a uint64.
func numDecimalDigits(n uint64) int {
	numBits := 64 - bits.LeadingZeros64(n)
	numDigits := numBits * 3 / 10
	if n < powersOf10[numDigits] {
		return numDigits
	}
	return numDigits + 1
}

// propagateNan returns the decimal pointer to the NaN that is to be propogated
func propagateNan(dp, ep *decParts) *Decimal64 {
	if dp.fl == flSNaN {
		return dp.dec
	}
	if ep.fl == flSNaN {
		return ep.dec
	}
	if dp.fl == flQNaN {
		return dp.dec
	}
	return ep.dec
}
