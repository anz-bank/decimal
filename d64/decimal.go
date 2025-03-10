package d64

import (
	"fmt"
	"math"
	"math/bits"
	"strconv"
)

type discardedDigit int

const (
	eq0 discardedDigit = 1 << iota
	lt5
	eq5
	gt5
)

type flavor int8

const (
	flInf      flavor = 0
	flNormal53 flavor = 1 << (iota - 1)
	flNormal51
	flQNaN
	flSNaN
	flNormal = flNormal53 | flNormal51
	flNaN    = flQNaN | flSNaN
)

func (f flavor) normal() bool {
	return f&flNormal != 0
}

func (f flavor) nan() bool {
	return f&flNaN != 0
}

func (f flavor) String() string {
	switch f {
	case flInf:
		return "Infinity"
	case flNormal53:
		return "Normal53"
	case flNormal51:
		return "Normal51"
	case flQNaN:
		return "QNaN"
	case flSNaN:
		return "SNaN"
	default:
		return fmt.Sprintf("Unknown flavor %d", f)
	}
}

// Rounding defines how arithmetic operations round numbers in certain operations.
type Rounding int8

const (
	// HalfUp rounds to the nearest number, rounding away from zero if the
	// number is exactly halfway between two possible roundings.
	HalfUp Rounding = iota

	// HalfEven rounds to the nearest number, rounding to the nearest even
	// number if the number is exactly halfway between two possible roundings.
	HalfEven

	// Down rounds towards zero.
	Down
)

func (r Rounding) String() string {
	switch r {
	case HalfUp:
		return "HalfUp"
	case HalfEven:
		return "HalfEven"
	case Down:
		return "Down"
	default:
		return fmt.Sprintf("Unknown rounding mode %d", r)
	}
}

// Context may be used to tune the behaviour of arithmetic operations.
type Context struct {
	// Rounding sets the rounding behaviour of arithmetic operations.
	Rounding Rounding

	// TODO: implement
	// // Signal causes arithmetic operations to panic when encountering a sNaN.
	// Signal bool
}

