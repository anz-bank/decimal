package decimal

import (
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/require"
)

func TestStringScannerSkipSpace(t *testing.T) {
	require := require.New(t)

	state := &scanner{reader: strings.NewReader(" \tx")}

	state.SkipSpace()

	r, size, err := state.ReadRune()
	require.NoError(err)
	require.Equal(1, size)
	require.Equal('x', r)

	require.NotPanics(func() { state.SkipSpace() })
}

func TestStringScannerTokenSkipSpace(t *testing.T) {
	require := require.New(t)

	state := &scanner{reader: strings.NewReader(" \txyz")}

	token, err := state.Token(false, unicode.IsLetter)
	require.NoError(err)
	require.Equal([]byte(nil), token)

	token, err = state.Token(true, unicode.IsLetter)
	require.NoError(err)
	require.Equal([]byte("xyz"), token)
}

func TestStringScannerRead(t *testing.T) {
	require := require.New(t)

	state := &scanner{reader: strings.NewReader("hello world!")}

	var hello [5]byte
	var world [10]byte

	n, err := state.Read(hello[:])
	require.NoError(err)
	require.Equal(5, n)
	require.Equal([]byte("hello"), hello[:])

	state.SkipSpace()

	n, err = state.Read(world[:])
	require.NoError(err)
	require.Equal(6, n)
	require.Equal([]byte("world!"), world[:n])
}
