#!/bin/bash

set -e  # Exit on any error

# Build HDF5 from source with thread-safety enabled.
# Previously this downloaded a pre-built HDF5 1.8.21 DLL (2018).
# Building from source gives us thread-safety and a current library version.
# The install directory is cached by GitHub Actions.

HDF5_INSTALL_DIR="$(pwd)/hdf5-install"

if [ -f "${HDF5_INSTALL_DIR}/lib/hdf5.lib" ] || [ -f "${HDF5_INSTALL_DIR}/lib/libhdf5.lib" ]; then
    echo "HDF5 found in cache at ${HDF5_INSTALL_DIR}"
else
    echo "Building HDF5 ${HDF5_VERSION:-1.14.6} from source (thread-safe, shared)..."
    HDF5_VERSION="${HDF5_VERSION:-1.14.6}"
    curl -sL "https://github.com/HDFGroup/hdf5/releases/download/hdf5_${HDF5_VERSION}/hdf5-${HDF5_VERSION}.tar.gz" | tar xz
    cd "hdf5-${HDF5_VERSION}"
    mkdir -p build && cd build
    cmake .. \
        -G "Visual Studio 17 2022" \
        -DCMAKE_INSTALL_PREFIX="${HDF5_INSTALL_DIR}" \
        -DHDF5_ENABLE_THREADSAFE=ON \
        -DHDF5_BUILD_HL_LIB=ON \
        -DALLOW_UNSUPPORTED=ON \
        -DHDF5_ENABLE_SZIP_SUPPORT=OFF \
        -DBUILD_SHARED_LIBS=ON \
        -DBUILD_TESTING=OFF \
        -DCMAKE_BUILD_TYPE=Release
    cmake --build . --config Release --parallel
    cmake --install . --config Release
    cd ../..
    echo "HDF5 installed to ${HDF5_INSTALL_DIR}"
fi

HDF5_DIR_WIN=$(cd "${HDF5_INSTALL_DIR}" && pwd -W)
HDF5_DIR_POSIX="${HDF5_INSTALL_DIR}"
echo "export CGO_CFLAGS=\"-I${HDF5_DIR_WIN}/include\"" > compilation_vars.txt
echo "export CGO_LDFLAGS=\"-L${HDF5_DIR_WIN}/lib -lhdf5 -lhdf5_hl\"" >> compilation_vars.txt
echo "export PATH=\"${HDF5_DIR_POSIX}/bin:\$PATH\"" >> compilation_vars.txt
echo "export VENV_DIR=Scripts" >> compilation_vars.txt

# Copy DLLs for artifact packaging
mkdir -p hdf5-dlls
cp "${HDF5_INSTALL_DIR}"/bin/*.dll hdf5-dlls/ 2>/dev/null || true

echo '--- compilation_vars.txt ---'
cat compilation_vars.txt
source compilation_vars.txt
