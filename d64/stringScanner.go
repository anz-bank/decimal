package d64

import (
	"bytes"
	"fmt"
	"io"
	"unicode"
)

type runeScanner interface {
	io.Reader
	io.RuneScanner
}

type scanner struct {
	reader runeScanner
}

var _ fmt.ScanState = (*scanner)(nil)

func (s *scanner) ReadRune() (r rune, size int, err error) {
	return s.reader.ReadRune()
}

func (s *scanner) UnreadRune() error {
	return s.reader.UnreadRune()
}

func (s *scanner) SkipSpace() {
	for {
		ch, _, err := s.ReadRune()
		if err != nil {
			break
		}
		if !unicode.IsSpace(ch) {
			if err := s.UnreadRune(); err != nil {
				panic("s.UnreadRune() failed")
			}
			break
		}
	}
}

func (s *scanner) Token(skipSpace bool, f func(rune) bool) (token []byte, err error) {
	if skipSpace {
		s.SkipSpace()
	}

	var buf bytes.Buffer
	for {
		r, _, err := s.ReadRune()
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}
		if !f(r) {
			if err := s.UnreadRune(); err != nil {
				return nil, err
			}
			break
		}
		buf.WriteRune(r)
	}
	return buf.Bytes(), nil
}

func (s *scanner) Width() (wid int, ok bool) {
	return 0, false
}

func (s *scanner) Read(buf []byte) (n int, err error) {
	return s.reader.Read(buf)
}
