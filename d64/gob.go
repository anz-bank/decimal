package d64

import (
	"encoding/gob"
)

var _ gob.GobDecoder = (*Decimal)(nil)
var _ gob.GobEncoder = Zero

// GobDecode implements encoding.GobDecoder.
func (d *Decimal) GobDecode(buf []byte) error {
	return d.UnmarshalBinary(buf)
}

// GobEncode implements encoding.GobEncoder.
func (d Decimal) GobEncode() ([]byte, error) {
	return d.MarshalBinary()
}
