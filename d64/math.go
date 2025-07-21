package d64

import "math/bits"

// Equal indicates whether two numbers are equal.
// It is equivalent to d.Cmp(e) == 0.
func (d Decimal) Equal(e Decimal) bool {
	return d.Cmp(e) == 0
}

// Abs computes ||d||.
func (d Decimal) Abs() Decimal {
	if d.flavor().nan() {
		return d
	}
	return d.abs()
}

func (d Decimal) abs() Decimal {
	return newDec(^neg & uint64(d.bits))
}

// Add computes d + e.
// It uses [DefaultContext] to call [Context.Add].
func (d Decimal) Add(e Decimal) Decimal {
	return DefaultContext.Add(d, e)
}

// FMA computes d × e + f.
// It uses [DefaultContext] to call [Context.FMA].
func (d Decimal) FMA(e, f Decimal) Decimal {
	return DefaultContext.FMA(d, e, f)
}

// Mul computes d × e.
// It uses [DefaultContext] to call [Context.Mul].
func (d Decimal) Mul(e Decimal) Decimal {
	return DefaultContext.Mul(d, e)
}

// Sub returns d - e.
// It uses [DefaultContext] to call [Context.Sub].
func (d Decimal) Sub(e Decimal) Decimal {
	return DefaultContext.Sub(d, e)
}

// Quo computes d ÷ e.
// It uses [DefaultContext] to call [Context.Quo].
func (d Decimal) Quo(e Decimal) Decimal {
	return DefaultContext.Quo(d, e)
}

// Cmp returns:
//
//	-2 if d or e is NaN
//	-1 if d <  e
//	 0 if d == e (incl. -0 == 0, -Inf == -Inf, and +Inf == +Inf)
//	+1 if d >  e
func (d Decimal) Cmp(e Decimal) int {
	var dp, ep decParts
	if _, nan := checkNan2(d, e, &dp, &ep); nan {
		return -2
	}
	return cmp(d, e, &dp, &ep)
}

// CmpDec is equivalent to Cmp but with a [Decimal] result.
// If d or e is NaN, it returns a corresponding NaN result.
func (d Decimal) CmpDec(e Decimal) Decimal {
	var dp, ep decParts
	if nan, is := checkNan2(d, e, &dp, &ep); is {
		return nan
	}
	switch cmp(d, e, &dp, &ep) {
	case -1:
		return NegOne
	case 1:
		return One
	default:
		return Zero
	}
}

func cmp(d, e Decimal, dp, ep *decParts) int {
	switch {
	case d == e, dp.isZero() && ep.isZero():
		return 0
	default:
		diff := d.Sub(e)
		return 1 - 2*int(diff.bits>>63)
	}
}

// Min returns the lower of d and e.
func (d Decimal) Min(e Decimal) Decimal {
	return d.min(e, 1)
}

// Max returns the lower of d and e.
func (d Decimal) Max(e Decimal) Decimal {
	return d.min(e, -1)
}

