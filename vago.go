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

// A Varnish struct represents a handler for Varnish Shared Memory and
// Varnish Shared Log.
type Varnish struct {
	vsm    *C.struct_VSM_data
	vsl    *C.struct_VSL_data
	vslq   *C.struct_VSLQ
	cursor *C.struct_VSL_cursor
	mu     sync.Mutex
	closed bool
	done   chan struct{}
}

// Config parameters to connect to a Varnish instance.
type Config struct {
	Path    string        // Path to Varnish Shared Memory file
	Timeout time.Duration // VSM connection timeout in milliseconds
}

var ptrHandles *handleList

func init() {
	ptrHandles = newHandle()
}

// Open opens a Varnish Shared Memory file. If successful, returns a new
// Varnish.
func Open(c *Config) (*Varnish, error) {
	v := &Varnish{}
	v.vsm = C.VSM_New()
	if v.vsm == nil {
		return nil, errors.New(C.GoString(C.VSM_Error(v.vsm)))
	}
	if c.Path != "" {
		cs := C.CString(c.Path)
		defer C.free(unsafe.Pointer(cs))
		if C.VSM_n_Arg(v.vsm, cs) != 1 {
			return nil, errors.New(C.GoString(C.VSM_Error(v.vsm)))
		}
	}
	end := time.Now().Add(c.Timeout * time.Millisecond)
	for {
		if C.VSM_Open(v.vsm) >= 0 {
			break
		}
		if c.Timeout <= 0 || time.Now().After(end) {
			return nil, errors.New(C.GoString(C.VSM_Error(v.vsm)))
		}
		C.VSM_ResetError(v.vsm)
		time.Sleep(500 * time.Millisecond)
	}

	v.done = make(chan struct{})
	v.closed = false

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
	if v.vsm != nil {
		C.VSM_Delete(v.vsm)
	}
}
