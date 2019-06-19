package decimal

// add128 adds two decParts with full precision in 128 bits of significand
func (dp *decParts) add128(ep *decParts) decParts {
	logicCheck(ep.exp == dp.exp, "ep.exp == dp.exp")
	var ans decParts
	ans.exp = dp.exp
	if dp.sign == ep.sign {
		ans.sign = dp.sign
		ans.significand = dp.significand.add(ep.significand)
	} else {
		if ep.significand.gt(dp.significand) {
			ans.sign = ep.sign
			ans.significand = ep.significand.sub(dp.significand)
		} else if ep.significand.lt(dp.significand) {
			ans.sign = dp.sign
			ans.significand = dp.significand.sub(ep.significand)
		} else {
			ans.significand = uint128T{0, 0}
		}
	}
	return ans
}

func (dp *decParts) matchScales128(ep *decParts) {
	expDiff := ep.exp - dp.exp
	if (ep.significand != uint128T{0, 0}) {
		if expDiff < 0 {
			dp.significand = dp.significand.mul(powerOfTen128(expDiff))
			dp.exp += expDiff
		} else if expDiff > 0 {
			ep.significand = ep.significand.mul(powerOfTen128(expDiff))
			ep.exp -= expDiff
		}
	}
}

func (dp *decParts) matchSignificandDigits(ep *decParts) {
	expDiff := ep.significand.numDecimalDigits() - dp.significand.numDecimalDigits()
	if expDiff >= 0 {
		dp.significand = dp.significand.mul(powerOfTen128(expDiff + 1))
		dp.exp -= expDiff + 1
		return
	}
	ep.significand = ep.significand.mul(powerOfTen128(-expDiff - 1))
	ep.exp -= -expDiff - 1
}

func (dp *decParts) roundToLo() discardedDigit {
	var rndStatus discardedDigit
	if dp.significand.numDecimalDigits() > 16 {
		var remainder uint64
		expDiff := dp.significand.numDecimalDigits() - 16
		dp.exp += expDiff
		dp.significand, remainder = dp.significand.divrem64(powersOf10[expDiff])
		rndStatus = roundStatus(remainder, 0, expDiff)
	}
	return rndStatus
}

func (dec *decParts) isZero() bool {
	return dec.significand.lo == 0 && dec.significand.hi == 0 && dec.fl == flNormal
}

func (dec *decParts) isInf() bool {
	return dec.fl == flInf
}

func (dec *decParts) isNaN() bool {
	return dec.fl == flQNaN || dec.fl == flSNaN
}

func (dec *decParts) isQNaN() bool {
	return dec.fl == flQNaN
}

func (dec *decParts) isSNaN() bool {
	return dec.fl == flSNaN
}

func (dec *decParts) isSubnormal() bool {
	return dec.significand.lo != 0 && dec.significand.lo < decimal64Base && dec.fl == flNormal
}

// separation gets the separation in decimal places of the MSD's of two decimal 64s
func (dec *decParts) separation(eDec decParts) int {
	return dec.mag + dec.exp - eDec.mag - eDec.exp
}

// removeZeros removes zeros and increments the exponent to match.
func (dec *decParts) removeZeros() {
	zeros := countTrailingZeros(dec.significand.lo)
	dec.significand.lo /= powersOf10[zeros]
	dec.exp += zeros
}

// updateMag updates the magnitude of the dec object
func (dec *decParts) updateMag() {
	dec.mag = dec.significand.numDecimalDigits()
}

// isinf returns true if the decimal is an infinty
func (dec *decParts) isinf() bool {
	return dec.fl == flInf
}

func (dec *decParts) rescale(targetExp int) (rndStatus discardedDigit) {
	expDiff := targetExp - dec.exp
	mag := dec.mag
	rndStatus = roundStatus(dec.significand.lo, dec.exp, targetExp)
	if expDiff > mag {
		dec.significand.lo, dec.exp = 0, targetExp
		return
	}
	divisor := powersOf10[expDiff]
	dec.significand.lo = dec.significand.lo / divisor
	dec.exp = targetExp
	return
}
