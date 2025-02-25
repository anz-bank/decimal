package d64

import "testing"

func TestPi(t *testing.T) {
	t.Parallel()

	equal(t, "3.141592653589793", Pi.String())
}

func TestE(t *testing.T) {
	t.Parallel()

	equal(t, "2.718281828459045", E.String())
}
