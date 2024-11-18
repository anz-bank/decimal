package decimal

import "testing"

func TestPi64(t *testing.T) {
	t.Parallel()

	equal(t, "3.141592653589793", Pi64.String())
}

func TestE64(t *testing.T) {
	t.Parallel()

	equal(t, "2.718281828459045", E64.String())
}
