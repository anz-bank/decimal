package reftest

/*
#include <stdlib.h>
#include "bid_conf.h"
#include "bid_functions.h"
*/
import "C"
import (
	"bytes"
	"unsafe"
)

type Dec64 uint64

func New64(s string) Dec64 {
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	return Dec64(C.__bid64_from_string(cs, C._IDEC_round(0), nil))
}

func (x Dec64) Add(y Dec64) Dec64 {
	return Dec64(C.__bid64_add(C.BID_UINT64(x), C.BID_UINT64(y), C._IDEC_round(0), nil))
}

func (x Dec64) Mul(y Dec64) Dec64 {
	return Dec64(C.__bid64_mul(C.BID_UINT64(x), C.BID_UINT64(y), C._IDEC_round(0), nil))
}

func (x Dec64) String() string {
	var buf [64]C.char
	p := &buf[0]
	C.__bid64_to_string(p, C.BID_UINT64(x), nil)
	data := C.GoBytes(unsafe.Pointer(p), C.int(len(buf)))
	return string(data[:bytes.IndexByte(data, 0)])
}
