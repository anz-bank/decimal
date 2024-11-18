package decimal

import (
	"strings"
	"testing"
	"unicode"
)

func TestStringScannerSkipSpace(t *testing.T) {
	t.Parallel()

	state := &scanner{reader: strings.NewReader(" \tx")}

	state.SkipSpace()

	r, size, err := state.ReadRune()
	isnil(t, err)
	equal(t, 1, size)
	equal(t, 'x', r)

	nopanic(t, func() { state.SkipSpace() })
}

func TestStringScannerTokenSkipSpace(t *testing.T) {
	t.Parallel()

	state := &scanner{reader: strings.NewReader(" \txyz")}

	token, err := state.Token(false, unicode.IsLetter)
	isnil(t, err)
	equal(t, 0, len(token))

	token, err = state.Token(true, unicode.IsLetter)
	isnil(t, err)
	equal(t, "xyz", string(token))
}

func TestStringScannerRead(t *testing.T) {
	t.Parallel()

	state := &scanner{reader: strings.NewReader("hello world!")}

	var hello [5]byte
	var world [10]byte

	n, err := state.Read(hello[:])
	isnil(t, err)
	equal(t, 5, n)
	equal(t, "hello", string(hello[:]))

	state.SkipSpace()

	n, err = state.Read(world[:])
	isnil(t, err)
	equal(t, 6, n)
	equal(t, "world!", string(world[:n]))
}
