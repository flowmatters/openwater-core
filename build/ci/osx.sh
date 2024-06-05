#!/bin/bash

brew update
brew install hdf5
brew install go
echo "######################"
ls -a /usr/local/include
find /usr -name "hdf5.h"
export CPATH="/usr/include/"