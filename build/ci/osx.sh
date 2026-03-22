#!/bin/bash

set -e  # Exit on any error

brew update
brew install go
brew install hdf5
# brew link hdf5
# echo "######################"
# sudo cp /opt/homebrew/include/* /usr/local/include/
# sudo cp /opt/homebrew/lib/* /usr/local/lib/
# arch=$(uname -i)
# if [[$arch == arm*]]; then
# sudo mkdir -p /usr/local
sudo ln -s /opt/homebrew/include /usr/local/include
sudo ln -s /opt/homebrew/lib /usr/local/lib

if [ "$STATIC_HDF5" = "1" ]; then
    echo "Configuring static HDF5 linking..."
    HDF5_LIB_DIR=$(brew --prefix hdf5)/lib

    # Remove shared libraries so -lhdf5 resolves to static archives
    rm -f "${HDF5_LIB_DIR}"/libhdf5*.dylib

    # HDF5 static archives depend on zlib
    echo "export CGO_LDFLAGS=\"-lz\"" > compilation_vars.txt
    echo "Static HDF5 linking configured (${HDF5_LIB_DIR})"
fi
# echo "usr"
# ls -a /usr/local/include | grep hdf5
# ls -a /usr/local/lib | grep hdf5
# echo "homebrew"
# ls -a /opt/homebrew/include | grep hdf5
# ls -a /opt/homebrew/lib | grep hdf5
# fi
# uname -a
# export PATH=$PATH:/opt/homebrew/include
# export PATH=$PATH:/opt/homebrew/lib
# brew install tree
# tree /
# find / -name "hdf5.h"
# export CGO_CFLAGS="-I/opt/homebrew/include"
# export CGO_LDFLAGS="-L/opt/homebrew/lib -lhdf5 -lhdf5_hl"
# export CFLAGS="-I/opt/homebrew/include/"
# export CPPFLAGS="-I/opt/homebrew/include/"
# export LDFLAGS="-L/opt/homebrew/lib"
# export C_INCLUDE_PATH=$C_INCLUDE_PATH:/opt/homebrew/include
# printenv CGO_FLAGS