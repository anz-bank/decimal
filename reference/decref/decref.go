package decref

/*
#cgo CXXFLAGS: -std=c++17 -I../decimal/include
#include "decref.h"
#include <stdlib.h>
*/
import "C"

import "unsafe"

type Dec64 C.Dec64

func Parse64(s string) Dec64 {
	cstr := C.CString(s)
	defer C.free(unsafe.Pointer(cstr))
	return Dec64(C.parse64(cstr))
}

func (d Dec64) String() string {
	cstr := C.string64(C.Dec64(d))
	defer C.free(unsafe.Pointer(cstr))
	return C.GoString(cstr)
}

func FromBid64(b uint64) Dec64 { return Dec64(C.frombid64(C.uint64_t(b))) }
func (d Dec64) ToBid() uint64  { return uint64(C.tobid64(C.Dec64(d))) }

func (d Dec64) IsNaN() bool { return C.isNaN64(C.Dec64(d)) != 0 }

func (d Dec64) Add(e Dec64) Dec64 { return Dec64(C.add64(C.Dec64(d), C.Dec64(e))) }
func (d Dec64) Sub(e Dec64) Dec64 { return Dec64(C.sub64(C.Dec64(d), C.Dec64(e))) }
func (d Dec64) Mul(e Dec64) Dec64 { return Dec64(C.mul64(C.Dec64(d), C.Dec64(e))) }
func (d Dec64) Quo(e Dec64) Dec64 { return Dec64(C.quo64(C.Dec64(d), C.Dec64(e))) }
