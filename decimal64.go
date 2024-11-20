package decimal

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
	flNormal53 flavor = 1 << iota
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

// Context64 may be used to tune the behaviour of arithmetic operations.
type Context64 struct {
	// Rounding sets the rounding behaviour of arithmetic operations.
	Rounding Rounding

	// TODO: implement
	// // Signal causes arithmetic operations to panic when encountering a sNaN.
	// Signal bool
}

var tenToThe = [...]uint64{
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

var ErrNaN64 error = Error("sNaN64")

var small64s = []Decimal64{
	new64str(newFromPartsRaw(1, -14, 1*decimal64Base).bits, "-10"),

	new64str(newFromPartsRaw(1, -15, 9*decimal64Base).bits, "-9"),
	new64str(newFromPartsRaw(1, -15, 8*decimal64Base).bits, "-8"),
	new64str(newFromPartsRaw(1, -15, 7*decimal64Base).bits, "-7"),
	new64str(newFromPartsRaw(1, -15, 6*decimal64Base).bits, "-6"),
	new64str(newFromPartsRaw(1, -15, 5*decimal64Base).bits, "-5"),
	new64str(newFromPartsRaw(1, -15, 4*decimal64Base).bits, "-4"),
	new64str(newFromPartsRaw(1, -15, 3*decimal64Base).bits, "-3"),
	new64str(newFromPartsRaw(1, -15, 2*decimal64Base).bits, "-2"),
	new64str(newFromPartsRaw(1, -15, 1*decimal64Base).bits, "-1"),

	// TODO: Decimal64{}?
	new64str(newFromPartsRaw(0, 0, 0).bits, "0"),

	new64str(newFromPartsRaw(0, -15, 1*decimal64Base).bits, "1"),
	new64str(newFromPartsRaw(0, -15, 2*decimal64Base).bits, "2"),
	new64str(newFromPartsRaw(0, -15, 3*decimal64Base).bits, "3"),
	new64str(newFromPartsRaw(0, -15, 4*decimal64Base).bits, "4"),
	new64str(newFromPartsRaw(0, -15, 5*decimal64Base).bits, "5"),
	new64str(newFromPartsRaw(0, -15, 6*decimal64Base).bits, "6"),
	new64str(newFromPartsRaw(0, -15, 7*decimal64Base).bits, "7"),
	new64str(newFromPartsRaw(0, -15, 8*decimal64Base).bits, "8"),
	new64str(newFromPartsRaw(0, -15, 9*decimal64Base).bits, "9"),

	new64str(newFromPartsRaw(0, -14, 1*decimal64Base).bits, "10"),
}

var small64Strings = func() map[uint64]string {
	m := make(map[uint64]string, len(small64s))
	for i := -10; i <= 10; i++ {
		m[small64s[10+i].bits] = strconv.Itoa(i)
	}
	return m
}()

// New64FromInt64 returns a new Decimal64 with the given value.
func New64FromInt64(i int64) Decimal64 {
	if i >= -10 && i <= 10 {
		return small64s[10+i]
	}
	return new64FromInt64(i)
}

func new64FromInt64(value int64) Decimal64 {
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
	numBits := bits.Len64(significand)
	numDigits := numBits * 3 / 10
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
	case significand < decimal64Base && exp > -expOffset:
		return exp - 1, significand * 10
	case significand >= 10*decimal64Base:
		return exp + 1, significand / 10
	default:
		return exp, significand
	}
}

