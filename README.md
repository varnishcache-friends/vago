[![Build Status](https://travis-ci.org/phenomenes/vago.svg?branch=master)](https://travis-ci.org/phenomenes/vago)

# vago
Go bindings for the Varnish API using cgo.

## Requirements

To build this package you will need:
- pkg-config
- libvarnishapi-dev >= 5.2.0

You will also need to set PKG_CONFIG_PATH to the directory where `varnishapi.pc`
is located before running `go get`. For example:

```
export PKG_CONFIG_PATH=/usr/local/lib/pkgconfig
```

## Installation

```
go get github.com/phenomenes/vago
```

## Examples

Same as running `varnishlog -g raw`:

```go
package main

import (
	"fmt"

	"github.com/phenomenes/vago"
)

func main() {
	// Open the default Varnish Shared Memory file
	c := vago.Config{}
	v, err := vago.Open(&c)
	if err != nil {
		fmt.Println(err)
		return
	}
	v.Log("", vago.RAW, vago.COPT_TAIL|vago.COPT_BATCH, func(vxid uint32, tag, _type, data string) int {
		fmt.Printf("%10d %-14s %s %s\n", vxid, tag, _type, data)
		// -1 : Stop after it finds the first record
		// >= 0 : Nothing to do but wait
		return 0
	})
	v.Close()
}
```

Same for `varnishstat -1`:

```go
package main

import (
	"fmt"

	"github.com/phenomenes/vago"
)

func main() {
	// Open the default Varnish Shared Memory file
	c := vago.Config{}
	v, err := vago.Open(&c)
	if err != nil {
		fmt.Println(err)
		return
	}
	stats := v.Stats()
	for field, value := range stats {
		fmt.Printf("%-35s\t%12d\n", field, value)
	}
	v.Close()
}
```
