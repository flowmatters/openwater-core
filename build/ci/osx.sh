#!/bin/bash

brew update
brew install hdf5
brew install go
echo "######################"
ls -a /opt/homebrew/include
export PATH=$PATH:/opt/homebrew/include
# brew install tree
# tree /
# find / -name "hdf5.h"
export CGO_FLAGS="-I /opt/homebrew/include"
printenv CGO_FLAGS