// roundStatus gives info about the truncated part of the significand that can't be fully stored in 16 decimal digits.
func roundStatus(significand uint64, exp int, targetExp int) discardedDigit {
	expDiff := targetExp - exp
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

func newFromParts(sign int, exp int, significand uint64) Decimal64 {
	return new64(newFromPartsRaw(sign, exp, significand).bits)
}

func newFromPartsRaw(sign int, exp int, significand uint64) Decimal64 {
	s := uint64(sign) << 63

	if significand < 0x8<<50 {
		// s EEeeeeeeee   (0)ttt tttttttttt tttttttttt tttttttttt tttttttttt tttttttttt
		//   EE ∈ {00, 01, 10}
		return Decimal64{bits: s | uint64(exp+expOffset)<<(63-10) | significand}
	}
	// s 11EEeeeeeeee (100)t tttttttttt tttttttttt tttttttttt tttttttttt tttttttttt
	//     EE ∈ {00, 01, 10}
	significand &= 0x8<<50 - 1
	return Decimal64{bits: s | uint64(0xc00|(exp+expOffset))<<(63-12) | significand}
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
		n = n.mul64(tenToThe[exp])
		exp = 0
	} else {
		// exp++ till it hits 0 or continuing would throw away digits.
		for step := 3; step >= 0; step-- {
			expStep := 1 << uint(step)
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
	whole128 := n.div64(10 * decimal64Base)
	frac128 := n.sub(whole128.mul64(10 * decimal64Base))
	return exp, whole128.lo, frac128.lo
}

// Float64 returns a float64 representation of d.
func (d Decimal64) Float64() float64 {
	fl, sign, exp, significand := d.parts()
	switch fl {
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
	panic(ErrNaN64)
}

// Int64 returns an int64 representation of d, clamped to [[math.MinInt64], [math.MaxInt64]].
func (d Decimal64) Int64() int64 {
	i, _ := d.Int64x()
	return i
}

// Int64 returns an int64 representation of d, clamped to [[math.MinInt64],
// [math.MaxInt64]].
// The second return value, exact indicates whether New64FromInt64(i) == d.
func (d Decimal64) Int64x() (i int64, exact bool) {
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
		panic(ErrNaN64)
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

// IsZero returns true if the Decimal encodes a zero value.
func (d Decimal64) IsZero() bool {
	fl, _, _, significand := d.parts()
	return significand == 0 && fl.normal()
}

// IsInf indicates whether d is ±∞.
func (d Decimal64) IsInf() bool {
	return d.flavor() == flInf
}

// IsNaN indicates whether d is not a number.
func (d Decimal64) IsNaN() bool {
	return d.flavor().nan()
}

// IsQNaN indicates whether d is a quiet NaN.
func (d Decimal64) IsQNaN() bool {
	return d.flavor() == flQNaN
}

// IsSNaN indicates whether d is a signalling NaN.
func (d Decimal64) IsSNaN() bool {
	return d.flavor() == flSNaN
}

// IsInt indicates whether d is an integer.
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

// quiet returns a quiet form of d, which must be a NaN.
func (d Decimal64) quiet() Decimal64 {
	return new64(d.bits &^ (2 << 56))
}

// IsSubnormal indicates whether d is a subnormal.
func (d Decimal64) IsSubnormal() bool {
	fl, _, _, significand := d.parts()
	return significand != 0 && significand < decimal64Base && fl.normal()
}

// Sign returns -1/0/1 if d is </=/> 0, respectively.
func (d Decimal64) Sign() int {
	if d == Zero64 || d == NegZero64 {
		return 0
	}
	return 1 - 2*int(d.bits>>63)
}

// Signbit indicates whether d is negative or -0.
func (d Decimal64) Signbit() bool {
	return d.bits>>63 == 1
}

func (d Decimal64) ScaleB(e Decimal64) Decimal64 {
	var dp decParts
	dp.unpack(d)
	var ep decParts
	ep.unpack(e)
	if r, nan := checkNan(&dp, &ep); nan {
		return r
	}

	if !dp.fl.normal() || dp.isZero() {
		return d
	}
	if !ep.fl.normal() {
		return QNaN64
	}

	i, exact := e.Int64x()
	if !exact {
		return QNaN64
	}
	return scaleBInt(&dp, int(i))
}

func (d Decimal64) ScaleBInt(i int) Decimal64 {
	var dp decParts
	dp.unpack(d)
	if !dp.fl.normal() || dp.isZero() {
		return d
	}
	return scaleBInt(&dp, i)
}

func scaleBInt(dp *decParts, i int) Decimal64 {
	dp.exp += i

	for dp.significand.lo < decimal64Base && dp.exp > -expOffset {
		dp.exp--
		dp.significand.lo *= 10
	}

	switch {
	case dp.exp > expMax:
		return Infinity64.CopySign(dp.original)
	case dp.exp < -expOffset:
		for dp.exp < -expOffset {
			dp.exp++
			dp.significand.lo /= 10
		}
		if dp.significand.lo == 0 {
			return Zero64.CopySign(dp.original)
		}
	}

	return dp.decimal64()
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
func (d Decimal64) Class() string {
	var dp decParts
	dp.unpack(d)
	if dp.isSNaN() {
		return "sNaN"
	} else if dp.isNaN() {
		return "NaN"
	}

	switch {
	case dp.isInf():
		return "+Infinity-Infinity"[9*dp.sign : 9*(dp.sign+1)]
	case dp.isZero():
		return "+Zero-Zero"[5*dp.sign : 5*(dp.sign+1)]
	case dp.isSubnormal():
		return "+Subnormal-Subnormal"[10*dp.sign : 10*(dp.sign+1)]
	}
	return "+Normal-Normal"[7*dp.sign : 7*(dp.sign+1)]
}

// numDecimalDigits returns the magnitude (number of digits) of a uint64.
func numDecimalDigits(n uint64) int {
	numBits := 64 - bits.LeadingZeros64(n)
	numDigits := numBits * 3 / 10
	if n < tenToThe[numDigits] {
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

// checkNan returns the decimal NaN that is to be propogated and true else first decimal and false
func checkNanV2(fld, fle flavor, d, e Decimal64) (Decimal64, bool) {
	if fld == flSNaN {
		return d, true
	}
	if fle == flSNaN {
		return e, true
	}
	if fld == flQNaN {
		return d, true
	}
	if fle == flQNaN {
		return e, true
	}
	return d, false
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
