package decimal

// Abs computes ||d||.
func (d Decimal64) Abs() Decimal64 {
	return Decimal64{^neg64 & uint64(d.bits)}
}

// Add computes d + e with default rounding
func (d Decimal64) Add(e Decimal64) Decimal64 {
	return DefaultContext.Add(d, e)
}

// FMA computes d*e + f with default rounding
func (d Decimal64) FMA(e, f Decimal64) Decimal64 {
	return DefaultContext.FMA(d, e, f)
}

// Mul computes d * e with default rounding
func (d Decimal64) Mul(e Decimal64) Decimal64 {
	return DefaultContext.Mul(d, e)
}

// Sub returns d - e.
func (d Decimal64) Sub(e Decimal64) Decimal64 {
	return d.Add(e.Neg())
}

// Cmp returns:
//
//   -2 if d or e is NaN
//   -1 if d <  e
//    0 if d == e (incl. -0 == 0, -Inf == -Inf, and +Inf == +Inf)
//   +1 if d >  e
//
func (d Decimal64) Cmp(e Decimal64) int {
	flavor1, _, _, significand1 := d.parts()
	flavor2, _, _, significand2 := e.parts()
	if flavor1 == flSNaN || flavor2 == flSNaN {
		return -2
	}
	if flavor1 == flQNaN || flavor2 == flQNaN {
		return -2
	}
	if significand1 == 0 && significand2 == 0 {
		return 0
	}
	if d == e {
		return 0
	}
	d = d.Sub(e)
	return 1 - 2*int(d.bits>>63)
}

// Neg computes -d.
func (d Decimal64) Neg() Decimal64 {
	return Decimal64{neg64 ^ d.bits}
}

// Quo computes d / e.
func (d Decimal64) Quo(e Decimal64) Decimal64 {
	flavor1, sign1, exp1, significand1 := d.parts()
	flavor2, sign2, exp2, significand2 := e.parts()
	if flavor1 == flSNaN || flavor2 == flSNaN {
		return SNaN64
	}
	if flavor1 == flQNaN || flavor2 == flQNaN {
		return QNaN64
	}
	sign := sign1 ^ sign2
	if d == Zero64 || d == NegZero64 {
		if e == Zero64 || e == NegZero64 {
			return QNaN64
		}
		return zeroes[sign]
	}
	if flavor1 == flInf {
		if flavor2 == flInf {
			return QNaN64
		}
		return infinities[sign]
	}
	if flavor2 == flInf {
		return zeroes[sign]
	}
	if e == Zero64 || e == NegZero64 {
		return infinities[sign1]
	}
	exp := exp1 - exp2 - 16
	significand := umul64(10*decimal64Base, significand1).div64(significand2)
	exp, significand.lo = renormalize(exp, significand.lo)
	if significand.lo > maxSig || exp > expMax {
		return infinities[sign]
	}

	return newFromParts(sign, exp, significand.lo)
}

