package decimal

import "fmt"

// Abs computes ||d||.
func (d Decimal64) Abs() Decimal64 {
	return Decimal64{^neg64 & uint64(d.bits)}
}

// Add computes d + e
func (d Decimal64) Add(e Decimal64) Decimal64 {
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
	ep.updateMag()
	dp.updateMag()
	sep := dp.separation(ep)
	var roundStatus discardedDigit
	roundStatus = matchScales(&dp, &ep)
	// if the seperation of the numbers are more than 16 then we just return the larger number
	if sep > 16 { // TODO: return rounded significand (for round down/Ceiling)
		return d
	} else if sep < -16 {
		return e
	}
	var ans decParts
	if ep.sign != dp.sign {
		if ep.significand == dp.significand {
			return Zero64
		}
		if dp.significand < ep.significand {
			dp, ep = ep, dp
		}
		ans.sign = dp.sign
		if dp.sign == 1 {
			dp.sign, ep.sign = ep.sign, dp.sign
		}
	} else if dp.sign == 1 {
		ans.sign = 1
		dp.sign, ep.sign = 0, 0
	} else {
		ans.sign = 0
	}
	ans.significand = dp.significand + uint64(1-2*(ep.sign))*ep.significand
	ans.exp = dp.exp
	ans.exp, ans.significand = renormalize(ans.exp, ans.significand)
	ans.significand = roundHalfUp.round(ans.significand, roundStatus)
	if ans.exp > expMax || ans.significand > maxSig {
		return infinities[ans.sign]
	}
	return newFromParts(ans.sign, ans.exp, ans.significand)
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

func (d Decimal64) FMA(e, f Decimal64) Decimal64 {
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
	ep.updateMag()
	dp.updateMag()
	fp.updateMag()
	var ans, ans2 decParts
	ans.sign = dp.sign ^ ep.sign
	var roundStatus discardedDigit
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
	significand := umul64(dp.significand, ep.significand)
	// ans.updateMag()
	ans.mag = numDecimalDigits(significand.lo) + numDecimalDigits(significand.hi)
	sep := ans.separation(fp)
	fmt.Println("significand.div64(decimal64Base)", significand.div64(decimal64Base), fp.significand)
	fmt.Println(sep, "sep", numDecimalDigits(significand.lo), numDecimalDigits(significand.hi))
	// if ans.exp+ans.mag > fp.exp+fp.mag {
	fmt.Println("asa", ans.exp-fp.exp)
	significand = umul64(fp.significand, powersOf10[sep+1]).sub(significand)
	// significand = significand.sub(umul64(fp.significand, powersOf10[sep+1])) //.sub(significand)

	ans.exp = dp.exp + ep.exp + numDecimalDigits(significand.hi)
	significand = significand.div64(powersOf10[numDecimalDigits(significand.hi)])
	ans.updateMag()
	ans.significand = significand.lo
	fmt.Println(sep, "significand.lo", significand.lo, significand.hi)
	ans.exp, ans.significand = renormalize(ans.exp, ans.significand)
	return newFromParts(ans.sign, ans.exp, ans.significand)
	// }
	if ans.exp >= -expOffset {
		ans.exp, ans.significand = renormalize(ans.exp, ans.significand)
	} else if ans.exp < 1-expMax {
		roundStatus = ans.rescale(-expOffset)
	}
	// 0.999999999999e-383
	// 0.999999999998999e-383
	if ans.isZero() {
		return f
	}
	if fp.isZero() {
		return newFromParts(ans.sign, ans.exp, ans.significand)
	}
	roundStatus = matchScales(&ans, &fp)
	if ans.sign != fp.sign {
		if ans.significand == fp.significand {
			return Zero64
		}
		if fp.significand < ans.significand {
			fp, ans = ans, fp
		}
		ans2.sign = fp.sign
		if fp.sign == 1 {
			fp.sign, ans.sign = ans.sign, fp.sign
		}
	} else if fp.sign == 1 {
		ans2.sign = 1
		fp.sign, ans.sign = 0, 0
	} else {
		ans2.sign = 0
	}
	ans2.significand = fp.significand + uint64(1-2*(ans.sign))*roundHalfUp.round(ans.significand, roundStatus)
	ans2.exp = fp.exp
	if ans2.isZero() {
		return Zero64
	}
	ans2.exp, ans2.significand = renormalize(ans2.exp, ans2.significand)
	if ans2.exp > expMax || ans2.significand > maxSig {
		return infinities[ans2.sign]
	}
	return newFromParts(ans2.sign, ans2.exp, ans2.significand)
}

// Mul computes d * e
func (d Decimal64) Mul(e Decimal64) Decimal64 {
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
	ans.significand = roundHalfEven.round(ans.significand, roundStatus)
	if ans.significand > maxSig || ans.exp > expMax {
		return infinities[ans.sign]
	}
	return newFromParts(ans.sign, ans.exp, ans.significand)
}

// Neg computes -d.
func (d Decimal64) Neg() Decimal64 {
	return Decimal64{neg64 ^ uint64(d.bits)}
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

// Sub returns d - e.
func (d Decimal64) Sub(e Decimal64) Decimal64 {
	return d.Add(e.Neg())
}
