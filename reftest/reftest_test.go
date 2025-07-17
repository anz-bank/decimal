package reftest_test

import (
	"fmt"
	"math/rand"
	"testing"
	"unsafe"

	"github.com/anz-bank/decimal/d64"
	. "github.com/anz-bank/decimal/reftest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestString(t *testing.T) {
	assert.Equal(t, "+123E-2", New64("1.23").String())
	assert.Equal(t, "+456E-2", New64("4.56").String())
	assert.Equal(t, "-4560E-2", New64("-45.60").String())
	assert.NotEqual(t, "-4560E-2", New64("-45.6").String())
	assert.Equal(t, "+Inf", New64("inf").String())
	assert.Equal(t, "+NaN", New64("nan").String())
}

func TestAdd(t *testing.T) {
	assert.Equal(t, "+579E-2", New64("1.23").Add(New64("4.56")).String())
	assert.Equal(t, "+580E-2", New64("1.23").Add(New64("4.57")).String())
	assert.Equal(t, "+580E-2", New64("1.24").Add(New64("4.56")).String())
}

func TestMul(t *testing.T) {
	assert.Equal(t, "+56088E-4", New64("1.23").Mul(New64("4.56")).String())
	assert.Equal(t, "+55510E-4", New64("1.22").Mul(New64("4.55")).String())
	assert.Equal(t, "+56500E-4", New64("1.25").Mul(New64("4.52")).String())
	assert.Equal(t, "+54900E-4", New64("1.22").Mul(New64("4.50")).String())
	assert.Equal(t, "+5490E-3", New64("1.22").Mul(New64("4.5")).String())
}

func TestD64Add(t *testing.T) {
	t.Parallel()

	r := rand.New(rand.NewSource(0))
	for i := range 100 {
		x1 := randomDec64(r)
		y1 := randomDec64(r)
		expected := x1.Add(y1)
		x2 := fromDec64(x1)
		y2 := fromDec64(y1)
		actual := x2.Add(y2)
		require.Equal(t, expected, toDec64(actual), "i = %d\nexpected: %v (%s) + %v (%s) = %v (%s)\nactual:   %v (%s) + %v (%s) = %v (%s %v)",
			i, x1, dec64Raw(x1), y1, dec64Raw(y1), expected, dec64Raw(expected), x2, d64Raw(x2), y2, d64Raw(y2), actual, d64Raw(actual), toDec64(actual))
	}
}

func TestD64AddOneOff(t *testing.T) {
	t.Parallel()

	x1 := Dec64(0b0001100010100001110101000000111110000101010100010100011000111001)
	y1 := Dec64(0b1101001111100001111011001000010111100010001011101000100101111000)
	expected := x1.Add(y1)
	x2 := fromDec64(x1)
	y2 := fromDec64(y1)
	actual := x2.Add(y2)
	require.Equal(t, expected, toDec64(actual), "expected: %v (%s) + %v (%s) = %v (%s)\nactual:   %v (%s) + %v (%s) = %v (%s %v)",
		x1, dec64Raw(x1), y1, dec64Raw(y1), expected, dec64Raw(expected), x2, d64Raw(x2), y2, d64Raw(y2), actual, d64Raw(actual), toDec64(actual))
}

func fromDec64(bits Dec64) d64.Decimal {
	var d d64.Decimal
	*(*Dec64)(unsafe.Pointer(&d)) = bits
	return d
}

func toDec64(x d64.Decimal) Dec64 {
	return *(*Dec64)(unsafe.Pointer(&x))
}

func dec64Raw(x Dec64) string {
	u := *(*uint64)(unsafe.Pointer(&x))
	return fmt.Sprintf("%01b:%013b:%050b", u>>63%(1<<1), u>>50%(1<<13), u%(1<<50))
}

func d64Raw(x d64.Decimal) string {
	u := *(*uint64)(unsafe.Pointer(&x))
	return fmt.Sprintf("%01b:%013b:%050b", u>>63, u>>50%(1<<13), u%(1<<50))
}

func randomDec64(r *rand.Rand) Dec64 {
loop:
	for {
		u := r.Uint64()
		// log.Printf("%016x", u)
		switch (u >> 59) & 0b1111 {
		case 0b1100, 0b1101, 0b1110:
			if 1<<53+u%(1<<51) > 10_000_000_000_000_000 {
				continue loop
			}
		}
		return Dec64(u)
	}
}

// expected: -8252036246683216E+311  (1:1011000101111:01010100010010111011101001001010000011111001010000) + +SNaN    (0:1111110001111:01000000101011100101010011010010110111110101111111) = +NaN.    (0:1111100000000:01000000101011100101010011010010110111110101111111)
// actual:   -8.252036246683216e+326 (1:1011000101111:01010100010010111011101001001010000011111001010000) + NaN�4847 (0:1111110001111:01000000101011100101010011010010110111110101111111) = NaN�2223 (0:1111110000000:01000000101011100101010011010010110111110101111111 +SNaN)