var tenToThe = [32]uint64{ // pad for efficient indexing
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

func (ctx Rounding) round(significand uint64, rndStatus discardedDigit) uint64 {
	switch ctx {
	case HalfUp:
		if rndStatus&(gt5|eq5) != 0 {
			return significand + 1
		}
	case HalfEven:
		if (rndStatus == eq5 && significand%2 == 1) || rndStatus == gt5 {
			return significand + 1
		}
	case Down: // TODO: implement proper down behaviour
		return significand
		// case roundFloor:// TODO: implement proper Floor behaviour
		// 	return significand
		// case roundCeiling: //TODO: fine tune ceiling,
	}
	return significand
}

var ErrNaN error = Error("sNaN64")

var smalls = []Decimal{
	newStr(newFromPartsRaw(1, -14, 1*decimalBase).bits, "-10"),

	newStr(newFromPartsRaw(1, -15, 9*decimalBase).bits, "-9"),
	newStr(newFromPartsRaw(1, -15, 8*decimalBase).bits, "-8"),
	newStr(newFromPartsRaw(1, -15, 7*decimalBase).bits, "-7"),
	newStr(newFromPartsRaw(1, -15, 6*decimalBase).bits, "-6"),
	newStr(newFromPartsRaw(1, -15, 5*decimalBase).bits, "-5"),
	newStr(newFromPartsRaw(1, -15, 4*decimalBase).bits, "-4"),
	newStr(newFromPartsRaw(1, -15, 3*decimalBase).bits, "-3"),
	newStr(newFromPartsRaw(1, -15, 2*decimalBase).bits, "-2"),
	newStr(newFromPartsRaw(1, -15, 1*decimalBase).bits, "-1"),

	// TODO: Decimal64{}?
	newStr(newFromPartsRaw(0, 0, 0).bits, "0"),

	newStr(newFromPartsRaw(0, -15, 1*decimalBase).bits, "1"),
	newStr(newFromPartsRaw(0, -15, 2*decimalBase).bits, "2"),
	newStr(newFromPartsRaw(0, -15, 3*decimalBase).bits, "3"),
	newStr(newFromPartsRaw(0, -15, 4*decimalBase).bits, "4"),
	newStr(newFromPartsRaw(0, -15, 5*decimalBase).bits, "5"),
	newStr(newFromPartsRaw(0, -15, 6*decimalBase).bits, "6"),
	newStr(newFromPartsRaw(0, -15, 7*decimalBase).bits, "7"),
	newStr(newFromPartsRaw(0, -15, 8*decimalBase).bits, "8"),
	newStr(newFromPartsRaw(0, -15, 9*decimalBase).bits, "9"),

	newStr(newFromPartsRaw(0, -14, 1*decimalBase).bits, "10"),
}

var smallStrings = func() map[uint64]string {
	m := make(map[uint64]string, len(smalls))
	for i := -10; i <= 10; i++ {
		m[smalls[10+i].bits] = strconv.Itoa(i)
	}
	return m
}()

// NewFromInt64 returns a new [Decimal] with the given value.
func NewFromInt64(i int64) Decimal {
	if i >= -10 && i <= 10 {
		return smalls[10+i]
	}
	return newFromInt64(i)
}

func NewFromFloat64(f float64) Decimal {
	// TODO: Find a more mathsy solution.
	return MustParse(strconv.FormatFloat(f, 'g', -1, 64))
}

func newFromInt64(value int64) Decimal {
	sign := int8(0)
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

func renormalize(exp int16, significand uint64) (int16, uint64) {
	numDigits := int16(bits.Len64(significand) * 3 / 10)
	normExp := 15 - numDigits
	if normExp > 0 {
		if normExp > exp+expOffset {
			normExp = exp + expOffset
		}
		exp -= normExp
		significand *= tenToThe[normExp]
	} else if normExp < -1 {
		normExp++
		if normExp < exp-expMax {
			normExp = exp - expMax
		}
		exp -= normExp
		significand /= tenToThe[-normExp]
	}
	switch {
	case significand < decimalBase && exp > -expOffset:
		return exp - 1, significand * 10
	case significand >= 10*decimalBase:
		return exp + 1, significand / 10
	default:
		return exp, significand
	}
}

// roundStatus gives info about the truncated part of the significand that can't be fully stored in 16 decimal digits.
func roundStatus(significand uint64, expDiff int16) discardedDigit {
	if expDiff > 19 && significand != 0 {
		return lt5
	}
	remainder := significand % tenToThe[expDiff]
	midpoint := 5 * tenToThe[expDiff-1]
	if remainder == 0 {
		return eq0
	} else if remainder < midpoint {
		return lt5
	} else if remainder == midpoint {
		return eq5
	}
	return gt5
}

func newFromParts(sign int8, exp int16, significand uint64) Decimal {
	return newDec(newFromPartsRaw(sign, exp, significand).bits)
}

func newFromPartsRaw(sign int8, exp int16, significand uint64) Decimal {
	s := uint64(sign) << 63

	if significand < 0x8<<50 {
		// s EEeeeeeeee   (0)ttt tttttttttt tttttttttt tttttttttt tttttttttt tttttttttt
		//   EE ∈ {00, 01, 10}
		return Decimal{bits: s | uint64(exp+expOffset)<<(63-10) | significand}
	}
	// s 11EEeeeeeeee (100)t tttttttttt tttttttttt tttttttttt tttttttttt tttttttttt
	//     EE ∈ {00, 01, 10}
	significand &= 0x8<<50 - 1
	return Decimal{bits: s | uint64(0xc00|(exp+expOffset))<<(63-12) | significand}
}

func (d Decimal) parts() (fl flavor, sign int8, exp int16, significand uint64) {
	sign = int8(d.bits >> 63)
	switch (d.bits >> (63 - 4)) & 0xf {
	case 15:
		switch (d.bits >> (63 - 6)) & 3 {
		case 0, 1:
			fl = flInf
		case 2:
			fl = flQNaN
			significand = d.bits & (1<<51 - 1)
			return
		case 3:
			fl = flSNaN
			significand = d.bits & (1<<51 - 1)
			return
		}
	case 12, 13, 14:
		// s 11EEeeeeeeee (100)t tttttttttt tttttttttt tttttttttt tttttttttt tttttttttt
		//     EE ∈ {00, 01, 10}
		fl = flNormal51
		exp = int16((d.bits>>(63-12))&(1<<10-1)) - expOffset
		significand = d.bits&(1<<51-1) | (1 << 53)
	default:
		// s EEeeeeeeee   (0)ttt tttttttttt tttttttttt tttttttttt tttttttttt tttttttttt
		//   EE ∈ {00, 01, 10}
		fl = flNormal53
		exp = int16((d.bits>>(63-10))&(1<<10-1)) - expOffset
		significand = d.bits & (1<<53 - 1)
		if significand == 0 {
			exp = 0
		}
	}
	return
}

func expWholeFrac(exp int16, significand uint64) (exp2 int16, whole uint64, frac uint64) {
	if significand == 0 {
		return 0, 0, 0
	}
	if exp >= 0 {
		return exp, significand, 0
	}
	n := uint128T{significand, 0}
	exp += 16
	if exp > 0 {
		n.mul64(&n, tenToThe[exp])
		exp = 0
	} else {
		// exp++ till it hits 0 or continuing would throw away digits.
		for step := 3; step >= 0; step-- {
			expStep := int16(1) << step
			powerOf10 := tenToThe[expStep]
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
	var whole128 uint128T
	whole128.div10base(&n)
	var x uint128T
	x.mul64(&whole128, 10*decimalBase)
	var frac128 uint128T
	frac128.sub(&n, &x)
	return exp, whole128.lo, frac128.lo
}

// Float64 returns a float64 representation of d.
func (d Decimal) Float64() float64 {
	fl, sign, exp, significand := d.parts()
	switch fl {
	case flNormal53, flNormal51:
		if significand == 0 {
			return 0.0 * float64(1-2*sign)
		}
		if exp&1 == 1 {
			exp--
			significand *= 10
		}
		return float64(1-2*sign) * float64(significand) * math.Pow10(int(exp))
	case flInf:
		return math.Inf(1 - 2*int(sign))
	case flQNaN:
		return math.NaN()
	}
	panic(ErrNaN)
}

// Int64 returns an int64 representation of d, clamped to [[math.MinInt64], [math.MaxInt64]].
func (d Decimal) Int64() int64 {
	i, _ := d.Int64x()
	return i
}

// Int64x returns an int64 representation of d, clamped to [[math.MinInt64],
// [math.MaxInt64]].
// The second return value, exact, indicates whether [NewFromInt64](i) == d.
func (d Decimal) Int64x() (i int64, exact bool) {
	fl, sign, exp, significand := d.parts()
	switch fl {
	case flInf:
		if sign == 0 {
			return math.MaxInt64, false
		}
		return math.MinInt64, false
	case flQNaN:
		return 0, false
	case flSNaN:
		panic(ErrNaN)
	}
	exp, whole, frac := expWholeFrac(exp, significand)
	for exp > 0 && whole < math.MaxInt64/10 {
		exp--
		whole *= 10
	}
	if exp > 0 {
		return math.MaxInt64, false
	}
	return int64(1-2*sign) * int64(whole), frac == 0
}

// IsZero returns true if the [Decimal] encodes a zero value.
func (d Decimal) IsZero() bool {
	fl, _, _, significand := d.parts()
	return significand == 0 && fl.normal()
}

// IsInf indicates whether d is ±∞.
func (d Decimal) IsInf() bool {
	return d.flavor() == flInf
}

// IsNaN indicates whether d is not a number.
func (d Decimal) IsNaN() bool {
	return d.flavor().nan()
}

// IsQNaN indicates whether d is a quiet NaN.
func (d Decimal) IsQNaN() bool {
	return d.flavor() == flQNaN
}

// IsSNaN indicates whether d is a signalling NaN.
func (d Decimal) IsSNaN() bool {
	return d.flavor() == flSNaN
}

// IsInt indicates whether d is an integer.
func (d Decimal) IsInt() bool {
	fl, _, exp, significand := d.parts()
	switch fl {
	case flNormal53, flNormal51:
		_, _, frac := expWholeFrac(exp, significand)
		return frac == 0
	default:
		return false
	}
}

// quiet returns a quiet form of d, which must be a NaN.
func (d Decimal) quiet() Decimal {
	return newDec(d.bits &^ (2 << 56))
}

// IsSubnormal indicates whether d is a subnormal.
func (d Decimal) IsSubnormal() bool {
	fl, _, _, significand := d.parts()
	return significand != 0 && significand < decimalBase && fl.normal()
}

// Sign returns -1/0/1 if d is </=/> 0, respectively.
func (d Decimal) Sign() int {
	if d == Zero || d == NegZero {
		return 0
	}
	return 1 - 2*int(d.bits>>63)
}

// Signbit indicates whether d is negative or -0.
func (d Decimal) Signbit() bool {
	return d.bits>>63 == 1
}

func (d Decimal) ScaleB(e Decimal) Decimal {
	var dp, ep decParts
	if nan, is := checkNan(d, e, &dp, &ep); is {
		return nan
	}

	if !dp.fl.normal() || dp.isZero() {
		return d
	}
	if !ep.fl.normal() {
		return QNaN
	}

	i, exact := e.Int64x()
	if !exact {
		return QNaN
	}
	return scaleBInt(d, &dp, int(i))
}

func (d Decimal) ScaleBInt(i int) Decimal {
	var dp decParts
	dp.unpack(d)
	if !dp.fl.normal() || dp.isZero() {
		return d
	}
	return scaleBInt(d, &dp, i)
}

func scaleBInt(d Decimal, dp *decParts, i int) Decimal {
	dp.exp += int16(i)

	for dp.significand.lo < decimalBase && dp.exp > -expOffset {
		dp.exp--
		dp.significand.lo *= 10
	}

	switch {
	case dp.exp > expMax:
		return Inf.CopySign(d)
	case dp.exp < -expOffset:
		for dp.exp < -expOffset {
			dp.exp++
			dp.significand.lo /= 10
		}
		if dp.significand.lo == 0 {
			return Zero.CopySign(d)
		}
	}

	return dp.decimal()
}

// Class returns a string representing the number's 'type' that the decimal is.
// It can be one of the following:
//
//   - "+Normal"
//   - "-Normal"
//   - "+Subnormal"
//   - "-Subnormal"
//   - "+Zero"
//   - "-Zero"
//   - "+Infinity"
//   - "-Infinity"
//   - "NaN"
//   - "sNaN"
func (d Decimal) Class() string {
	var dp decParts
	dp.unpack(d)
	if dp.fl == flSNaN {
		return "sNaN"
	} else if dp.fl.nan() {
		return "NaN"
	}

	switch {
	case dp.fl == flInf:
		return "+Infinity-Infinity"[9*dp.sign : 9*(dp.sign+1)]
	case dp.isZero():
		return "+Zero-Zero"[5*dp.sign : 5*(dp.sign+1)]
	case dp.isSubnormal():
		return "+Subnormal-Subnormal"[10*dp.sign : 10*(dp.sign+1)]
	}
	return "+Normal-Normal"[7*dp.sign : 7*(dp.sign+1)]
}

func checkNan(d, e Decimal, dp, ep *decParts) (Decimal, bool) {
	dp.fl = d.flavor()
	ep.fl = e.flavor()
	switch {
	case dp.fl == flSNaN:
		return d, true
	case ep.fl == flSNaN:
		return e, true
	case dp.fl == flQNaN:
		return d, true
	case ep.fl == flQNaN:
		return e, true
	default:
		dp.unpackV2(d)
		ep.unpackV2(e)
		return Decimal{}, false
	}
}

// checkNan3 returns the decimal NaN that is to be propogated and true else first decimal and false
func checkNan3(d, e, f Decimal, dp, ep, fp *decParts) (Decimal, bool) {
	dp.fl = d.flavor()
	ep.fl = e.flavor()
	fp.fl = f.flavor()
	switch {
	case dp.fl == flSNaN:
		return d, true
	case ep.fl == flSNaN:
		return e, true
	case fp.fl == flSNaN:
		return f, true
	case dp.fl == flQNaN:
		return d, true
	case ep.fl == flQNaN:
		return e, true
	case fp.fl == flQNaN:
		return f, true
	default:
		dp.unpackV2(d)
		ep.unpackV2(e)
		fp.unpackV2(f)
		return Decimal{}, false
	}
}
