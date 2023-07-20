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
	"errors"
	"time"
	"unsafe"
)

const (
	lenmask       = 0xffff
	clientmarker  = uint32(1) << 30
	backendmarker = uint32(1) << 31
	identmask0    = ^(uint32(3) << 30)
	identmask1    = ^(uint64(0b1111111111111) << 51)
	vermask       = uint32(3) << 16
	// Cursor options
	COPT_TAIL     = 1 << 0
	COPT_BATCH    = 1 << 1
	COPT_TAILSTOP = 1 << 2
	// VSM status bitmap
	vsmWrkRestarted = 1 << 11
	// VSL_Dispatch errors
	unknownVersion = 2
)

var (
	ErrAbandoned      = errors.New("log abandoned")
	ErrOverrun        = errors.New("log overrun")
	ErrUnknownVersion = errors.New("log version unknown")
)

type ErrVSL string

func (e ErrVSL) Error() string { return string(e) }

// LogCallback defines a callback function.
// It's used by Log.
type LogCallback func(vxid uint64, tag, _type, data string) int

// Log calls the given callback for any transactions matching the query
// and grouping.
func (v *Varnish) Log(query string, grouping uint32, copt uint, logCallback LogCallback) error {
	v.vsl = C.VSL_New()
	handle := ptrHandles.track(logCallback)
	defer ptrHandles.untrack(handle)
	if grouping > max {
		grouping = VXID
	}
	if query != "" {
		cs := C.CString(query)
		defer C.free(unsafe.Pointer(cs))
		v.vslq = C.VSLQ_New(v.vsl, nil, grouping, cs)
	} else {
		v.vslq = C.VSLQ_New(v.vsl, nil, grouping, nil)
	}
	if v.vslq == nil {
		return ErrVSL(C.GoString(C.VSL_Error(v.vsl)))
	}
	hasCursor := -1
DispatchLoop:
	for v.alive() {
		if v.vsm != nil && (C.VSM_Status(v.vsm)&vsmWrkRestarted) != 0 {
			if hasCursor < 1 {
				C.VSLQ_SetCursor(v.vslq, nil)
				hasCursor = 0
			}
		}
		if v.vsm != nil && hasCursor < 1 {
			// Reconnect VSM
			v.cursor = C.VSL_CursorVSM(v.vsl, v.vsm, C.uint(copt))
			if v.cursor == nil {
				C.VSL_ResetError(v.vsl)
				continue
			}
			hasCursor = 1
			C.VSLQ_SetCursor(v.vslq, &v.cursor)
		}
		i := C.VSLQ_Dispatch(v.vslq,
			(*C.VSLQ_dispatch_f)(C.dispatchCallback),
			handle)
		switch i {
		case unknownVersion:
			// Unknown version
			return ErrUnknownVersion
		case 1:
			// Call again
			continue
		case 0:
			// Nothing to do but wait
			time.Sleep(10 * time.Millisecond)
			continue
		case -1:
			// EOF
			break DispatchLoop
		case -2:
			// Abandoned
			if !v.vslReattach {
				return ErrAbandoned
			}
			// Re-acquire the log cursor
			C.VSLQ_SetCursor(v.vslq, nil)
			hasCursor = 0
		default:
			// Overrun
			return ErrOverrun
		}
	}

	return nil
}

// dispatchCallback walks through the transaction and calls a function of
// type LogCallback.
//
//export dispatchCallback
func dispatchCallback(vsl *C.struct_VSL_data, pt **C.struct_VSL_transaction, handle unsafe.Pointer) C.int {
	tx := uintptr(unsafe.Pointer(pt))
	var _type string
	logCallback := ptrHandles.get(handle)
	for {
		if tx == 0 {
			break
		}
		t := (**C.struct_VSL_transaction)(unsafe.Pointer(tx))
		if *t == nil {
			break
		}
		for {
			i := C.VSL_Next((*t).c)
			if i < 0 {
				return C.int(i)
			}
			if i == 0 {
				break
			}
			if C.VSL_Match(vsl, (*t).c) == 0 {
				continue
			}
			h1 := uint32(*(*t).c.rec.ptr)
			tag := C.GoString(C.VSL_tags[h1>>24])
			length := C.int(h1 & lenmask)
			ver := C.int((h1 & vermask) >> 16)
			pHeader := uintptr(unsafe.Pointer((*t).c.rec.ptr))
			stride := unsafe.Sizeof(C.uint32_t(0))
			var pData uintptr
			var vxid uint64
			switch ver {
			case 0: // Varnish < 7.3.0
				h2 := uint32(*(*C.uint32_t)(unsafe.Pointer(pHeader + stride)))
				pData = pHeader + 2*stride
				vxid = uint64(h2 & identmask0)
				switch {
				case h2&clientmarker != 0:
					_type = "c"
				case h2&backendmarker != 0:
					_type = "b"
				default:
					_type = "-"
				}
			case 1: // Varnish >= 7.3.0
				h2 := uint32(*(*C.uint32_t)(unsafe.Pointer(pHeader + stride)))
				h3 := uint32(*(*C.uint32_t)(unsafe.Pointer(pHeader + 2*stride)))
				pData = pHeader + 3*stride
				vxid = (uint64(h3)<<32 | uint64(h2)) & identmask1
				switch {
				case h3&clientmarker != 0:
					_type = "c"
				case h3&backendmarker != 0:
					_type = "b"
				default:
					_type = "-"
				}
			default: // Newer Varnish version we are not aware of: fail.
				return C.int(unknownVersion)
			}
			data := C.GoStringN((*C.char)(unsafe.Pointer(pData)), length-1)
			ret := logCallback.(LogCallback)(vxid, tag, _type, data)
			if ret != 0 {
				return C.int(ret)
			}
		}
		tx += unsafe.Sizeof(t)
	}
	return 0
}
