package decimal

import "math/bits"

// Equal indicates whether two numbers are equal.
// It is equivalent to d.Cmp(e) == 0.
func (d Decimal64) Equal(e Decimal64) bool {
	return d.Cmp(e) == 0
}

// Abs computes ||d||.
func (d Decimal64) Abs() Decimal64 {
	if d.flavor().nan() {
		return d
	}
	return new64(^neg64 & uint64(d.bits))
}

// Add computes d + e.
// It uses [DefaultContext64] to call [Context64.Add].
func (d Decimal64) Add(e Decimal64) Decimal64 {
	return DefaultContext64.Add(d, e)
}

// FMA computes d × e + f.
// It uses [DefaultContext64] to call [Context64.FMA].
func (d Decimal64) FMA(e, f Decimal64) Decimal64 {
	return DefaultContext64.FMA(d, e, f)
}

// Mul computes d × e.
// It uses [DefaultContext64] to call [Context64.Mul].
func (d Decimal64) Mul(e Decimal64) Decimal64 {
	return DefaultContext64.Mul(d, e)
}

// Sub returns d - e.
// It uses [DefaultContext64] to call [Context64.Sub].
func (d Decimal64) Sub(e Decimal64) Decimal64 {
	return DefaultContext64.Sub(d, e)
}

// Quo computes d ÷ e.
// It uses [DefaultContext64] to call [Context64.Quo].
func (d Decimal64) Quo(e Decimal64) Decimal64 {
	return DefaultContext64.Quo(d, e)
}

// Cmp returns:
//
//	-2 if d or e is NaN
//	-1 if d <  e
//	 0 if d == e (incl. -0 == 0, -Inf == -Inf, and +Inf == +Inf)
//	+1 if d >  e
func (d Decimal64) Cmp(e Decimal64) int {
	var dp, ep decParts
	if checkNan(d, e, &dp, &ep) != nil {
		return -2
	}
	return cmp(&dp, &ep)
}

// Cmp64 returns the same output as Cmp as a Decimal64, unless d or e is NaN, in
// which case it returns a corresponding NaN result.
func (d Decimal64) Cmp64(e Decimal64) Decimal64 {
	var dp, ep decParts
	if nan := checkNan(d, e, &dp, &ep); nan != nil {
		return *nan
	}
	switch cmp(&dp, &ep) {
	case -1:
		return NegOne64
	case 1:
		return One64
	default:
		return Zero64
	}
}

func cmp(dp, ep *decParts) int {
	switch {
	case dp.isZero() && ep.isZero(), dp.original == ep.original:
		return 0
	default:
		diff := dp.original.Sub(ep.original)
		return 1 - 2*int(diff.bits>>63)
	}
}

// Min returns the lower of d and e.
func (d Decimal64) Min(e Decimal64) Decimal64 {
	return d.min(e, 1)
}

// Max returns the lower of d and e.
func (d Decimal64) Max(e Decimal64) Decimal64 {
	return d.min(e, -1)
}

// Min returns the lower of d and e.
func (d Decimal64) min(e Decimal64, sign int) Decimal64 {
	var dp, ep decParts
	dp.unpack(d)
	ep.unpack(e)

	dnan := dp.fl.nan()
	enan := ep.fl.nan()

	switch {
	case !dnan && !enan: // Fast path for non-NaNs.
		if sign*cmp(&dp, &ep) < 0 {
			return d
		}
		return e

	case dp.fl == flSNaN:
		return d.quiet()
	case ep.fl == flSNaN:
		return e.quiet()

	case !enan:
		return e
	default:
		return d
	}
}

// MinMag returns the lower of d and e.
func (d Decimal64) MinMag(e Decimal64) Decimal64 {
	return d.minMag(e, 1)
}

// MaxMag returns the lower of d and e.
func (d Decimal64) MaxMag(e Decimal64) Decimal64 {
	return d.minMag(e, -1)
}

// MinMag returns the lower of d and e.
func (d Decimal64) minMag(e Decimal64, sign int) Decimal64 {
	var dp decParts
	dp.unpack(d.Abs())
	var ep decParts
	ep.unpack(e.Abs())

	dnan := dp.fl.nan()
	enan := ep.fl.nan()

	switch {
	case !dnan && !enan: // Fast path for non-NaNs.
		switch sign * cmp(&dp, &ep) {
		case -1:
			return d
		case 1:
			return e
		default:
			if 2*int(d.bits>>63) == 1+sign {
				return d
			}
			return e
		}
	case dp.fl == flSNaN:
		return d.quiet()
	case ep.fl == flSNaN:
		return e.quiet()
	case !enan:
		return e
	default:
		return d
	}
}

