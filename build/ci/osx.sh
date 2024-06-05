#!/bin/bash

brew update
brew install hdf5
brew install go
echo "######################"
ls -a /usr/local/include
brew install tree
tree /usr
find /usr -name "hdf5.h"
export CPATH="/usr/include/"