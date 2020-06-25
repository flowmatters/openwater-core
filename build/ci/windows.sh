#!/bin/bash

choco install python --version=3.7.2
ls -l /c/Python37
ls -l /c/Python37/bin

curl -o hdf5.zip https://www.flowmatters.com.au/ow/1.10.5.zip
unzip -q hdf5.zip
export HDF5_DIR=`pwd -W`/1.10.5
echo export CGO_CFLAGS=\"-I$HDF5_DIR/include\" > compilation_vars.txt
echo export CGO_LDFLAGS=\"-L$HDF5_DIR/lib -lhdf5 -lhdf5_hl\"  >> compilation_vars.txt
echo export PATH=\"$PATH:/c/Python37/bin\" >> compilation_vars.txt

echo '--- compilation_vars.txt ---'

cat compilation_vars.txt

source compilation_vars.txt
echo HDF5_DIR: $HDF5_DIR
echo CGO_CFLAGS: $CGO_CFLAGS
echo CGO_LDFLAGS: $CGO_LDFLAGS


echo $PATH