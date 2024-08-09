package decimal

// Abs computes ||d||.
func (d Decimal64) Abs() Decimal64 {
	if d.IsNaN() {
		return d
	}
	return Decimal64{bits: ^neg64 & uint64(d.bits)}.debug()
}

// Add computes d + e with default rounding
func (d Decimal64) Add(e Decimal64) Decimal64 {
	return DefaultContext.Add(d, e)
}

// FMA computes d*e + f with default rounding.
func (d Decimal64) FMA(e, f Decimal64) Decimal64 {
	return DefaultContext.FMA(d, e, f)
}

// Mul computes d * e with default rounding.
func (d Decimal64) Mul(e Decimal64) Decimal64 {
	return DefaultContext.Mul(d, e)
}

// Sub returns d - e.
func (d Decimal64) Sub(e Decimal64) Decimal64 {
	return d.Add(e.Neg())
}

// Quo computes d / e with default rounding.
func (d Decimal64) Quo(e Decimal64) Decimal64 {
	return DefaultContext.Quo(d, e)
}

// Cmp returns:
//
//	-2 if d or e is NaN
//	-1 if d <  e
//	 0 if d == e (incl. -0 == 0, -Inf == -Inf, and +Inf == +Inf)
//	+1 if d >  e
func (d Decimal64) Cmp(e Decimal64) int {
	var dp decParts
	dp.unpack(d)
	var ep decParts
	ep.unpack(e)
	if _, isNan := checkNan(&dp, &ep); isNan {
		return -2
	}
	return cmp(&dp, &ep)
}

