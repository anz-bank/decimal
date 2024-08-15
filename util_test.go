package decimal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiv10_64(t *testing.T) {
	for i := uint64(0); i <= 10000; i++ {
		d := uint128T{i, 0}.divBy10().lo
		require.EqualValues(t, i/10, d, "%d/10 ≠ %d (expecting %d)", i, d, i/10)
	}
}

func TestDiv10_64_po10(t *testing.T) {
	for i, u := range tenToThe128 {
		var e uint128T
		if i > 0 {
			e = tenToThe128[i-1]
		}
		a := u.divBy10()
		require.EqualValues(t, e, a, "%v/10 ≠ %v (expecting %d)", u, a, e)
	}
}

func TestUmul64_po10(t *testing.T) {
	for i, u := range tenToThe128 {
		if u.hi == 0 {
			for j, v := range tenToThe128 {
				if v.hi == 0 {
					e := tenToThe128[i+j]
					a := umul64(u.lo, v.lo)
					require.EqualValues(t, e, a, "%v/10 ≠ %v (expecting %d)", u, a, e)
				}
			}
		}
	}
}