// Neg computes -d.
func (d Decimal64) Neg() Decimal64 {
	if d.flavor().nan() {
		return d
	}
	return new64(neg64 ^ d.bits)
}

// Logb return the integral log10 of d.
func (d Decimal64) Logb() Decimal64 {
	fl := d.flavor()
	switch {
	case fl.nan():
		return d
	case d.IsZero():
		return NegInfinity64
	case fl == flInf:
		return Infinity64
	default:
		var dp decParts
		dp.unpack(d)

		// Adjust for subnormals.
		e := dp.exp
		for s := dp.significand.lo; s < decimal64Base; s *= 10 {
			e--
		}

		return New64FromInt64(int64(15 + e))
	}
}

// CopySign copies d, but with the sign taken from e.
func (d Decimal64) CopySign(e Decimal64) Decimal64 {
	return new64(d.bits&^neg64 | e.bits&neg64)
}

// Quo computes d / e.
// Rounding rules are applied as per the context.
func (ctx Context64) Quo(d, e Decimal64) Decimal64 {
	var dp, ep decParts
	if nan := checkNan(d, e, &dp, &ep); nan != nil {
		return *nan
	}
	var ans decParts
	ans.sign = dp.sign ^ ep.sign
	if dp.isZero() {
		if ep.isZero() {
			return QNaN64
		}
		return zeroes64[ans.sign]
	}
	if dp.isinf() {
		if ep.isinf() {
			return QNaN64
		}
		return infinities64[ans.sign]
	}
	if ep.isinf() {
		return zeroes64[ans.sign]
	}
	if ep.isZero() {
		return infinities64[ans.sign]
	}

	const ampl = 1000
	hi, lo := bits.Mul64(dp.significand.lo, ampl*decimal64Base)
	q, _ := bits.Div64(hi, lo, ep.significand.lo)
	exp := dp.exp - ep.exp - 15

	for q < ampl*decimal64Base && exp > -expOffset {
		q *= 10
		exp--
	}
	for q >= 10*ampl*decimal64Base {
		q /= 10
		exp++
	}
	for exp < -expOffset {
		q /= 10
		exp++
	}
	if exp > expMax {
		return infinities64[ans.sign]
	}

	switch ctx.Rounding {
	case HalfUp:
		q = (q + ampl/2) / ampl
	case HalfEven:
		d := q / ampl
		rem := q - d*ampl
		q = d
		if rem > ampl/2 || rem == ampl/2 && d%2 == 1 {
			q++
		}
	case Down:
		q /= ampl
	}

	return newFromParts(ans.sign, exp, q)
}

// Sqrt computes √d.
func (d Decimal64) Sqrt() Decimal64 {
	flav, sign, exp, significand := d.parts()
	switch flav {
	case flInf:
		if sign == 1 {
			return QNaN64
		}
		return d
	case flQNaN:
		return d
	case flSNaN:
		return SNaN64
	case flNormal53, flNormal51:
	}
	if significand == 0 {
		return d
	}
	if sign == 1 {
		return QNaN64
	}
	if exp&1 == 1 {
		exp++
		significand *= 10
	} else {
		significand *= 100
	}
	s := sqrtu64(significand) / 10
	exp, significand = renormalize(exp/2, s)
	return newFromParts(sign, exp, significand)
}

// Add computes d + e
func (ctx Context64) Add(d, e Decimal64) Decimal64 {
	var dp, ep decParts
	if nan := checkNan(d, e, &dp, &ep); nan != nil {
		return *nan
	}
	if dp.fl == flInf || ep.fl == flInf {
		if dp.fl != flInf {
			return e
		}
		if ep.fl != flInf || ep.sign == dp.sign {
			return d
		}
		return QNaN64
	}
	if dp.significand.lo == 0 {
		return e
	} else if ep.significand.lo == 0 {
		return d
	}
	sep := dp.separation(&ep)

	if sep < 0 {
		dp, ep = ep, dp
		sep = -sep
	}
	if sep > 17 {
		return dp.original
	}
	var rndStatus discardedDigit
	var ans decParts
	ans.add128(&dp, &ep)
	rndStatus = ans.roundToLo()
	if ans.exp < -expOffset {
		rndStatus = ans.rescale(-expOffset)
	}
	ans.significand.lo = ctx.Rounding.round(ans.significand.lo, rndStatus)
	if ans.exp >= -expOffset && ans.significand.lo != 0 {
		ans.exp, ans.significand.lo = renormalize(ans.exp, ans.significand.lo)
	}
	if ans.exp > expMax || ans.significand.lo > maxSig {
		return infinities64[ans.sign]
	}
	return ans.decimal64()
}