// Cmp64 returns the same output as Cmp as a Decimal64, unless d or e is NaN, in
// which case it returns a corresponding NaN result.
func (d Decimal64) Cmp64(e Decimal64) Decimal64 {
	var dp decParts
	dp.unpack(d)
	var ep decParts
	ep.unpack(e)
	if n, isNan := checkNan(&dp, &ep); isNan {
		return n
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
	var dp decParts
	dp.unpack(d)
	var ep decParts
	ep.unpack(e)

	dnan := dp.isNaN()
	enan := ep.isNaN()

	switch {
	case !dnan && !enan: // Fast path for non-NaNs.
		if sign*cmp(&dp, &ep) < 0 {
			return d
		}
		return e

	case dp.isSNaN():
		return d.quiet()
	case ep.isSNaN():
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

	dnan := dp.isNaN()
	enan := ep.isNaN()

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
	case dp.isSNaN():
		return d.quiet()
	case ep.isSNaN():
		return e.quiet()
	case !enan:
		return e
	default:
		return d
	}
}

// Neg computes -d.
func (d Decimal64) Neg() Decimal64 {
	if d.IsNaN() {
		return d
	}
	return Decimal64{bits: neg64 ^ d.bits}.debug()
}

// Logb return the integral log10 of d.
func (d Decimal64) Logb() Decimal64 {
	switch {
	case d.IsNaN():
		return d
	case d.IsZero():
		return NegInfinity64
	case d.IsInf():
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
	return Decimal64{bits: d.bits&^neg64 | e.bits&neg64}
}

// Quo computes d / e.
func (ctx Context64) Quo(d, e Decimal64) Decimal64 {
	var dp decParts
	dp.unpack(d)
	var ep decParts
	ep.unpack(e)
	if nan, isNan := checkNan(&dp, &ep); isNan {
		return nan
	}
	var ans decParts
	ans.sign = dp.sign ^ ep.sign
	if dp.isZero() {
		if ep.isZero() {
			return QNaN64
		}
		return zeroes[ans.sign]
	}
	if dp.isinf() {
		if ep.isinf() {
			return QNaN64
		}
		return infinities[ans.sign]
	}
	if ep.isinf() {
		return zeroes[ans.sign]
	}
	if ep.isZero() {
		return infinities[ans.sign]
	}
	dp.matchSignificandDigits(&ep)
	ans.exp = dp.exp - ep.exp
	for {
		for dp.significand.gt(ep.significand) {
			dp.significand = dp.significand.sub(ep.significand)
			ans.significand = ans.significand.add(uint128T{1, 0})
		}
		if dp.significand == (uint128T{}) || ans.significand.numDecimalDigits() == 16 {
			break
		}
		ans.significand = ans.significand.mulBy10()
		dp.significand = dp.significand.mulBy10()
		ans.exp--
	}
	var rndStatus discardedDigit
	dp.significand = dp.significand.mul64(2)
	if dp.significand == (uint128T{}) {
		rndStatus = eq0
	} else if dp.significand.gt(ep.significand) {
		rndStatus = gt5
	} else if dp.significand.lt(ep.significand) {
		rndStatus = lt5
	} else {
		rndStatus = eq5
	}
	ans.significand.lo = ctx.roundingMode.round(ans.significand.lo, rndStatus)
	if ans.exp < -expOffset {
		rndStatus = ans.rescale(-expOffset)
		ans.significand.lo = ctx.roundingMode.round(ans.significand.lo, rndStatus)
	}
	if ans.exp >= -expOffset && ans.significand.lo != 0 {
		ans.exp, ans.significand.lo = renormalize(ans.exp, ans.significand.lo)
	}
	if ans.significand.lo > maxSig || ans.exp > expMax {
		return infinities[ans.sign]
	}
	return newFromParts(ans.sign, ans.exp, ans.significand.lo)
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
	case flNormal:
	}
	if significand == 0 {
		return d
	}
	if sign == 1 {
		return QNaN64
	}
	if exp&1 == 1 {
		exp--
		significand *= 10
	}
	sqrt := umul64(10*decimal64Base, significand).sqrt()
	exp, significand = renormalize(exp/2-8, sqrt)
	return newFromParts(sign, exp, significand)
}

// Add computes d + e
func (ctx Context64) Add(d, e Decimal64) Decimal64 {
	var dp decParts
	dp.unpack(d)
	var ep decParts
	ep.unpack(e)
	if nan, isNan := checkNan(&dp, &ep); isNan {
		return nan
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
	ep.removeZeros()
	dp.removeZeros()
	sep := dp.separation(&ep)

	if sep < 0 {
		dp, ep = ep, dp
		sep = -sep
	}
	if sep > 17 {
		return dp.original
	}
	var rndStatus discardedDigit
	dp.matchScales128(&ep)
	ans := dp.add128(&ep)
	rndStatus = ans.roundToLo()
	if ans.exp < -expOffset {
		rndStatus = ans.rescale(-expOffset)
	}
	ans.significand.lo = ctx.roundingMode.round(ans.significand.lo, rndStatus)
	if ans.exp >= -expOffset && ans.significand.lo != 0 {
		ans.exp, ans.significand.lo = renormalize(ans.exp, ans.significand.lo)
	}
	if ans.exp > expMax || ans.significand.lo > maxSig {
		return infinities[ans.sign]
	}
	return newFromParts(ans.sign, ans.exp, ans.significand.lo)
}

// FMA computes d*e + f
func (ctx Context64) FMA(d, e, f Decimal64) Decimal64 {
	var dp decParts
	dp.unpack(d)
	var ep decParts
	ep.unpack(e)
	var fp decParts
	fp.unpack(f)
	if nan, isNan := checkNan3(&dp, &ep, &fp); isNan {
		return nan
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
	ans.significand = umul64(dp.significand.lo, ep.significand.lo)
	sep := ans.separation(&fp)
	if fp.significand.lo != 0 {
		if sep < -17 {
			return f
		} else if sep <= 17 {
			ans = ans.add128(&fp)
		}
	}
	rndStatus = ans.roundToLo()
	if ans.exp < -expOffset {
		rndStatus = ans.rescale(-expOffset)
	}
	ans.significand.lo = ctx.roundingMode.round(ans.significand.lo, rndStatus)
	if ans.exp >= -expOffset && ans.significand.lo != 0 {
		ans.exp, ans.significand.lo = renormalize(ans.exp, ans.significand.lo)
	}
	if ans.exp > expMax || ans.significand.lo > maxSig {
		return infinities[ans.sign]
	}
	return newFromParts(ans.sign, ans.exp, ans.significand.lo)
}

// Mul computes d * e
func (ctx Context64) Mul(d, e Decimal64) Decimal64 {
	var dp decParts
	dp.unpack(d)
	var ep decParts
	ep.unpack(e)
	if nan, isNan := checkNan(&dp, &ep); isNan {
		return nan
	}
	var ans decParts
	ans.sign = dp.sign ^ ep.sign
	if dp.fl == flInf || ep.fl == flInf {
		if ep.isZero() || dp.isZero() {
			return QNaN64
		}
		return infinities[ans.sign]
	}
	if ep.significand.lo == 0 || dp.significand.lo == 0 {
		return zeroes[ans.sign]
	}
	var roundStatus discardedDigit
	ans.significand = umul64(dp.significand.lo, ep.significand.lo)
	ans.exp = dp.exp + ep.exp + 15
	ans.significand = ans.significand.div64(decimal64Base)
	if ans.exp >= -expOffset {
		ans.exp, ans.significand.lo = renormalize(ans.exp, ans.significand.lo)
	} else if ans.exp < 1-expMax {
		roundStatus = ans.rescale(-expOffset)
	}
	ans.significand.lo = ctx.roundingMode.round(ans.significand.lo, roundStatus)
	if ans.significand.lo > maxSig || ans.exp > expMax {
		return infinities[ans.sign]
	}
	return newFromParts(ans.sign, ans.exp, ans.significand.lo)
}
