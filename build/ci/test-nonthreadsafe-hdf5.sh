#!/bin/bash
#
# Run the io tests against a libhdf5 built WITHOUT thread-safety. The main
# build (all.sh) links the thread-safe library, so without this step nothing in
# CI exercises the Go-side mutex path in io/hdf5_util.go — including the
# lock-ordering guard, which skips itself when libhdf5 is thread-safe.
#
# The distro package is no substitute: Debian and Ubuntu build libhdf5-dev
# with thread-safety enabled. So build a second library with build-hdf5.sh into
# its own prefix. See docs/HDF5.md.
#
# Run from the repository root.

set -e  # Exit on any error

export HDF5_INSTALL_DIR="${HDF5_NONTHREADSAFE_INSTALL_DIR:-$HOME/hdf5-install-nonthreadsafe}"
HDF5_THREADSAFE=OFF ./build/hdf5/build-hdf5.sh

# shellcheck disable=SC1091
source ./build/hdf5/env.sh

OW_EXPECT_HDF5_THREADSAFE=false OW_TEST_PATH=$PWD/test/files go test -count=1 -v ./io/
