package decimal

// decParts stores the constituting decParts of a decimal64.
type decParts struct {
	fl          flavor
	sign        int
	exp         int
	significand uint128T
	original    Decimal64
}

// add128 adds two decParts with full precision in 128 bits of significand
func (dp *decParts) add128(ep *decParts) decParts {
	dp.matchScales128(ep)
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

func (dp *decParts) isZero() bool {
	return (dp.significand == uint128T{}) && dp.significand.hi == 0 && dp.fl == flNormal
}

func (dp *decParts) isInf() bool {
	return dp.fl == flInf
}

func (dp *decParts) isNaN() bool {
	return dp.fl&(flQNaN|flSNaN) != 0
}

func (dp *decParts) isQNaN() bool {
	return dp.fl == flQNaN
}

func (dp *decParts) isSNaN() bool {
	return dp.fl == flSNaN
}

func (dp *decParts) nanWeight() int {
	return int(dp.significand.lo)
}

func (dp *decParts) isSubnormal() bool {
	return (dp.significand != uint128T{}) && dp.significand.lo < decimal64Base && dp.fl == flNormal
}

// separation gets the separation in decimal places of the MSD's of two decimal 64s
func (dp *decParts) separation(ep *decParts) int {
	return dp.significand.numDecimalDigits() + dp.exp - ep.significand.numDecimalDigits() - ep.exp
}

// removeZeros removes zeros and increments the exponent to match.
func (dp *decParts) removeZeros() {
	zeros := countTrailingZeros(dp.significand.lo)
	dp.significand.lo /= powersOf10[zeros]
	dp.exp += zeros
}

// isinf returns true if the decimal is an infinty
func (dp *decParts) isinf() bool {
	return dp.fl == flInf
}

func (dp *decParts) rescale(targetExp int) (rndStatus discardedDigit) {
	expDiff := targetExp - dp.exp
	mag := dp.significand.numDecimalDigits()
	rndStatus = roundStatus(dp.significand.lo, dp.exp, targetExp)
	if expDiff > mag {
		dp.significand.lo, dp.exp = 0, targetExp
		return
	}
	divisor := powersOf10[expDiff]
	dp.significand.lo = dp.significand.lo / divisor
	dp.exp = targetExp
	return
}

func (dp *decParts) unpack(d Decimal64) {
	dp.original = d
	dp.sign = int(d.bits >> 63)
	switch (d.bits >> (63 - 4)) & 0xf {
	case 15:
		switch (d.bits >> (63 - 6)) & 3 {
		case 0, 1:
			dp.fl = flInf
		case 2:
			dp.fl = flQNaN
			dp.significand.lo = d.bits & (1<<51 - 1) // Payload
			return
		case 3:
			dp.fl = flSNaN
			dp.significand.lo = d.bits & (1<<51 - 1) // Payload
			return
		}
	case 12, 13, 14:
		// s 11EEeeeeeeee (100)t tttttttttt tttttttttt tttttttttt tttttttttt tttttttttt
		//     EE ∈ {00, 01, 10}
		dp.fl = flNormal
		dp.exp = int((d.bits>>(63-12))&(1<<10-1)) - expOffset
		dp.significand.lo = d.bits&(1<<51-1) | (1 << 53)
	default:
		// s EEeeeeeeee   (0)ttt tttttttttt tttttttttt tttttttttt tttttttttt tttttttttt
		//   EE ∈ {00, 01, 10}
		dp.fl = flNormal
		dp.exp = int((d.bits>>(63-10))&(1<<10-1)) - expOffset
		dp.significand.lo = d.bits & (1<<53 - 1)
		if dp.significand.lo == 0 {
			dp.exp = 0
		}
	}
}
