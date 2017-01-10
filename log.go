// Package vago

package vago

/*
#cgo pkg-config: varnishapi
#cgo LDFLAGS: -lvarnishapi -lm
#include <sys/types.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <vapi/vsc.h>
#include <vapi/vsm.h>
#include <vapi/vsl.h>

int dispatchCallback(struct VSL_data *vsl, struct VSL_transaction **trans, void *priv);
*/
import "C"

import (
	"encoding/binary"
	"errors"
	"time"
	"unsafe"
)

const (
	lenmask       = 0xffff
	clientmarker  = uint32(1) << 30
	backendmarker = uint32(1) << 31
	identmask     = ^(uint32(3) << 30)
)

// LogCallback defines a callback function.
// It's used by Log.
type LogCallback func(vxid uint32, tag, _type, data string) int

// Log calls the given callback for any transactions matching the query
// and grouping.
func (v *Varnish) Log(query string, grouping uint32, logCallback LogCallback) error {
	v.vsl = C.VSL_New()
	handle := ptrHandles.track(logCallback)
	defer ptrHandles.untrack(handle)
	for {
		v.cursor = C.VSL_CursorVSM(v.vsl, v.vsm, 1)
		if v.cursor != nil {
			break
		}
	}
	if grouping < 0 || grouping > 4 {
		grouping = VXID
	}
	if query != "" {
		cs := C.CString(query)
		defer C.free(unsafe.Pointer(cs))
		v.vslq = C.VSLQ_New(v.vsl, &v.cursor, grouping, cs)
	} else {
		v.vslq = C.VSLQ_New(v.vsl, &v.cursor, grouping, nil)
	}
	if v.vslq == nil {
		return errors.New(C.GoString(C.VSL_Error(v.vsl)))
	}
	for v.alive {
		i := C.VSLQ_Dispatch(v.vslq,
			(*C.VSLQ_dispatch_f)(unsafe.Pointer(C.dispatchCallback)),
			handle)
		if i == 1 {
			continue
		}
		if i == 0 {
			time.Sleep(1000)
			continue
		}
		if i == -1 {
			break
		}
	}
	return nil
}

// dispatchCallback walks through the transaction and calls a function of
// type LogCallback.
//export dispatchCallback
func dispatchCallback(vsl *C.struct_VSL_data, pt **C.struct_VSL_transaction, handle unsafe.Pointer) C.int {
	var tx = uintptr(unsafe.Pointer(pt))
	var _type string
	logCallback := ptrHandles.get(handle)
	for {
		if tx == 0 {
			break
		}
		t := ((**C.struct_VSL_transaction)(unsafe.Pointer(tx)))
		if *t == nil {
			break
		}
		for {
			i := C.VSL_Next((*t).c)
			if i < 0 {
				return i
			}
			if i == 0 {
				break
			}
			if C.VSL_Match(vsl, (*t).c) == 0 {
				continue
			}
			s1 := cui32tosl((*t).c.rec.ptr, 8)
			tag := C.GoString(C.VSL_tags[s1[0]>>24])
			vxid := s1[1] & identmask
			length := C.int(s1[0] & lenmask)
			switch {
			case s1[1]&(clientmarker) != 0:
				_type = "c"
			case s1[1]&(backendmarker) != 0:
				_type = "b"
			default:
				_type = "-"
			}
			s2 := cui32tosl((*t).c.rec.ptr, (length+2)*4)
			data := ui32tostr(&s2[2], length)
			ret := logCallback.(LogCallback)(vxid, tag, _type, data)
			if ret != 0 {
				return C.int(ret)
			}
		}
		tx += unsafe.Sizeof(t)
	}
	return 0
}

// Convert C.uint32_t to slice of uint32
func cui32tosl(ptr *C.uint32_t, length C.int) []uint32 {
	b := C.GoBytes(unsafe.Pointer(ptr), length)
	s := make([]uint32, length/4)
	for i := range s {
		s[i] = uint32(binary.LittleEndian.Uint32(b[i*4 : (i+1)*4]))
	}
	return s
}

// Convert uint32 to string
func ui32tostr(val *uint32, length C.int) string {
	return C.GoStringN((*C.char)(unsafe.Pointer(val)), length-1)
}
