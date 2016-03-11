package vago

/*
#include <stdlib.h>
*/
import "C"

import (
	"sync"
	"unsafe"
)

type handleList struct {
	sync.RWMutex
	handles map[unsafe.Pointer]interface{}
}

func newHandle() *handleList {
	return &handleList{
		handles: make(map[unsafe.Pointer]interface{}),
	}
}

func (h *handleList) track(ptr interface{}) unsafe.Pointer {
	handle := C.malloc(1)
	h.Lock()
	h.handles[handle] = ptr
	h.Unlock()
	return handle
}

func (h *handleList) untrack(handle unsafe.Pointer) {
	h.Lock()
	delete(h.handles, handle)
	C.free(handle)
	h.Unlock()
}

func (h *handleList) get(handle unsafe.Pointer) interface{} {
	h.RLock()
	defer h.RUnlock()
	ptr, ok := h.handles[handle]
	if !ok {
		panic("invalid handle")
	}
	return ptr
}
