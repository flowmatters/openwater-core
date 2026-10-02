#!/bin/bash

set -e  # Exit on any error

# Build (or reuse a cached) thread-safe libhdf5 and point cgo at it.
# build-hdf5.sh forces a shared build on Windows — MSVC static archives and
# cgo do not mix. The install directory is cached by GitHub Actions.
export HDF5_INSTALL_DIR="$(pwd)/hdf5-install"
./build/hdf5/build-hdf5.sh

HDF5_DIR_WIN=$(cd "${HDF5_INSTALL_DIR}" && pwd -W)

# all.sh sources compilation_vars.txt before building. The Windows toolchain
# needs the native (drive-letter) paths, so the flags are written directly
# rather than taken from env.sh.
{
    echo "export CGO_CFLAGS=\"-I${HDF5_DIR_WIN}/include\""
    echo "export CGO_LDFLAGS=\"-L${HDF5_DIR_WIN}/lib -lhdf5_hl -lhdf5\""
    echo "export PATH=\"${HDF5_INSTALL_DIR}/bin:\$PATH\""
    echo "export VENV_DIR=Scripts"
} > compilation_vars.txt

# Copy DLLs for artifact packaging
mkdir -p hdf5-dlls
cp "${HDF5_INSTALL_DIR}"/bin/*.dll hdf5-dlls/ 2>/dev/null || true

echo '--- compilation_vars.txt ---'
cat compilation_vars.txt
source compilation_vars.txt
