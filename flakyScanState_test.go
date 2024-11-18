package decimal

import "fmt"

type flakyScanState struct {
	actual fmt.ScanState
	offset int
	failAt int
}

func (s *flakyScanState) ReadRune() (r rune, size int, err error) {
	r, size, err = s.actual.ReadRune()
	err = s.failNow(size, err)
	return
}

func (s *flakyScanState) UnreadRune() error {
	return s.actual.UnreadRune()
}

func (s *flakyScanState) SkipSpace() {
	s.actual.SkipSpace()
}

func (s *flakyScanState) Token(skipSpace bool, f func(rune) bool) (token []byte, err error) {
	token, err = s.actual.Token(skipSpace, f)
	err = s.failNow(len(token), err)
	return
}

func (s *flakyScanState) Width() (wid int, ok bool) {
	return s.actual.Width()
}

func (s *flakyScanState) Read(buf []byte) (n int, err error) {
	err = s.failNow(s.actual.Read(buf))
	return
}

func (s *flakyScanState) failNow(size int, err error) error {
	if err != nil {
		return err
	}
	s.offset += size
	if s.offset > s.failAt {
		return fmt.Errorf("flakyScanState read failed")
	}
	return nil
}
