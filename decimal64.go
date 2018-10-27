package decimal

import (
	"math"
)

// Decimal64 represents an IEEE 754 64-bit floating point decimal number.
// It uses the binary representation method.
type Decimal64 struct {
	bits uint64
}

type flavor int

const (
	flNormal flavor = iota
	flInf
	flQNaN
	flSNaN
)

func signalNaN64() Decimal64 {
	// TODO: What's the right behavior?
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

	// TODO: Optimize to O(1) with bits.LeadingZeros64
	if significand < 10*decimal64Base/100000000 && exp > -1-expOffset+8 {
		exp -= 8
		significand *= 100000000
	}
	if significand < 10*decimal64Base/10000 && exp > -1-expOffset+4 {
		exp -= 4
		significand *= 10000
	}
	if significand < 10*decimal64Base/100 && exp > -1-expOffset+2 {
		exp -= 2
		significand *= 100
	}
	if significand < 10*decimal64Base/10 && exp > -1-expOffset+1 {
		exp--
		significand *= 10
	}
	if significand >= 10*decimal64Base && exp < expMax {
		exp++
		significand /= 10
	}
	return exp, significand
}

func rescale(exp int, significand uint64, targetExp int) (uint64, int) {
	exp -= targetExp
	var divisor uint64 = 1
	for ; exp < -7 && divisor < significand; exp += 8 {
		divisor *= 100000000
	}
	for ; exp < -3 && divisor < significand; exp += 4 {
		divisor *= 10000
	}
	for ; exp < -1 && divisor < significand; exp += 2 {
		divisor *= 100
	}
	for ; exp < 0 && divisor < significand; exp++ {
		divisor *= 10
	}
	return significand / divisor, targetExp
}

func matchScales(exp1 int, significand1 uint64, exp2 int, significand2 uint64) (int, uint64, int, uint64) {
	if significand1 == 0 {
		exp1 = exp2
	} else if significand2 == 0 {
		exp2 = exp1
	} else if exp1 < exp2 {
		significand1, exp1 = rescale(exp1, significand1, exp2)
	} else if exp2 < exp1 {
		significand2, exp2 = rescale(exp2, significand2, exp1)
	}
	return exp1, significand1, exp2, significand2
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

func (d Decimal64) parts() (fl flavor, sign int, exp int, significand uint64) {
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
	return
}

func expWholeFrac(exp int, significand uint64) (exp2 int, whole uint64, frac uint64) {
	if significand == 0 {
		return 0, 0, 0
	}
	if exp >= 0 {
		return exp, significand, 0
	}
	n := uint128T{significand, 0}.mul64(10 * decimal64Base)
	// exp++ till it hits 0 or continuing would throw away digits.
	for ; exp < 0; exp++ {
		nOver10 := n.divBy10()
		rem := n.sub(nOver10.mul64(10))
		if rem.lo > 0 {
			break
		}
		n = nOver10
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
	flavor, _, exp, significand := d.parts()
	switch flavor {
	case flNormal:
		_, _, frac := expWholeFrac(exp, significand)
		return frac == 0
	default:
		return false
	}
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