// Add computes d + e
func (ctx Context64) Sub(d, e Decimal64) Decimal64 {
	return d.Add(e.Neg())
}

// FMA computes d*e + f
func (ctx Context64) FMA(d, e, f Decimal64) Decimal64 {
	var dp, ep decParts
	fp := decParts{original: f}
	if nan := checkNan3(d, e, f, &dp, &ep, &fp); nan != nil {
		return *nan
	}
	var ans decParts
	ans.sign = dp.sign ^ ep.sign
	if dp.fl == flInf || ep.fl == flInf {
		if fp.fl == flInf && ans.sign != fp.sign {
			return QNaN64
		}
		if ep.isZero() || dp.isZero() {
			return QNaN64
		}
		return infinities64[ans.sign]
	}
	if ep.significand.lo == 0 || dp.significand.lo == 0 {
		return f
	}
	if fp.fl == flInf {
		return infinities64[fp.sign]
	}

	var rndStatus discardedDigit
	ep.removeZeros()
	dp.removeZeros()
	ans.exp = dp.exp + ep.exp
	ans.significand.umul64(dp.significand.lo, ep.significand.lo)
	sep := ans.separation(&fp)
	if fp.significand.lo != 0 {
		if sep < -17 {
			return f
		} else if sep <= 17 {
			ans.add128(&ans, &fp)
		}
	}
	rndStatus = ans.roundToLo()
	if ans.exp < -expOffset {
		rndStatus = ans.rescale(-expOffset)
	}
	ans.significand.lo = ctx.Rounding.round(ans.significand.lo, rndStatus)
	if ans.exp >= -expOffset && ans.significand.lo != 0 {
		ans.exp, ans.significand.lo = renormalize(ans.exp, ans.significand.lo)
	}
	if ans.exp > expMax || ans.significand.lo > maxSig {
		return infinities64[ans.sign]
	}
	return ans.decimal64()
}

// Mul computes d * e
func (ctx Context64) Mul(d, e Decimal64) Decimal64 {
	// fld := flav(d)
	// fle := flav(e)
	var dp, ep decParts
	if nan := checkNan(d, e, &dp, &ep); nan != nil {
		return *nan
	}
	var ans decParts
	ans.sign = dp.sign ^ ep.sign
	if dp.fl == flInf || ep.fl == flInf {
		if ep.isZero() || dp.isZero() {
			return QNaN64
		}
		return infinities64[ans.sign]
	}
	if ep.significand.lo == 0 || dp.significand.lo == 0 {
		return zeroes64[ans.sign]
	}
	var roundStatus discardedDigit
	ans.significand.umul64(dp.significand.lo, ep.significand.lo)
	ans.exp = dp.exp + ep.exp + 15
	ans.significand.divbase(&ans.significand)
	if ans.exp >= -expOffset {
		ans.exp, ans.significand.lo = renormalize(ans.exp, ans.significand.lo)
	} else if ans.exp < 1-expMax {
		roundStatus = ans.rescale(-expOffset)
	}
	ans.significand.lo = ctx.Rounding.round(ans.significand.lo, roundStatus)
	if ans.significand.lo > maxSig || ans.exp > expMax {
		return infinities64[ans.sign]
	}
	return ans.decimal64()
}

// NextPlus returns the next value above d.
func (d Decimal64) NextPlus() Decimal64 {
	flav, sign, exp, significand := d.parts()
	switch {
	case flav == flInf:
		if sign == 1 {
			return NegMax64
		}
		return Infinity64
	case !flav.normal():
		return d
	case significand == 0:
		return Min64
	case sign == 1:
		switch {
		case significand > decimal64Base:
			return new64(d.bits - 1)
		case exp == -398:
			if significand > 1 {
				return new64(d.bits - 1)
			}
			return Zero64
		default:
			return newFromParts(sign, exp-1, 10*decimal64Base-1)
		}
	default:
		switch {
		case significand < 10*decimal64Base-1:
			return new64(d.bits + 1)
		case exp == 369:
			return Infinity64
		default:
			return newFromParts(sign, exp+1, decimal64Base)
		}
	}
}

