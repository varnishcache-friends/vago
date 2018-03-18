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
	"fmt"
	"sync"
	"time"
	"unsafe"
)

const (
	// Grouping mode
	RAW  = C.VSL_g_raw
	VXID = C.VSL_g_vxid
	REQ  = C.VSL_g_request
	SESS = C.VSL_g_session
)

const (
	none tribool = iota
	Yes
	No
)

// A Varnish struct represents a handler for Varnish Shared Memory and
// Varnish Shared Log.
type Varnish struct {
	vsc         *C.struct_vsc
	vsm         *C.struct_vsm
	vsl         *C.struct_VSL_data
	vslq        *C.struct_VSLQ
	cursor      *C.struct_VSL_cursor
	mu          sync.Mutex
	closed      bool
	done        chan struct{}
	vslReattach bool
}

// Config parameters to connect to a Varnish instance.
type Config struct {
	// Path to Varnish Shared Memory file
	Path string
	// VSM connection timeout in milliseconds
	// -1 for no timeout
	Timeout time.Duration
	// Whether to reacquire the to the log
	// Values can be Yes or No. Default Yes
	VslReattach tribool
}

type tribool uint8

var ptrHandles *handleList

func init() {
	ptrHandles = newHandle()
}

// Open opens a Varnish Shared Memory file. If successful, returns a new
// Varnish.
func Open(c *Config) (*Varnish, error) {
	v := &Varnish{closed: true}
	v.vsm = C.VSM_New()
	if v.vsm == nil {
		return nil, errors.New(C.GoString(C.VSM_Error(v.vsm)))
	}
	v.vsc = C.VSC_New()
	if v.vsc == nil {
		defer v.Close()
		return nil, errors.New(C.GoString(C.VSM_Error(v.vsm)))
	}
	if c.Path != "" {
		cs := C.CString(c.Path)
		defer C.free(unsafe.Pointer(cs))
		arg := C.CString("n")
		defer C.free(unsafe.Pointer(arg))
		if C.VSM_Arg(v.vsm, *arg, cs) != 1 {
			defer v.Close()
			return nil, errors.New(C.GoString(C.VSM_Error(v.vsm)))
		}
	}
	var cs *C.char
	switch {
	case c.Timeout < 0:
		cs = C.CString("off")
	default:
		cs = C.CString(fmt.Sprintf("%d", c.Timeout/1000))
	}
	defer C.free(unsafe.Pointer(cs))
	arg := C.CString("t")
	defer C.free(unsafe.Pointer(arg))
	if C.VSM_Arg(v.vsm, *arg, cs) != 1 {
		defer v.Close()
		return nil, errors.New(C.GoString(C.VSM_Error(v.vsm)))
	}
	if C.VSM_Attach(v.vsm, -1) != 0 {
		defer v.Close()
		return nil, errors.New(C.GoString(C.VSM_Error(v.vsm)))
	}

	v.done = make(chan struct{})
	v.closed = false
	if c.VslReattach != No {
		v.vslReattach = true
	}

	return v, nil
}

func (v *Varnish) alive() bool {
	select {
	case <-v.done:
		return false
	default:
		return true
	}
}

// Stop stops processing Varnish events.
func (v *Varnish) Stop() {
	v.mu.Lock()
	defer v.mu.Unlock()

	if !v.closed {
		close(v.done)
		v.closed = true
	}
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
	if v.vsc != nil {
		C.VSC_Destroy(&v.vsc, v.vsm)
	}
	if v.vsm != nil {
		C.VSM_Destroy(&v.vsm)
	}
}
