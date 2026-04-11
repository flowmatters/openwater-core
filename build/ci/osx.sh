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
        -DBUILD_SHARED_LIBS=OFF \
        -DBUILD_TESTING=OFF \
        -DCMAKE_C_FLAGS="-fPIC" \
        -DCMAKE_BUILD_TYPE=Release
    make -j$(sysctl -n hw.ncpu)
    make install
    cd ../..
    echo "HDF5 installed to ${HDF5_INSTALL_DIR}"
fi

# Point compiler/linker at the custom HDF5 install.
# gonum/hdf5 expects headers/libs in standard include/lib paths, so we
# symlink to /usr/local (macOS doesn't have a default HDF5 in /usr/local).
sudo ln -sf "${HDF5_INSTALL_DIR}/include" /usr/local/include
sudo ln -sf "${HDF5_INSTALL_DIR}/lib" /usr/local/lib

if [ "$STATIC_HDF5" = "1" ]; then
    echo "Configuring static HDF5 linking..."
    # Remove shared libs if any leaked through
    rm -f "${HDF5_INSTALL_DIR}/lib"/libhdf5*.dylib 2>/dev/null || true
    echo "export CGO_CFLAGS=\"-I${HDF5_INSTALL_DIR}/include\"" > compilation_vars.txt
    echo "export CGO_LDFLAGS=\"-L${HDF5_INSTALL_DIR}/lib -lhdf5_hl -lhdf5 -lz\"" >> compilation_vars.txt
    echo "Static HDF5 linking configured (${HDF5_INSTALL_DIR})"
fi
