package decimal

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
	roundingMode := newRoundContext(roundHalfUp)
	roundingMode.matchScales(&dp, &ep)
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
	ans.significand = roundingMode.round(ans.significand)
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

// Mul computes d * e.
func (d Decimal64) Mul(e Decimal64) Decimal64 {
	flavor1, sign1, exp1, significand1 := d.parts()
	flavor2, sign2, exp2, significand2 := e.parts()
	if flavor1 == flSNaN || flavor2 == flSNaN {
		return SNaN64
	}
	if flavor1 == flQNaN || flavor2 == flQNaN {
		return QNaN64
	}
	sign := sign1 ^ sign2
	if d == Zero64 || d == NegZero64 || e == Zero64 || e == NegZero64 {
		return zeroes[sign]
	}
	if flavor1 == flInf || flavor2 == flInf {
		return infinities[sign]
	}
	exp := exp1 + exp2 + 15
	significand := umul64(significand1, significand2).div64(decimal64Base)
	exp, significand.lo = renormalize(exp, significand.lo)
	if significand.lo > maxSig || exp > expMax {
		return infinities[sign]
	}
	return newFromParts(sign, exp, significand.lo)
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