// Sqrt computes √d.
func (d Decimal64) Sqrt() Decimal64 {
	flavor, sign, exp, significand := d.parts()
	switch flavor {
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
func (dp *decParts) round128to64() discardedDigit {
	var rndStatus discardedDigit
	if dp.significand128.numDecimalDigits() > 16 {
		var remainder uint64
		expDiff := dp.significand128.numDecimalDigits() - 16
		dp.exp += expDiff
		dp.significand128, remainder = dp.significand128.divrem64(powersOf10[expDiff])
		rndStatus = roundStatus(remainder, 0, expDiff)
	}
	dp.significand = dp.significand128.lo
	return rndStatus
}

func (dp *decParts) Add128(ep *decParts) decParts {
	targetExp := ep.exp - dp.exp
	var ans decParts
	if (ep.significand128 != uint128T{0, 0}) {
		if targetExp < 0 {
			dp.significand128 = dp.significand128.mul(powerOfTen128(targetExp))
			dp.exp += targetExp
		} else if targetExp > 0 {
			ep.significand128 = ep.significand128.mul(powerOfTen128(targetExp))
			ep.exp -= (targetExp)
		}
		if dp.sign == ep.sign {
			ans.sign = dp.sign
			dp.significand128 = dp.significand128.add(ep.significand128)
		} else {
			if ep.significand128.gt(dp.significand128) {
				ans.sign = ep.sign
				dp.sign = ep.sign
				dp.significand128 = ep.significand128.sub(dp.significand128)
			} else if ep.significand128.lt(dp.significand128) {
				ans.sign = dp.sign
				dp.significand128 = dp.significand128.sub(ep.significand128)
			} else {
				ans.significand = 0
				ans.exp = 0
				return ans
			}
		}
	}
	ans.significand128 = dp.significand128
	ans.exp += dp.exp
	return ans
}

// Add computes d + e
func (ctx Context64) Add(d, e Decimal64) Decimal64 {
	dp := d.getParts()
	ep := e.getParts()
	if dp.isNan() || ep.isNan() {
		return *propagateNan(&dp, &ep)
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
	if dp.significand == 0 {
		return e
	} else if ep.significand == 0 {
		return d
	}
	ep.removeZeros()
	dp.removeZeros()
	ep.updateMag()
	dp.updateMag()
	dp.significand128.lo = dp.significand
	ep.significand128.lo = ep.significand
	sep := dp.separation(ep)

	if sep < 0 {
		dp, ep = ep, dp
		sep = -sep
	}
	if sep > 17 {
		return *dp.dec
	}
	var rndStatus discardedDigit
	ans := dp.Add128(&ep)
	rndStatus = ans.round128to64()
	ans.updateMag()
	if ans.exp < -expOffset {
		rndStatus = ans.rescale(-expOffset)
	}
	ans.significand = ctx.roundingMode.round(ans.significand, rndStatus)
	if ans.exp >= -expOffset && ans.significand != 0 {
		ans.exp, ans.significand = renormalize(ans.exp, ans.significand)
	}
	if ans.exp > expMax || ans.significand > maxSig {
		return infinities[ans.sign]
	}
	return newFromParts(ans.sign, ans.exp, ans.significand)
}

// FMA computes d*e + f
func (ctx Context64) FMA(d, e, f Decimal64) Decimal64 {
	dp := d.getParts()
	ep := e.getParts()
	fp := f.getParts()
	if dp.fl == flSNaN {
		return d
	}
	if ep.fl == flSNaN {
		return e
	}
	if fp.fl == flSNaN {
		return f
	}
	if dp.fl == flQNaN {
		return d
	}
	if ep.fl == flQNaN {
		return e
	}
	if fp.fl == flQNaN {
		return f
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
	if ep.significand == 0 || dp.significand == 0 {
		return f
	}
	if fp.fl == flInf {
		return infinities[fp.sign]
	}

	var rndStatus discardedDigit
	ep.updateMag()
	dp.updateMag()
	fp.updateMag()
	ep.removeZeros()
	dp.removeZeros()
	ans.exp = dp.exp + ep.exp
	ans.significand128 = umul64(dp.significand, ep.significand)
	fp.significand128 = uint128T{fp.significand, 0}
	ans.mag = ans.significand128.numDecimalDigits()
	sep := ans.separation(fp)
	if fp.significand != 0 {
		if sep < -17 {
			return f
		} else if sep <= 17 {
			ans = ans.Add128(&fp)
		}
	}
	rndStatus = ans.round128to64()
	if ans.exp < -expOffset {
		rndStatus = ans.rescale(-expOffset)
	}
	ans.updateMag()
	ans.significand = ctx.roundingMode.round(ans.significand, rndStatus)
	if ans.exp >= -expOffset && ans.significand != 0 {
		ans.exp, ans.significand = renormalize(ans.exp, ans.significand)
	}
	if ans.exp > expMax || ans.significand > maxSig {
		return infinities[ans.sign]
	}
	return newFromParts(ans.sign, ans.exp, ans.significand)
}

// Mul computes d * e
func (ctx Context64) Mul(d, e Decimal64) Decimal64 {
	dp := d.getParts()
	ep := e.getParts()
	if ep.fl == flQNaN || ep.fl == flSNaN || dp.fl == flQNaN || dp.fl == flSNaN {
		return *propagateNan(&dp, &ep)
	}
	var ans decParts
	ans.sign = dp.sign ^ ep.sign
	if dp.fl == flInf || ep.fl == flInf {
		if ep.isZero() || dp.isZero() {
			return QNaN64
		}
		return infinities[ans.sign]
	}
	if ep.significand == 0 || dp.significand == 0 {
		return zeroes[ans.sign]
	}
	ep.updateMag()
	dp.updateMag()
	var roundStatus discardedDigit
	significand := umul64(dp.significand, ep.significand)
	ans.exp = dp.exp + ep.exp + 15
	significand = significand.div64(decimal64Base)
	ans.significand = significand.lo
	ans.updateMag()
	if ans.exp >= -expOffset {
		ans.exp, ans.significand = renormalize(ans.exp, ans.significand)
	} else if ans.exp < 1-expMax {
		roundStatus = ans.rescale(-expOffset)
	}
	ans.significand = ctx.roundingMode.round(ans.significand, roundStatus)
	if ans.significand > maxSig || ans.exp > expMax {
		return infinities[ans.sign]
	}
	return newFromParts(ans.sign, ans.exp, ans.significand)
}
