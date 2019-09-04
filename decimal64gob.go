package decimal

import (
	"encoding/binary"
)

// GobDecode implements encoding.GobDecoder.
func (d *Decimal64) GobDecode(buf []byte) error {
	d.bits = binary.BigEndian.Uint64(buf)
	// TODO: Check for out of bounds significand.
	return nil
}

// GobEncode implements encoding.GobEncoder.
func (d Decimal64) GobEncode() ([]byte, error) {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, d.bits)
	return buf, nil
}
