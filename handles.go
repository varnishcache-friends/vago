// The MIT License
//
// Copyright (c) 2013 The git2go contributors
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

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
