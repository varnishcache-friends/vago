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
*/
import "C"

import (
	"fmt"
	"unsafe"
)

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
