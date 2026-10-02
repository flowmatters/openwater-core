# openwater-core

![Build Status](https://travis-ci.org/flowmatters/openwater-core.svg?branch=master)

## Building

OpenWater links against HDF5 through cgo. A thread-safe libhdf5 is the
recommended configuration, and not every distro package is built that way — see
[docs/HDF5.md](docs/HDF5.md) for how to build one and switch between it and
the system library.
