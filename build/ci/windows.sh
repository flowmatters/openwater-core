#!/bin/bash

# choco install python --version=3.7.2
# cp /c/Python37/python.exe /c/Python37/python3.exe

curl -o hdf5.zip https://www.flowmatters.com.au/ow/1.8.21.zip
unzip -q hdf5.zip
export HDF5_DIR=`pwd -W`/1.8.21
export HDF5_DIR_POSIX=`pwd`/1.8.21
echo export CGO_CFLAGS=\"-I$HDF5_DIR/include\" > compilation_vars.txt
echo export CGO_LDFLAGS=\"-L$HDF5_DIR/lib -lhdf5 -lhdf5_hl\"  >> compilation_vars.txt
# echo export PATH=\"/c/Python37:/c/Python37/Scripts:$HDF5_DIR_POSIX/bin:$PATH\" >> compilation_vars.txt
echo export VENV_DIR=Scripts >> compilation_vars.txt

echo '--- compilation_vars.txt ---'

cat compilation_vars.txt

source compilation_vars.txt
echo HDF5_DIR: $HDF5_DIR
echo CGO_CFLAGS: $CGO_CFLAGS
echo CGO_LDFLAGS: $CGO_LDFLAGS
echo $PATH

#python -m venv ow-test
# pip install virtualenv
# virtualenv ow-test