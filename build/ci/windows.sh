#!/bin/bash

curl -o hdf5.zip https://www.flowmatters.com.au/ow/1.10.5.zip
unzip hdf5.zip
echo export HDF5_DIR=$PWD/1.10.5 > compilation_vars.txt
echo export CGO_CFLAGS="-I$HDF5_DIR/include" >> compilation_vars.txt
echo export CGO_LDFLAGS="-L$HDF5_DIR/lib -lhdf5 -lhdf5_hl" 
cat compilation_vars.txt

source compilation_vars.txt
echo HDF5_DIR: $HDF5_DIR
echo CGO_CFLAGS: $CGO_CFLAGS
echo CGO_LDFLAGS: $CGO_LDFLAGS