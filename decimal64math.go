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

// Add computes d + e
func (ctx Context64) Add(d, e Decimal64) Decimal64 {
	dp := d.getParts()
	ep := e.getParts()
	if ok, dec := propagateNan(&dp, &ep); ok {
		return *dec
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
	ep.updateMag()
	dp.updateMag()
	sep := dp.separation(ep)

	if sep < 0 {
		dp, ep = ep, dp
		sep = -sep
	}
	if sep > 17 {
		return *dp.dec
	}
	var rndStatus discardedDigit
	dp.matchScales128(&ep)
	ans := dp.add128(&ep)
	rndStatus = ans.roundToLo()
	ans.updateMag()
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
	dp := d.getParts()
	ep := e.getParts()
	fp := f.getParts()
	if ok, dec := propagateNan(&dp, &ep, &fp); ok {
		return *dec
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
	ep.updateMag()
	dp.updateMag()
	fp.updateMag()
	ep.removeZeros()
	dp.removeZeros()
	ans.exp = dp.exp + ep.exp
	ans.significand = umul64(dp.significand.lo, ep.significand.lo)
	ans.mag = ans.significand.numDecimalDigits()
	sep := ans.separation(fp)
	if fp.significand.lo != 0 {
		if sep < -17 {
			return f
		} else if sep <= 17 {
			ans.matchScales128(&fp)
			ans = ans.add128(&fp)
		}
	}
	rndStatus = ans.roundToLo()
	if ans.exp < -expOffset {
		rndStatus = ans.rescale(-expOffset)
	}
	ans.updateMag()
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
	dp := d.getParts()
	ep := e.getParts()
	if ok, dec := propagateNan(&dp, &ep); ok {
		return *dec
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
	ep.updateMag()
	dp.updateMag()
	var roundStatus discardedDigit
	significand := umul64(dp.significand.lo, ep.significand.lo)
	ans.exp = dp.exp + ep.exp + 15
	significand = significand.div64(decimal64Base)
	ans.significand.lo = significand.lo
	ans.updateMag()
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
