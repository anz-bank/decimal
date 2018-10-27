package decimal

import (
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/require"
)

func TestStringScannerSkipSpace(t *testing.T) {
	require := require.New(t)

	state := &stringScanner{reader: strings.NewReader(" \tx")}

	state.SkipSpace()

	r, size, err := state.ReadRune()
	require.NoError(err)
	require.Equal(1, size)
	require.Equal('x', r)

	require.NotPanics(func() { state.SkipSpace() })
}

func TestStringScannerTokenSkipSpace(t *testing.T) {
	require := require.New(t)

	state := &stringScanner{reader: strings.NewReader(" \txyz")}

	token, err := state.Token(false, unicode.IsLetter)
	require.NoError(err)
	require.Equal([]byte(nil), token)

	token, err = state.Token(true, unicode.IsLetter)
	require.NoError(err)
	require.Equal([]byte("xyz"), token)
}

func TestStringScannerWidth(t *testing.T) {
	require := require.New(t)

	state := &stringScanner{reader: strings.NewReader("foo"), wid: 0, widSet: false}
	wid, ok := state.Width()
	require.False(ok)

	state = &stringScanner{reader: strings.NewReader("foo"), wid: 10, widSet: true}
	wid, ok = state.Width()
	require.True(ok)
	require.Equal(10, wid)
}

func TestStringScannerRead(t *testing.T) {
	require := require.New(t)

	state := &stringScanner{reader: strings.NewReader("hello world!")}

	hello := make([]byte, 5)
	world := make([]byte, 10)

	n, err := state.Read(hello)
	require.NoError(err)
	require.Equal(5, n)
	require.Equal([]byte("hello"), hello)

	state.SkipSpace()

	n, err = state.Read(world)
	require.NoError(err)
	require.Equal(6, n)
	require.Equal([]byte("world!"), world[:n])
}
