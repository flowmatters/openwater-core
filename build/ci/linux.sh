#!/bin/bash

set -e  # Exit on any error

sudo apt-get update
sudo apt-get install -y python3 python3-venv hdf5-tools libaec-dev cmake build-essential zlib1g-dev

# Build (or reuse a cached) thread-safe libhdf5 and point cgo at it.
# We build our own rather than depend on how the distro configured its
# package, and link it statically so the binaries are self-contained. See
# build/hdf5/build-hdf5.sh.
# The install directory is cached by GitHub Actions (see build-ow-core.yml).
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

"${HDF5_INSTALL_DIR}/bin/h5dump" --version 2>/dev/null || echo "h5dump not found in install"
