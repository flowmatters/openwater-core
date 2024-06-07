#!/bin/bash

brew update
brew install go
brew install hdf5
brew link hdf5
echo "######################"
# sudo cp /opt/homebrew/include/* /usr/local/include/
# sudo cp /opt/homebrew/lib/* /usr/local/lib/
sudo mkdir -p /usr/local
sudo ln -s /opt/homebrew/include /usr/local/include
sudo ln -s /opt/homebrew/lib /usr/local/lib
echo "usr"
ls -a /usr/local/include | grep hdf5
ls -a /usr/local/lib | grep hdf5
echo "homebrew"
ls -a /opt/homebrew/include | grep hdf5
ls -a /opt/homebrew/lib | grep hdf5
uname -a
export PATH=$PATH:/opt/homebrew/include
export PATH=$PATH:/opt/homebrew/lib
# brew install tree
# tree /
# find / -name "hdf5.h"
# export CGO_CFLAGS="-I/opt/homebrew/include"
# export CGO_LDFLAGS="-L/opt/homebrew/lib -lhdf5 -lhdf5_hl"
# export CFLAGS="-I/opt/homebrew/include/"
# export CPPFLAGS="-I/opt/homebrew/include/"
# export LDFLAGS="-L/opt/homebrew/lib"
# export C_INCLUDE_PATH=$C_INCLUDE_PATH:/opt/homebrew/include
printenv CGO_FLAGS