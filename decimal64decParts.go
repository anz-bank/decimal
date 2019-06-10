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
	expDiff := ep.significand.numDecimalDigits() - dp.significand.numDecimalDigits() + 1

	if expDiff > 0 {
		dp.significand = dp.significand.mul(powerOfTen128(expDiff))
		dp.exp -= expDiff
	}
	fmt.Println("d")
	return
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
