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
