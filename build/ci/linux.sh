#!/bin/bash

set -e  # Exit on any error

sudo apt-get update
sudo apt-get install -y python3 python3-venv hdf5-tools libaec-dev cmake build-essential zlib1g-dev

# Build HDF5 from source with thread-safety enabled.
# Ubuntu's packaged libhdf5 is not built with --enable-threadsafe, which
# means concurrent HDF5 calls (even on different files) crash in C code.
# Thread-safe HDF5 uses an internal recursive mutex that serializes at
# the C-call level rather than the Go-function level, significantly
# reducing contention between concurrent readers and writers in ow-sim.
#
# HDF5 1.14+ supports thread-safety with the HL library.
# The install directory is cached by GitHub Actions to avoid rebuilding
# on every push (see build-ow-core.yml).

HDF5_INSTALL_DIR="${HDF5_INSTALL_DIR:-$HOME/hdf5-install}"

if [ -f "${HDF5_INSTALL_DIR}/lib/libhdf5.a" ]; then
    echo "HDF5 found in cache at ${HDF5_INSTALL_DIR}"
else
    echo "Building HDF5 ${HDF5_VERSION:-1.14.6} from source (thread-safe, static)..."
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
    make -j$(nproc)
    make install
    cd ../..
    echo "HDF5 installed to ${HDF5_INSTALL_DIR}"
fi

if [ "$STATIC_HDF5" = "1" ]; then
    echo "Configuring static HDF5 linking..."
    echo "export CGO_CFLAGS=\"-I${HDF5_INSTALL_DIR}/include\"" > compilation_vars.txt
    echo "export CGO_LDFLAGS=\"-L${HDF5_INSTALL_DIR}/lib -lhdf5_hl -lhdf5 -lsz -lz -ldl -lm -lpthread\"" >> compilation_vars.txt
    echo "Static HDF5 linking configured (${HDF5_INSTALL_DIR})"
fi

"${HDF5_INSTALL_DIR}/bin/h5dump" --version 2>/dev/null || echo "h5dump not found in install"
