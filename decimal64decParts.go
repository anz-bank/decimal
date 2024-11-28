package decimal

// decParts stores the constituting decParts of a decimal64.
type decParts struct {
	significand uint128T
	exp         int16
	sign        int8
	fl          flavor
}

func unpack(d Decimal64) decParts {
	var dp decParts
	dp.unpack(d)
	return dp
}

func (dp *decParts) decimal64() Decimal64 {
	return newFromParts(dp.sign, dp.exp, dp.significand.lo)
}

// add128 adds two decParts with full precision in 128 bits of significand
func (ans *decParts) add128(dp, ep *decParts) {
	dp.matchScales128(ep)
	ans.exp = dp.exp
	if dp.sign == ep.sign {
		ans.sign = dp.sign
		ans.significand.add(&dp.significand, &ep.significand)
	} else {
		if dp.significand.lt(&ep.significand) {
			ans.sign = ep.sign
			ans.significand.sub(&ep.significand, &dp.significand)
		} else if ep.significand.lt(&dp.significand) {
			ans.sign = dp.sign
			ans.significand.sub(&dp.significand, &ep.significand)
		} else {
			ans.significand = uint128T{0, 0}
		}
	}
}

// add64 adds the low 64 bits of two decParts
func (ans *decParts) add64(dp, ep *decParts) {
	ans.exp = dp.exp
	switch {
	case dp.sign == ep.sign:
		ans.sign = dp.sign
		ans.significand.lo = dp.significand.lo + ep.significand.lo
	case dp.significand.lt(&ep.significand):
		ans.sign = ep.sign
		ans.significand.lo = ep.significand.lo - dp.significand.lo
	case ep.significand.lt(&dp.significand):
		ans.sign = dp.sign
		ans.significand.lo = dp.significand.lo - ep.significand.lo
	}
}

// add128 adds two decParts with full precision in 128 bits of significand
func (ans *decParts) add128V2(dp, ep *decParts) {
	ans.exp = dp.exp
	switch {
	case dp.sign == ep.sign:
		ans.sign = dp.sign
		ans.significand.add(&dp.significand, &ep.significand)
	case dp.significand.lt(&ep.significand):
		ans.sign = ep.sign
		ans.significand.sub(&ep.significand, &dp.significand)
	case ep.significand.lt(&dp.significand):
		ans.sign = dp.sign
		ans.significand.sub(&dp.significand, &ep.significand)
	}
}

func (dp *decParts) matchScales128(ep *decParts) {
	expDiff := ep.exp - dp.exp
	if (ep.significand != uint128T{0, 0}) {
		if expDiff < 0 {
			dp.significand.mul(&dp.significand, &tenToThe128[-expDiff])
			dp.exp += expDiff
		} else if expDiff > 0 {
			ep.significand.mul(&ep.significand, &tenToThe128[expDiff])
			ep.exp -= expDiff
		}
	}
}

func (dp *decParts) roundToLo() discardedDigit {
	var rndStatus discardedDigit

	if ds := &dp.significand; ds.hi > 0 || ds.lo >= 10*decimal64Base {
		var remainder uint64
		expDiff := int16(ds.numDecimalDigits()) - 16
		dp.exp += expDiff
		remainder = ds.divrem64(ds, tenToThe[expDiff])
		rndStatus = roundStatus(remainder, expDiff)
	}
	return rndStatus
}

func (dp *decParts) isZero() bool {
	return dp.significand == uint128T{} && dp.fl.normal()
}

func (dp *decParts) isSubnormal() bool {
	return (dp.significand != uint128T{}) && dp.significand.lo < decimal64Base && dp.fl.normal()
}

// separation gets the separation in decimal places of the MSD's of two decimal 64s
func (dp *decParts) separation(ep *decParts) int16 {
	sep := int16(dp.significand.numDecimalDigits()) + dp.exp
	sep -= int16(ep.significand.numDecimalDigits()) + ep.exp
	return sep
}

// separation gets the separation in decimal places of the MSD's of two decimal 64s
func (dp *decParts) separationV2(ep *decParts) int16 {
	sep := int16(numDecimalDigitsU64(dp.significand.lo)) + dp.exp
	sep -= int16(numDecimalDigitsU64(ep.significand.lo)) + ep.exp
	return sep
}

