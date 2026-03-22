#!/bin/bash

set -e  # Exit on any error

sudo apt-get update
sudo apt-get install -y python3 libhdf5-serial-dev hdf5-tools python3-venv libaec-dev

if [ "$STATIC_HDF5" = "1" ]; then
    echo "Configuring static HDF5 linking..."
    ARCH=$(dpkg-architecture -qDEB_HOST_MULTIARCH 2>/dev/null || echo "x86_64-linux-gnu")
    HDF5_LIB_DIR="/usr/lib/${ARCH}/hdf5/serial"

    # Remove shared libraries so -lhdf5 resolves to static archives
    sudo rm -f "${HDF5_LIB_DIR}"/libhdf5.so* "${HDF5_LIB_DIR}"/libhdf5_hl.so*

    # Repeat hdf5_hl/hdf5 after the package's own -lhdf5 -lhdf5_hl to resolve
    # circular static archive dependencies (hdf5_hl references hdf5 symbols).
    # Also link szip (libaec) and other transitive deps of the static archives.
    echo "export CGO_LDFLAGS=\"-L${HDF5_LIB_DIR} -lhdf5_hl -lhdf5 -lsz -lz -ldl -lm -lpthread\"" > compilation_vars.txt
    echo "Static HDF5 linking configured (${HDF5_LIB_DIR})"
fi

h5ls --version