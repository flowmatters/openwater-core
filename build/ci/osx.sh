#!/bin/bash

set -e  # Exit on any error

brew update
brew install go cmake

# Build (or reuse a cached) thread-safe libhdf5 and point cgo at it.
# Homebrew's hdf5 formula does not enable thread-safety. See
# build/hdf5/build-hdf5.sh for the rationale.
export HDF5_INSTALL_DIR="${HDF5_INSTALL_DIR:-$HOME/hdf5-install}"
./build/hdf5/build-hdf5.sh

# all.sh sources compilation_vars.txt before building. Derive it from env.sh so
# CI and local developers get identical flags.
(
    # shellcheck disable=SC1091
    source ./build/hdf5/env.sh > /dev/null
    echo "export CGO_CFLAGS=\"${CGO_CFLAGS}\""
    echo "export CGO_LDFLAGS=\"${CGO_LDFLAGS}\""
) > compilation_vars.txt

echo '--- compilation_vars.txt ---'
cat compilation_vars.txt
