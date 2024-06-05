#!/bin/bash

brew update
brew install go
brew install hdf5@1.8
brew link hdf5
echo "######################"
ls -a /opt/homebrew/include | grep hdf5
ls -a /opt/homebrew/lib | grep hdf5
uname -a
export PATH=$PATH:/opt/homebrew/include
export PATH=$PATH:/opt/homebrew/lib
# brew install tree
# tree /
# find / -name "hdf5.h"
export CGO_FLAGS="-I/opt/homebrew/include"
export CFLAGS="-I/opt/homebrew/include/"
export CPPFLAGS="-I/opt/homebrew/include/"
export LDFLAGS="-L/opt/homebrew/lib"
export C_INCLUDE_PATH=$C_INCLUDE_PATH:/opt/homebrew/include
printenv CGO_FLAGS