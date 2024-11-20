package decimal

import "testing"

func errorf(t *testing.T, format string, args ...any) {
	t.Helper()
	t.Errorf(format, args...)
}

func repeatOnFail(t *testing.T, f func()) {
	t.Helper()
	alreadyFailed := t.Failed()
	f()
	if !alreadyFailed && t.Failed() {
		f()
	}
}

func check(t *testing.T, ok bool) bool {
	t.Helper()
	if !ok {
		errorf(t, "expected true")
		return false
	}
	return true
}

func epsilon(t *testing.T, a, b float64) bool {
	t.Helper()
	if a/b-1 > 0.00000001 {
		errorf(t, "%f and %f too dissimilar", a, b)
		return false
	}
	return true
}

func equal[T comparable](t *testing.T, a, b T) bool {
	t.Helper()
	if a != b {
		errorf(t, "expected %+v, got %+v", a, b)
		return false
	}
	return true
}

func equalD64(t *testing.T, expected, actual Decimal64) {
	t.Helper()
	equal(t, expected.bits, actual.bits)
}

func isnil(t *testing.T, a any) bool {
	t.Helper()
	if a != nil {
		errorf(t, "expected nil, got %+v", a)
		return false
	}
	return true
}

func nopanic(t *testing.T, f func()) (b bool) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			errorf(t, "panic: %+v", r)
			b = false
		}
	}()
	f()
	return true
}

func notequal[T comparable](t *testing.T, a, b T) bool {
	t.Helper()
	if a == b {
		errorf(t, "equal values %+v", a)
		return false
	}
	return true
}

func notnil(t *testing.T, a any) bool {
	t.Helper()
	if a == nil {
		errorf(t, "expected non-nil")
		return false
	}
	return true
}

func panics(t *testing.T, f func()) (b bool) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			errorf(t, "expected panic")
			b = false
		}
	}()
	f()
	return true
}

func TestUmul64_po10(t *testing.T) {
	t.Parallel()

	for i, u := range tenToThe128 {
		if u.hi == 0 {
			for j, v := range tenToThe128 {
				if v.hi == 0 {
					e := tenToThe128[i+j]
					a := umul64(u.lo, v.lo)
					equal(t, e, a)
				}
			}
		}
	}
}