// NextMinus returns the next value above d.
func (d Decimal64) NextMinus() Decimal64 {
	flav, sign, exp, significand := d.parts()
	switch {
	case flav == flInf:
		if sign == 0 {
			return Max64
		}
		return NegInfinity64
	case !flav.normal():
		return d
	case significand == 0:
		return NegMin64
	case sign == 0:
		switch {
		case significand > decimal64Base:
			return new64(d.bits - 1)
		case exp == -398:
			if significand > 1 {
				return new64(d.bits - 1)
			}
			return Zero64
		default:
			return newFromParts(sign, exp-1, 10*decimal64Base-1)
		}
	default:
		switch {
		case significand < 10*decimal64Base-1:
			return new64(d.bits + 1)
		case exp == 369:
			return NegInfinity64
		default:
			return newFromParts(sign, exp+1, decimal64Base)
		}
	}
}

// Round rounds a number to a given power-of-10 value.
// The e argument should be a power of ten, such as 1, 10, 100, 1000, etc.
// It uses [DefaultContext64] to call [Context64.Round].
func (d Decimal64) Round(e Decimal64) Decimal64 {
	return DefaultContext64.Round(d, e)
}

// Round rounds a number to a given power of ten value.
// The e argument should be a power of ten, such as 1, 10, 100, 1000, etc.
func (ctx Context64) Round(d, e Decimal64) Decimal64 {
	return new64(ctx.roundRaw(d, e).bits)
}

func (ctx Context64) roundRaw(d, e Decimal64) Decimal64 {
	var dp, ep decParts
	return ctx.roundRefRaw(d, e, &dp, &ep)
}

var (
	zero64Raw = new64nostr(newFromPartsRaw(0, 0, 0).bits)
	qNaN64Raw = new64nostr(0x7c << 56)
)

func (ctx Context64) roundRefRaw(d, e Decimal64, dp, ep *decParts) Decimal64 {
	if nan := checkNan(d, e, dp, ep); nan != nil {
		return *nan
	}
	if dp.fl == flInf || ep.fl == flInf {
		if dp.fl == flInf && ep.fl == flInf {
			return d
		}
		return qNaN64Raw
	}

	dexp, dsignificand := unsubnormal(dp.exp, dp.significand.lo)
	eexp, _ := unsubnormal(ep.exp, ep.significand.lo)

	delta := dexp - eexp
	if delta < -1 { // -1 avoids rounding range
		return zero64Raw
	}
	if delta > 14 {
		return d
	}
	p := tenToThe[14-delta]
	s, grew := ctx.round(dsignificand, p)
	exp := dexp
	if grew {
		s /= 10
		exp++ // Cannot max out because final digit never rounds up.
	}
	exp, s = resubnormal(exp, s)
	return newFromPartsRaw(dp.sign, exp, s)
}

// ToIntegral rounds d to a nearby integer.
// It uses [DefaultContext64] to call [Context64.ToIntegral].
func (d Decimal64) ToIntegral() Decimal64 {
	return DefaultContext64.ToIntegral(d)
}

var decPartsOne64 decParts = unpack(One64)

// ToIntegral rounds d to a nearby integer.
func (ctx Context64) ToIntegral(d Decimal64) Decimal64 {
	var dp decParts
	dp.unpack(d)
	if !dp.fl.normal() || dp.exp >= 0 {
		return d
	}
	return new64(ctx.roundRefRaw(d, One64, &dp, &decPartsOne64).bits)
}

func (ctx Context64) round(s, p uint64) (uint64, bool) {
	p5 := p * 5
	p10 := p5 * 2
	div := s / p10
	rem := s - p10*div
	if rem == 0 {
		return s, false
	}
	s -= rem
	up := false
	switch ctx.Rounding {
	case HalfUp:
		up = rem >= p5
	case HalfEven:
		up = rem > p5 || rem == p5 && div%2 == 1
	}
	if up {
		return s + p10, div == 0
	}
	return s, false
}

func unsubnormal(exp int16, significand uint64) (int16, uint64) {
	if significand != 0 {
		for significand < decimal64Base {
			significand *= 10
			exp--
		}
	}
	return exp, significand
}

func resubnormal(exp int16, significand uint64) (int16, uint64) {
	for exp < -expOffset || significand >= 10*decimal64Base {
		significand /= 10
		exp++
	}
	return exp, significand
}
