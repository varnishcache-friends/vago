# vago v1.3.0 (01-Feb-2019)

## Changes

- Test against Varnish 6.0 and 6.1
- Test against go 1.9, 1.10 and 1.11

## Fixes

- #14 Re-attach to the cursor when varnishd restarts
- dispatchCallback should return a C.int

# vago v1.2.0 (17-Feb-2018)

## Changes

- #12 Support Varnish 5.2

# vago v1.0.1 (26-Mar-2017)

## Fixes

- #9 Handle abandoned and overrun cases
- #8 Make Close() goroutine safe

## Changes

- #10 Add support to handle cursor options
- #7 Add support for configuration parameters
- Improve error handling
- Test against go1.8

# vago v1.0.0 (01-Feb-2017)

- Initial release
