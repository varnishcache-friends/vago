# vago
Go bindings for the Varnish API using cgo.

## Requirements
To build this package you will need:
- pkg-config
- libvarnishapi-dev >= 4.0.0

You will also need to set PKG_CONFIG_PATH to the directory where varnishapi.pc
is located before running `go get`. For example:
```
export PKG_CONFIG_PATH=/usr/local/lib/pkgconfig
```

## Installation
```
go get github.com/phenomenes/vago
```

## Example
```
// Sample code that mimics varnishlog -g raw
package main

import (
	"fmt"

	"github.com/phenomenes/vago"
)

func main() {
	// Open a Varnish Shared Memory file
	v, err := vago.Open("/var/lib/varnish/foobar")
	if err != nil {
		fmt.Println(err)
		return
	}
	v.Log("", vago.RAW, func(vxid uint32, tag string, _type string, data string) int {
		fmt.Printf("%10d %-14s %s %s\n", vxid, tag, _type, data)
		// -1  : Stop after it finds the first record
		// > 0 : Nothing to do but wait
		return 0
	})
	v.Close()
}
```
