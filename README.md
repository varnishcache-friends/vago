# vago

![ci](https://github.com/varnishcache-friends/vago/workflows/ci/badge.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/varnishcache-friends/vago.svg)](https://pkg.go.dev/github.com/varnishcache-friends/vago)

Go bindings for Varnish 6.0, 7.2 and 7.3.

Older Varnish versions are not supported.

## Requirements

To build this package you will need:

* pkg-config
* libvarnishapi-dev

You will also need to set PKG_CONFIG_PATH to the directory where `varnishapi.pc`
is located before running `go get`. For example:

```
export PKG_CONFIG_PATH=/usr/local/lib/pkgconfig
```

## Installation

```
go get github.com/varnishcache-friends/vago
```

## Examples

Same as running `varnishlog -g raw`:

```go
package main

import (
	"fmt"
	"log"

	"github.com/varnishcache-friends/vago"
)

func main() {
	// Open the default Varnish Shared Memory file
	c := vago.Config{}
	v, err := vago.Open(&c)
	if err != nil {
		log.Fatal(err)
	}
	v.Log("", vago.RAW, vago.COPT_TAIL|vago.COPT_BATCH, func(vxid uint64, tag, _type, data string) int {
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
	"log"

	"github.com/varnishcache-friends/vago"
)

func main() {
	// Open the default Varnish Shared Memory file
	c := vago.Config{}
	v, err := vago.Open(&c)
	if err != nil {
		log.Fatal(err)
	}
	stats := v.Stats()
	for field, value := range stats {
		fmt.Printf("%-35s\t%12d\n", field, value)
	}
	v.Close()
}
```