// removeZeros removes zeros and increments the exponent to match.
func (dp *decParts) removeZeros() {
	e := dp.exp
	n := dp.significand.lo
	b := n / 1_0000_0000_0000_0000
	if n == b*1_0000_0000_0000_0000 {
		e += 16
		n = b
	}
	b = n / 1_0000_0000
	if n == b*1_0000_0000 {
		e += 8
		n = b
	}
	b = n / 10000
	if n == b*10000 {
		e += 4
		n = b
	}
	b = n / 100
	if n == b*100 {
		e += 2
		n = b
	}
	b = n / 10
	if n == b*10 {
		e++
		n = b
	}
	dp.significand.lo = n
	dp.exp = e
}

// isinf returns true if the decimal is an infinty
func (dp *decParts) isinf() bool {
	return dp.fl == flInf
}

func (dp *decParts) rescale(targetExp int16) discardedDigit {
	expDiff := targetExp - dp.exp
	rndStatus := roundStatus(dp.significand.lo, expDiff)
	if expDiff > int16(dp.significand.numDecimalDigits()) {
		dp.significand.lo, dp.exp = 0, targetExp
		return rndStatus
	}
	divisor := tenToThe[expDiff]
	dp.significand.lo = dp.significand.lo / divisor
	dp.exp = targetExp
	return rndStatus
}

func (dp *decParts) unpack(d Decimal64) {
	dp.fl = d.flavor()
	dp.unpackV2(d)
}

func (dp *decParts) unpackV2(d Decimal64) {
	dp.sign = int8(d.bits >> 63)
	switch dp.fl {
	case flNormal53:
		// s EEeeeeeeee   (0)ttt tttttttttt tttttttttt tttttttttt tttttttttt tttttttttt
		//   EE ∈ {00, 01, 10}
		dp.exp = int16((d.bits>>(63-10))&(1<<10-1)) - expOffset
		dp.significand.lo = d.bits & (1<<53 - 1)
	case flNormal51:
		// s 11EEeeeeeeee (100)t tttttttttt tttttttttt tttttttttt tttttttttt tttttttttt
		//     EE ∈ {00, 01, 10}
		dp.exp = int16((d.bits>>(63-12))&(1<<10-1)) - expOffset
		dp.significand.lo = d.bits&(1<<51-1) | (1 << 53)
	case flInf:
	default: // NaN
		dp.significand.lo = d.bits & (1<<51 - 1) // Payload
	}
}

// https://en.wikipedia.org/wiki/Decimal64_floating-point_format#Binary_integer_significand_field
var flavMap = [...]flavor{
	/* 0000xx */ flNormal53, flNormal53, flNormal53, flNormal53,
	/* 0001xx */ flNormal53, flNormal53, flNormal53, flNormal53,
	/* 0010xx */ flNormal53, flNormal53, flNormal53, flNormal53,
	/* 0011xx */ flNormal53, flNormal53, flNormal53, flNormal53,
	/* 0100xx */ flNormal53, flNormal53, flNormal53, flNormal53,
	/* 0101xx */ flNormal53, flNormal53, flNormal53, flNormal53,
	/* 0110xx */ flNormal53, flNormal53, flNormal53, flNormal53,
	/* 0111xx */ flNormal53, flNormal53, flNormal53, flNormal53,
	/* 1000xx */ flNormal53, flNormal53, flNormal53, flNormal53,
	/* 1001xx */ flNormal53, flNormal53, flNormal53, flNormal53,
	/* 1010xx */ flNormal53, flNormal53, flNormal53, flNormal53,
	/* 1011xx */ flNormal53, flNormal53, flNormal53, flNormal53,
	/* 1100xx */ flNormal51, flNormal51, flNormal51, flNormal51,
	/* 1101xx */ flNormal51, flNormal51, flNormal51, flNormal51,
	/* 1110xx */ flNormal51, flNormal51, flNormal51, flNormal51,
	/* 11110x */ flInf, flInf,
	/* 111110 */ flQNaN,
	/* 111111 */ flSNaN,
}

func (d Decimal64) flavor() flavor {
	return flavMap[int(d.bits>>(64-7))%len(flavMap)]
}
