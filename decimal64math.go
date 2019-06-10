package decimal

import "fmt"

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

// if C1 < C2
// nd ¼ digitsðC2Þ  digitsðC1Þ // table lookup
// C10 ¼ C1  10nd
// scale ¼ p  1
// ifðC10 < C2)
// scale ¼ scale þ 1
// endif
// C1	 ¼ C10  10scale
// Q0 ¼ 0
// e ¼ e1  e2  scale  nd // expected exponent
// else
// Q0 ¼ bC1=C2c, R ¼ C1  Q  C2 // long integer
// divide and remainder
// if ðR ¼¼ 0Þ
// return Q  10e1e2 // result is exact
// endif
// scale ¼ p  digitsðQÞ
// C1	 ¼ R  10scale
// Q0 ¼ Q0  10scale

// func (d Decimal64) Quo(e Decimal64) Decimal64 {
// 	dp, ep := d.getParts(), e.getParts()
// 	var ans decParts
// 	// if C1 < C2
// 	Q0 := uint128T{}
// 	if dp.significand.lt(ep.significand) {
// 		// nd = digits(C2) − digits(C1)
// 		nd := ep.significand.numDecimalDigits() - dp.significand.numDecimalDigits()
// 		// C1 = C1 · 10nd
// 		dp.significand = dp.significand.mul64(powersOf10[nd])
// 		// scale = p − 1
// 		scale := 16 - 1
// 		// if(C1 < C2)
// 		if dp.significand.gt(ep.significand) {
// 			// scale = scale + 1
// 			scale++
// 			// endif
// 		}
// 		// C1∗ = C1 · 10scale
// 		dp.significand = dp.significand.mul64(powersOf10[scale])
// 		// Q0=0
// 		// Q0 := uint128T{}
//
// 		// e = e1 − e2 − scale − nd // expected exponent
// 		ans.exp = dp.exp - ep.exp - scale - nd
// 	} else {
// 		// Q0 = C1/C2, R = C1 − Q0 · C2 // long
// 		// // integer divide and remainder
// 		Q0 := dp.significand.div64(ep.significand.lo)
// 		R := dp.significand.sub(Q0).mul(ep.significand)
// 		// if (R == 0)
// 		if (R == uint128T{}) {
// 			// return Q0 · 10e1−e2 // result is exact
// 			ans.significand = Q0
// 			ans.exp = dp.exp - ep.exp
// 			return newFromParts(ans.sign, ans.exp, ans.significand.lo)
//
// 			// endif
// 		}
// 		// scale = p − digits(Q0)
// 		scale := 16 - Q0.numDecimalDigits()
// 		// C1∗ = R · 10scale
// 		dp.significand = R.mul64(powersOf10[scale])
// 		// Q0 = Q0 · 10scale
// 		Q0 = Q0.mul64(powersOf10[scale])
// 		// e = e1 − e2 − scale // expected exponent
// 		ans.exp = dp.exp - ep.exp - scale
//
// 		// endif
// 	}
// 	// Q1 = C1∗ / C2, R = C1∗ − Q1 · C2
// 	logicCheck(ep.significand.hi == 0, "ep.significand.hi == 0")
// 	Q1 := dp.significand.div64(ep.significand.lo)
// 	R := dp.significand.sub(Q1.mul(ep.significand))
// 	// // multiprecision integer divide
// 	// Q = Q0 + Q1
// 	Q := Q0.add(Q1)
// 	// if (R == 0)
// 	if (R == uint128T{}) {
// 		// eliminate trailing zeros from Q:
// 		// find largest integer d s.t. Q/10d is exact
// 		// Q = Q/10d
// 		// e = e + d // adjust expected exponent
// 		ans.significand = Q
// 		ans.removeZeros()
// 		// if (e ≥ EMIN)
// 		if ans.exp >= -expOffset {
// 			// return Q · 10e
// 			return newFromParts(ans.sign, ans.exp, ans.significand.lo)
// 		}
// 		// endif
// 	}
// 	// if (e ≥ EMIN)
// 	if ans.exp >= -expOffset {
// 	}
// 	return newFromParts(ans.sign, ans.exp, ans.significand.lo)
// 	// round Q · 10e according to current rounding
// 	// mode
// 	// // rounding to nearest based on comparing
// 	// // C2 and 2 · R
// 	// else
// 	// compute correct result based on Property 1
// 	// // underflow
// 	// endif
// }

// Quo computes d / e.
func (d Decimal64) Quo(e Decimal64) Decimal64 {
	dp := d.getParts()
	ep := e.getParts()
	if ep.isNan() || dp.isNan() {
		return *propagateNan(&dp, &ep)
	}
	var ans decParts
	ans.sign = dp.sign ^ ep.sign
	if d == Zero64 || d == NegZero64 {
		if e == Zero64 || e == NegZero64 {
			return QNaN64
		}
		return zeroes[ans.sign]
	}
	if dp.fl == flInf {
		if ep.fl == flInf {
			return QNaN64
		}
		return infinities[ans.sign]
	}
	if ep.fl == flInf {
		return zeroes[ans.sign]
	}
	if ep.isZero() {
		return infinities[dp.sign]
	}
	// adjust := 0
	if dp.isZero() {
		return zeroes[ans.sign]
	}

	dp.matchSignificandDigits(&ep)
	logicCheck(dp.significand.gt(ep.significand), fmt.Sprintf("dp: %v ep: %v", dp.significand, ep.significand))
	ans.exp = dp.exp - ep.exp
	for {
		for dp.significand.gt(ep.significand) {

			dp.significand = dp.significand.sub(ep.significand)
			ans.significand = ans.significand.add(uint128T{1, 0})

		}
		if dp.significand == (uint128T{}) || ans.significand.numDecimalDigits() >= 20 {
			break
		}
		ans.significand = ans.significand.mulBy10()
		dp.significand = dp.significand.mulBy10()
		ans.exp--
	}
	logicCheck(ep.significand.hi == 0, "ep.significand == 0")
	// ans.significand = dp.significand.div64(ep.significand.lo)
	rndStatus := ans.roundToLo()
	logicCheck(ans.significand.hi == 0, "anssighi==0")
	ans.significand.lo = roundHalfUp.round(ans.significand.lo, rndStatus)
	ans.exp, ans.significand.lo = renormalize(ans.exp, ans.significand.lo)
	if ans.significand.lo > maxSig || ans.exp > expMax {
		return infinities[ans.sign]
	}

	return newFromParts(ans.sign, ans.exp, ans.significand.lo)
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
	if dec := propagateNan(&dp, &ep); dec != nil {
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
	if dec := propagateNan(&dp, &ep, &fp); dec != nil {
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
	if dec := propagateNan(&dp, &ep); dec != nil {
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
