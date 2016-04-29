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

int listCallback(void *priv, struct VSC_point *);
int dispatchCallback(struct VSL_data *vsl, struct VSL_transaction **trans, void *priv);
*/
import "C"

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"
	"unsafe"
)

const (
	// Grouping mode
	RAW  = C.VSL_g_raw
	VXID = C.VSL_g_vxid
	REQ  = C.VSL_g_request
	SESS = C.VSL_g_session

	lenmask       = 0xffff
	clientmarker  = uint32(1) << 30
	backendmarker = uint32(1) << 31
	identmask     = ^(uint32(3) << 30)
)

var ptrHandles *handleList

func init() {
	ptrHandles = newHandle()
}

// A Varnish struct represents a handler for Varnish Shared Memory and
// Varnish Shared Log.
type Varnish struct {
	vsm    *C.struct_VSM_data
	vsl    *C.struct_VSL_data
	vslq   *C.struct_VSLQ
	cursor *C.struct_VSL_cursor
}

// Open opens a Varnish Shared Memory file. If successful, returns a new
// Varnish.
func Open(path string) (*Varnish, error) {
	v := Varnish{}
	v.vsm = C.VSM_New()
	if v.vsm == nil {
		return nil, errors.New(C.GoString(C.VSM_Error(v.vsm)))
	}
	if path != "" {
		cs := C.CString(path)
		defer C.free(unsafe.Pointer(cs))
		if C.VSM_n_Arg(v.vsm, cs) != 1 {
			return nil, errors.New(C.GoString(C.VSM_Error(v.vsm)))
		}
	}
	if C.VSM_Open(v.vsm) < 0 {
		return nil, errors.New(C.GoString(C.VSM_Error(v.vsm)))
	}
	return &v, nil
}

// Close closes and unmaps the Varnish Shared Memory.
func (v *Varnish) Close() {
	if v.vslq != nil {
		C.VSLQ_Delete(&v.vslq)
	}
	if v.vsl != nil {
		C.VSL_Delete(v.vsl)
	}
	if v.vsm != nil {
		C.VSM_Delete(v.vsm)
	}
}

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
	for {
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

			// ptr is an uint32_t pointer array, we use GoBytes to
			// back it in a Go byte slice to retrieve its 32 bits
			// elements.
			b := C.GoBytes(unsafe.Pointer((*t).c.rec.ptr), 8)
			s := make([]uint32, 2)
			for i := range s {
				s[i] = uint32(binary.LittleEndian.Uint32(b[i*4 : (i+1)*4]))
			}
			tag := C.GoString(C.VSL_tags[s[0]>>24])
			vxid := s[1] & identmask
			_type := "-"
			if s[1]&(clientmarker) != 0 {
				_type = "c"
			} else if s[1]&(backendmarker) != 0 {
				_type = "b"
			}
			lenght := C.int(s[0] & lenmask)
			u32 := cui32tosl((*t).c.rec.ptr, (lenght+2)*4)
			data := ui32tostr(&u32[2], lenght)
			ret := logCallback.(LogCallback)(vxid, tag, _type, data)
			if ret != 0 {
				return C.int(ret)
			}
		}
		tx += unsafe.Sizeof(t)
	}
	return 0
}

// Stats returns a map with all stat counters and their values.
func (v *Varnish) Stats() map[string]uint64 {
	items := make(map[string]uint64)
	handle := ptrHandles.track(items)
	defer ptrHandles.untrack(handle)
	C.VSC_Iter(v.vsm, nil,
		(*C.VSC_iter_f)(unsafe.Pointer(C.listCallback)),
		handle)
	return items
}

// Stat takes a Varnish stat field and returns its value and true if found,
// 0 and false otherwise.
func (v *Varnish) Stat(s string) (uint64, bool) {
	stats := v.Stats()
	value, ok := stats[s]
	return value, ok
}

//export listCallback
func listCallback(handle unsafe.Pointer, pt *C.struct_VSC_point) C.int {
	priv := ptrHandles.get(handle)
	var name string
	if pt == nil || priv == nil {
		return 1
	}
	_type := C.GoString(&pt.section.fantom._type[0])
	ident := C.GoString(&pt.section.fantom.ident[0])
	field := C.GoString(pt.desc.name)
	value := *(*uint64)(unsafe.Pointer(pt.ptr))
	s := fmt.Sprint(_type)
	if ident != "" {
		name = fmt.Sprint(s, ".", ident, ".", field)
	} else {
		name = fmt.Sprint(s, ".", field)
	}
	items, ok := priv.(map[string]uint64)
	if !ok {
		return 1
	}
	items[name] = value
	return 0
}

// Convert uint32 to string
func ui32tostr(val *uint32, lenght C.int) string {
	return C.GoStringN((*C.char)(unsafe.Pointer(val)), lenght)
}

// Convert C.uint32_t to slice of uint32
func cui32tosl(ptr *C.uint32_t, lenght C.int) []uint32 {
	b := C.GoBytes(unsafe.Pointer(ptr), lenght)
	s := make([]uint32, lenght/4)
	for i := range s {
		s[i] = uint32(binary.LittleEndian.Uint32(b[i*4 : (i+1)*4]))
	}
	return s
}
