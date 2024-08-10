package decimal

import (
	"encoding/gob"
)

var _ gob.GobDecoder = (*Decimal64)(nil)
var _ gob.GobEncoder = Zero64

// GobDecode implements encoding.GobDecoder.
func (d *Decimal64) GobDecode(buf []byte) error {
	return d.UnmarshalBinary(buf)
}

// GobEncode implements encoding.GobEncoder.
func (d Decimal64) GobEncode() ([]byte, error) {
	return d.MarshalBinary()
}
