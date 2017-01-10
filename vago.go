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
*/
import "C"

import (
	"errors"
	"unsafe"
)

const (
	// Grouping mode
	RAW  = C.VSL_g_raw
	VXID = C.VSL_g_vxid
	REQ  = C.VSL_g_request
	SESS = C.VSL_g_session
)

// A Varnish struct represents a handler for Varnish Shared Memory and
// Varnish Shared Log.
type Varnish struct {
	vsm    *C.struct_VSM_data
	vsl    *C.struct_VSL_data
	vslq   *C.struct_VSLQ
	cursor *C.struct_VSL_cursor
	alive  bool
}

var ptrHandles *handleList

func init() {
	ptrHandles = newHandle()
}

// Open opens a Varnish Shared Memory file. If successful, returns a new
// Varnish.
func Open(path string) (*Varnish, error) {
	v := &Varnish{}
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
	v.alive = true
	return v, nil
}

// Stop stops processing Varnish events.
func (v *Varnish) Stop() {
	v.alive = false
}

// Close closes and unmaps the Varnish Shared Memory.
func (v *Varnish) Close() {
	v.Stop()
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
