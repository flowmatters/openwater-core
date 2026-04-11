#!/bin/bash

set -e  # Exit on any error

brew update
brew install go cmake

# Build HDF5 from source with thread-safety enabled.
# Homebrew's hdf5 formula does not enable thread-safety by default.
# See linux.sh for rationale.

HDF5_INSTALL_DIR="${HDF5_INSTALL_DIR:-$HOME/hdf5-install}"

if [ -f "${HDF5_INSTALL_DIR}/lib/libhdf5.a" ]; then
    echo "HDF5 found in cache at ${HDF5_INSTALL_DIR}"
else
    echo "Building HDF5 ${HDF5_VERSION:-1.14.6} from source (thread-safe)..."
    HDF5_VERSION="${HDF5_VERSION:-1.14.6}"
    curl -sL "https://github.com/HDFGroup/hdf5/releases/download/hdf5_${HDF5_VERSION}/hdf5-${HDF5_VERSION}.tar.gz" | tar xz
    cd "hdf5-${HDF5_VERSION}"
    mkdir -p build && cd build
    cmake .. \
        -DCMAKE_INSTALL_PREFIX="${HDF5_INSTALL_DIR}" \
        -DHDF5_ENABLE_THREADSAFE=ON \
        -DHDF5_BUILD_HL_LIB=ON \
        -DALLOW_UNSUPPORTED=ON \
        -DHDF5_ENABLE_SZIP_SUPPORT=OFF \
        -DHDF5_ENABLE_Z_LIB_SUPPORT=OFF \
        -DBUILD_SHARED_LIBS=OFF \
        -DBUILD_TESTING=OFF \
        -DCMAKE_C_FLAGS="-fPIC" \
        -DCMAKE_BUILD_TYPE=Release
    make -j$(sysctl -n hw.ncpu)
    make install
    cd ../..
    echo "HDF5 installed to ${HDF5_INSTALL_DIR}"
fi

# Point compiler/linker at the custom HDF5 install via CGO flags.
# We always write compilation_vars.txt (not just when STATIC_HDF5 is set)
# because the symlink approach is fragile on macOS runners.
echo "export CGO_CFLAGS=\"-I${HDF5_INSTALL_DIR}/include\"" > compilation_vars.txt
echo "export CGO_LDFLAGS=\"-L${HDF5_INSTALL_DIR}/lib -lhdf5_hl -lhdf5\"" >> compilation_vars.txt
echo "CGO flags configured for ${HDF5_INSTALL_DIR}"
