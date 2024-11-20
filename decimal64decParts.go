package decimal

// decParts stores the constituting decParts of a decimal64.
type decParts struct {
	fl          flavor
	sign        int
	exp         int
	significand uint128T
	original    Decimal64
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
func (dp *decParts) add128(ep *decParts) decParts {
	dp.matchScales128(ep)
	var ans decParts
	ans.exp = dp.exp
	if dp.sign == ep.sign {
		ans.sign = dp.sign
		ans.significand = dp.significand.add(ep.significand)
	} else {
		if dp.significand.lt(ep.significand) {
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
			dp.significand = dp.significand.mul(tenToThe128[-expDiff])
			dp.exp += expDiff
		} else if expDiff > 0 {
			ep.significand = ep.significand.mul(tenToThe128[expDiff])
			ep.exp -= expDiff
		}
	}
}

func (dp *decParts) roundToLo() discardedDigit {
	var rndStatus discardedDigit

	if dsig := dp.significand; dsig.hi > 0 || dsig.lo >= 10*decimal64Base {
		var remainder uint64
		expDiff := dsig.numDecimalDigits() - 16
		dp.exp += expDiff
		dp.significand, remainder = dsig.divrem64(tenToThe[expDiff])
		rndStatus = roundStatus(remainder, 0, expDiff)
	}
	return rndStatus
}

func (dp *decParts) isZero() bool {
	return dp.significand == uint128T{} && dp.fl.normal()
}

func (dp *decParts) isInf() bool {
	return dp.fl == flInf
}

func (dp *decParts) isNaN() bool {
	return dp.fl&(flQNaN|flSNaN) != 0
}

func (dp *decParts) isSNaN() bool {
	return dp.fl == flSNaN
}

func (dp *decParts) isSubnormal() bool {
	return (dp.significand != uint128T{}) && dp.significand.lo < decimal64Base && dp.fl.normal()
}

// separation gets the separation in decimal places of the MSD's of two decimal 64s
func (dp *decParts) separation(ep *decParts) int {
	return dp.significand.numDecimalDigits() + dp.exp - ep.significand.numDecimalDigits() - ep.exp
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

func (dp *decParts) rescale(targetExp int) (rndStatus discardedDigit) {
	expDiff := targetExp - dp.exp
	mag := dp.significand.numDecimalDigits()
	rndStatus = roundStatus(dp.significand.lo, dp.exp, targetExp)
	if expDiff > mag {
		dp.significand.lo, dp.exp = 0, targetExp
		return
	}
	divisor := tenToThe[expDiff]
	dp.significand.lo = dp.significand.lo / divisor
	dp.exp = targetExp
	return
}

func (dp *decParts) unpack(d Decimal64) {
	dp.unpackV2(d, d.flavor())
}

func (dp *decParts) unpackV2(d Decimal64, fl flavor) {
	dp.original = d
	dp.sign = int(d.bits >> 63)
	dp.fl = fl
	switch fl {
	case flNormal53:
		// s EEeeeeeeee   (0)ttt tttttttttt tttttttttt tttttttttt tttttttttt tttttttttt
		//   EE ∈ {00, 01, 10}
		dp.exp = int((d.bits>>(63-10))&(1<<10-1)) - expOffset
		dp.significand.lo = d.bits & (1<<53 - 1)
		if dp.significand.lo == 0 {
			dp.exp = 0
		}
	case flNormal51:
		// s 11EEeeeeeeee (100)t tttttttttt tttttttttt tttttttttt tttttttttt tttttttttt
		//     EE ∈ {00, 01, 10}
		dp.exp = int((d.bits>>(63-12))&(1<<10-1)) - expOffset
		dp.significand.lo = d.bits&(1<<51-1) | (1 << 53)
	case flInf:
	default: // NaN
		dp.significand.lo = d.bits & (1<<51 - 1) // Payload
		return
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

var flavorLookup = []flavor{
	flNormal, flNormal, flNormal, flNormal,
	flNormal, flNormal, flNormal, flNormal,
	flNormal, flNormal, flNormal, flNormal,
	flNormal, flNormal, flNormal, flNormal,
	flNormal, flNormal, flNormal, flNormal,
	flNormal, flNormal, flNormal, flNormal,
	flNormal, flNormal, flNormal, flNormal,
	flNormal, flNormal, flNormal, flNormal,
	flNormal, flNormal, flNormal, flNormal,
	flNormal, flNormal, flNormal, flNormal,
	flNormal, flNormal, flNormal, flNormal,
	flNormal, flNormal, flNormal, flNormal,
	flNormal, flNormal, flNormal, flNormal,
	flNormal, flNormal, flNormal, flNormal,
	flNormal, flNormal, flNormal, flNormal,
	flInf, flInf, flQNaN, flSNaN,
}

func flav(d Decimal64) flavor {
	return flavorLookup[(d.bits>>(64-7))%uint64(len(flavorLookup))]
}
