package decref_test

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/anz-bank/decimal/d64"
	"github.com/anz-bank/decimal/reference/decref"
)

func TestParseString(t *testing.T) {
	t.Parallel()

	d := decref.Parse64("3.142")
	if d.String() != "3.142" {
		t.Errorf("expected 3.142, got %s", d.String())
	}
}

func TestAdd(t *testing.T) {
	t.Parallel()

	rng := rand.New(rand.NewPCG(0, 0))
	for i := range 1000000 {
		decrefA := randDec(i, rng)
		decrefB := randDec(i, rng)

		aStr := formatDec(decrefA)
		bStr := formatDec(decrefB)

		d64A := d64.MustParse(aStr)
		d64B := d64.MustParse(bStr)

		// Perform addition
		decrefResult := decrefA.Add(decrefB)
		d64Result := d64A.Add(d64B)

		d64Str := d64Result.Text('e', -1)
		decrefStr := formatDec(decrefResult)

		if d64Str != decrefStr {
			t.Errorf("Mismatch: d64Result=%s, decrefResult=%s for a=%s, b=%s, %s", d64Str, decrefStr, aStr, bStr, decrefResult.String())
			t.FailNow()
		}
	}
}

func formatDec(d decref.Dec64) string {
	decrefStr := d.String()
	// Trim mantissa trailing zeros.
	if a, b, cut := strings.Cut(decrefStr, "0e"); cut {
		decrefStr = strings.TrimRight(a, "0") + "e" + b
	}
	// Trim exponent leading zeros and all-zeros.
	if a, b, cut := strings.Cut(decrefStr, "+0"); cut {
		decrefStr = a[:len(a)-1]
		if exp := strings.TrimLeft(b, "0"); exp != "" {
			decrefStr += "e+" + exp
		}
	}
	// Trim negative-exponent leading zeros.
	if a, b, cut := strings.Cut(decrefStr, "-0"); cut {
		decrefStr = a + "-" + strings.TrimLeft(b, "0")
	}
	return decrefStr
}

func randDec(i int, rng *rand.Rand) decref.Dec64 {
	sign := [2]string{"", "-"}[rng.IntN(2)]
	const explimit = 385
	exp := rng.IntN(2*explimit+1) - (explimit - 1)
	hidigit := rng.IntN(10)
	lodigits := rng.Int64N(1_000_000_000_000_000)
	s := fmt.Sprintf("%s%d.%dE%d", sign, hidigit, lodigits, exp)
	d := decref.Parse64(s)
	if d.IsNaN() {
		panic(fmt.Errorf("[%d] unparsable: %s (d = %s)", i, s, d))
	}
	return d
}
