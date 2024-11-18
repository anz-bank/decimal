package decimal

import "testing"

func check(t *testing.T, ok bool) bool {
	t.Helper()
	if !ok {
		t.Errorf("expected true")
		return false
	}
	return true
}

func epsilon(t *testing.T, a, b float64) bool {
	t.Helper()
	if a/b-1 > 0.00000001 {
		t.Errorf("%f and %f too dissimilar", a, b)
		return false
	}
	return true
}

func equal[T comparable](t *testing.T, a, b T) bool {
	t.Helper()
	if a != b {
		t.Errorf("expected %+v, got %+v", a, b)
		return false
	}
	return true
}

func equalD64(t *testing.T, expected, actual Decimal64, fmtAndArgs ...any) {
	t.Helper()
	equal(t, expected.bits, actual.bits)
}

func isnil(t *testing.T, a any) bool {
	t.Helper()
	if a != nil {
		t.Errorf("expected nil, got %+v", a)
		return false
	}
	return true
}

func nopanic(t *testing.T, f func()) (b bool) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("panic: %+v", r)
			b = false
		}
	}()
	f()
	return true
}

func notequal[T comparable](t *testing.T, a, b T) bool {
	t.Helper()
	if a == b {
		t.Errorf("equal values %+v", a)
		return false
	}
	return true
}

func notnil(t *testing.T, a any) bool {
	t.Helper()
	if a == nil {
		t.Errorf("expected non-nil")
		return false
	}
	return true
}

func panics(t *testing.T, f func()) (b bool) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
			b = false
		}
	}()
	f()
	return true
}

func TestDiv10_64(t *testing.T) {
	t.Parallel()

	for i := uint64(0); i <= 10000; i++ {
		d := uint128T{i, 0}.divBy10().lo
		equal(t, i/10, d)
	}
}

func TestDiv10_64_po10(t *testing.T) {
	t.Parallel()

	for i, u := range tenToThe128 {
		var e uint128T
		if i > 0 {
			e = tenToThe128[i-1]
		}
		a := u.divBy10()
		equal(t, e, a)
	}
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