// Min returns the lower of d and e.
func (d Decimal) min(e Decimal, sign int) Decimal {
	var dp, ep decParts
	dp.unpack(d)
	ep.unpack(e)

	dnan := dp.fl.nan()
	enan := ep.fl.nan()

	switch {
	case !dnan && !enan: // Fast path for non-NaNs.
		if sign*cmp(d, e, &dp, &ep) < 0 {
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
func (d Decimal) MinMag(e Decimal) Decimal {
	return d.minMag(e, 1)
}

// MaxMag returns the lower of d and e.
func (d Decimal) MaxMag(e Decimal) Decimal {
	return d.minMag(e, -1)
}

// MinMag returns the lower of d and e.
func (d Decimal) minMag(e Decimal, sign int) Decimal {
	da, ea := d.abs(), e.abs()
	var dp decParts
	dp.unpack(da)
	var ep decParts
	ep.unpack(ea)

	dnan := dp.fl.nan()
	enan := ep.fl.nan()

	switch {
	case !dnan && !enan: // Fast path for non-NaNs.
		switch sign * cmp(da, ea, &dp, &ep) {
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
func (d Decimal) Neg() Decimal {
	if d.flavor().nan() {
		return d
	}
	return newDec(neg ^ d.bits)
}

// Logb return the integral log10 of d.
func (d Decimal) Logb() Decimal {
	fl := d.flavor()
	switch {
	case fl.nan():
		return d
	case d.IsZero():
		return NegInf
	case fl == flInf:
		return Inf
	default:
		var dp decParts
		dp.unpack(d)

		// Adjust for subnormals.
		e := dp.exp
		for s := dp.significand.lo; s < decimalBase; s *= 10 {
			e--
		}

		return NewFromInt64(int64(15 + e))
	}
}

// CopySign copies d, but with the sign taken from e.
func (d Decimal) CopySign(e Decimal) Decimal {
	return newDec(d.bits&^neg | e.bits&neg)
}

// Quo computes d / e.
// Rounding rules are applied as per the context.
func (ctx Context) Quo(d, e Decimal) Decimal {
	var dp, ep decParts
	if nan, is := checkNan2(d, e, &dp, &ep); is {
		return nan
	}
	var ans decParts
	ans.sign = dp.sign ^ ep.sign
	if dp.isZero() {
		if ep.isZero() {
			return QNaN
		}
		return zeroes[ans.sign]
	}
	if dp.isinf() {
		if ep.isinf() {
			return QNaN
		}
		return infinities[ans.sign]
	}
	if ep.isinf() {
		return zeroes[ans.sign]
	}
	if ep.isZero() {
		return infinities[ans.sign]
	}

	const ampl = 1000
	hi, lo := bits.Mul64(dp.significand.lo, ampl*decimalBase)
	q, _ := bits.Div64(hi, lo, ep.significand.lo)
	exp := dp.exp - ep.exp - 15

	for q < ampl*decimalBase && exp > -expOffset {
		q *= 10
		exp--
	}
	for q >= 10*ampl*decimalBase {
		q /= 10
		exp++
	}
	for exp < -expOffset {
		q /= 10
		exp++
	}
	if exp > expMax {
		return infinities[ans.sign]
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
func (d Decimal) Sqrt() Decimal {
	flav, sign, exp, significand := d.parts()
	switch flav {
	case flInf:
		if sign == 1 {
			return QNaN
		}
		return d
	case flQNaN:
		return d
	case flSNaN:
		return SNaN
	case flNormal53, flNormal51:
	}
	if significand == 0 {
		return d
	}
	if sign == 1 {
		return QNaN
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
func (ctx Context) Add(d, e Decimal) Decimal {
	var dp, ep decParts
	if checkFinite2(d, e, &dp, &ep) {
		return ctx.add(d, e, &dp, &ep)
	}
	if nan, is := checkNan2(d, e, &dp, &ep); is {
		return nan
	}
	if dp.fl != flInf {
		return e.noSigInf()
	}
	if ep.fl != flInf || ep.sign == dp.sign {
		return d.noSigInf()
	}
	return QNaN
}

// noSigInf returns the same inf but with all ignored bits set to zero.
func (d Decimal) noSigInf() Decimal {
	return newDec(d.bits &^ ((1 << (64 - 1 - 5)) - 1))
}

// Add computes d + e
func (ctx Context) add(d, e Decimal, dp, ep *decParts) Decimal {
	if dp.exp == ep.exp && dp.sign == ep.sign {
		if sig := dp.significand.lo + ep.significand.lo; sig < 10*decimalBase {
			dp.significand.lo = sig
			return dp.decimal()
		}
	}
	if dp.significand.lo == 0 {
		return e
	} else if ep.significand.lo == 0 {
		return d
	}

	var ans decParts

	sep := dp.exp - ep.exp
	switch {
	case sep < -17:
		ans = *ep
	case sep > 17:
		ans = *dp
	default:
		if sep < 0 {
			dp, ep = ep, dp
			sep = -sep
		}
		var rndStatus discardedDigit
		switch {
		case sep == 0:
			ans.add64(dp, ep)
		case sep < 4:
			dp.significand.lo *= tenToThe[sep]
			dp.exp -= sep
			ans.add64(dp, ep)
		default:
			dp.significand.mul64(&dp.significand, tenToThe[17])
			dp.exp -= 17
			ep.significand.mul64(&ep.significand, tenToThe[17-sep])
			ep.exp -= 17 - sep
			ans.add128V2(dp, ep)
		}
		rndStatus = ans.roundToLo()
		if ans.exp < -expOffset {
			rndStatus = ans.rescale(-expOffset)
		}
		ans.significand.lo = ctx.Rounding.round(ans.significand.lo, rndStatus)
	}

	prefexp := min(dp.exp, ep.exp)
	// TODO: replace O(n) loops with O(1) or O(log n) rescaling.
	for ans.exp < prefexp && ans.significand.lo%10 == 0 {
		ans.significand.lo /= 10
		ans.exp++
	}
	for ans.exp > prefexp && ans.significand.lo < decimalBase {
		ans.significand.lo *= 10
		ans.exp--
	}
	// if ans.significand.lo != 0 {
	// 	ans.exp, ans.significand.lo = renormalize(ans.exp, ans.significand.lo)
	// }
	if ans.exp > expMax || ans.significand.lo > maxSig {
		return infinities[ans.sign]
	}
	return ans.decimal()
}

// Add computes d + e
func (ctx Context) Sub(d, e Decimal) Decimal {
	return d.Add(newDec(neg ^ e.bits))
}

// FMA computes d*e + f
func (ctx Context) FMA(d, e, f Decimal) Decimal {
	var dp, ep, fp decParts
	if nan, is := checkNan3(d, e, f, &dp, &ep, &fp); is {
		return nan
	}
	var ans decParts
	ans.sign = dp.sign ^ ep.sign
	if dp.fl == flInf || ep.fl == flInf {
		if fp.fl == flInf && ans.sign != fp.sign {
			return QNaN
		}
		if ep.isZero() || dp.isZero() {
			return QNaN
		}
		return infinities[ans.sign]
	}
	if ep.significand.lo == 0 || dp.significand.lo == 0 {
		return f
	}
	if fp.fl == flInf {
		return infinities[fp.sign]
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
		return infinities[ans.sign]
	}
	return ans.decimal()
}

// Mul computes d * e.
func (ctx Context) Mul(d, e Decimal) Decimal {
	var dp, ep decParts
	if nan, is := checkNan2(d, e, &dp, &ep); is {
		return nan
	}
	var ans decParts
	ans.sign = dp.sign ^ ep.sign
	if dp.fl == flInf || ep.fl == flInf {
		if dp.isZero() || ep.isZero() {
			return QNaN
		}
		return infinities[ans.sign]
	}
	return ctx.mul(&dp, &ep, &ans)
}

func (ctx Context) mul(dp, ep, ans *decParts) Decimal {
	if ep.significand.lo == 0 || dp.significand.lo == 0 {
		return zeroes[ans.sign]
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
		return infinities[ans.sign]
	}
	return ans.decimal()
}

// NextPlus returns the next value above d.
func (d Decimal) NextPlus() Decimal {
	flav, sign, exp, significand := d.parts()
	switch {
	case flav == flInf:
		if sign == 1 {
			return NegMax
		}
		return Inf
	case !flav.normal():
		return d
	case significand == 0:
		return Min
	case sign == 1:
		switch {
		case significand > decimalBase:
			return newDec(d.bits - 1)
		case exp == -398:
			if significand > 1 {
				return newDec(d.bits - 1)
			}
			return Zero
		default:
			return newFromParts(sign, exp-1, 10*decimalBase-1)
		}
	default:
		switch {
		case significand < 10*decimalBase-1:
			return newDec(d.bits + 1)
		case exp == 369:
			return Inf
		default:
			return newFromParts(sign, exp+1, decimalBase)
		}
	}
}

// NextMinus returns the next value above d.
func (d Decimal) NextMinus() Decimal {
	flav, sign, exp, significand := d.parts()
	switch {
	case flav == flInf:
		if sign == 0 {
			return Max
		}
		return NegInf
	case !flav.normal():
		return d
	case significand == 0:
		return NegMin
	case sign == 0:
		switch {
		case significand > decimalBase:
			return newDec(d.bits - 1)
		case exp == -398:
			if significand > 1 {
				return newDec(d.bits - 1)
			}
			return Zero
		default:
			return newFromParts(sign, exp-1, 10*decimalBase-1)
		}
	default:
		switch {
		case significand < 10*decimalBase-1:
			return newDec(d.bits + 1)
		case exp == 369:
			return NegInf
		default:
			return newFromParts(sign, exp+1, decimalBase)
		}
	}
}

// Round rounds a number to a given power-of-10 value.
// The e argument should be a power of ten, such as 1, 10, 100, 1000, etc.
// It uses [DefaultContext] to call [Context.Round].
func (d Decimal) Round(e Decimal) Decimal {
	return DefaultContext.Round(d, e)
}

// Round rounds a number to a given power of ten value.
// The e argument should be a power of ten, such as 1, 10, 100, 1000, etc.
func (ctx Context) Round(d, e Decimal) Decimal {
	return newDec(ctx.roundRaw(d, e).bits)
}

func (ctx Context) roundRaw(d, e Decimal) Decimal {
	var dp, ep decParts
	return ctx.roundRefRaw(d, e, &dp, &ep)
}

var (
	zeroRaw = newNostr(newFromPartsRaw(0, 0, 0).bits)
	qNaNRaw = newNostr(0x7c << 56)
)

func (ctx Context) roundRefRaw(d, e Decimal, dp, ep *decParts) Decimal {
	if nan, is := checkNan2(d, e, dp, ep); is {
		return nan
	}
	if dp.fl == flInf || ep.fl == flInf {
		if dp.fl == flInf && ep.fl == flInf {
			return d
		}
		return qNaNRaw
	}

	dexp, dsignificand := unsubnormal(dp.exp, dp.significand.lo)
	eexp, _ := unsubnormal(ep.exp, ep.significand.lo)

	delta := dexp - eexp
	if delta < -1 { // -1 avoids rounding range
		return zeroRaw
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
// It uses [DefaultContext] to call [Context.ToIntegral].
func (d Decimal) ToIntegral() Decimal {
	return DefaultContext.ToIntegral(d)
}

var decPartsOne decParts = unpack(One)

// ToIntegral rounds d to a nearby integer.
func (ctx Context) ToIntegral(d Decimal) Decimal {
	var dp decParts
	dp.unpack(d)
	if !dp.fl.normal() || dp.exp >= 0 {
		return d
	}
	return newDec(ctx.roundRefRaw(d, One, &dp, &decPartsOne).bits)
}

func (ctx Context) round(s, p uint64) (uint64, bool) {
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
		for significand < decimalBase {
			significand *= 10
			exp--
		}
	}
	return exp, significand
}

func resubnormal(exp int16, significand uint64) (int16, uint64) {
	for exp < -expOffset || significand >= 10*decimalBase {
		significand /= 10
		exp++
	}
	return exp, significand
}
