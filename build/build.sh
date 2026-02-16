#!/bin/bash

set -e
echo build.sh $PWD
export CMD_PATH=./cmd

BUILD_SHA=$(git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS="-X github.com/flowmatters/openwater-core/util.BuildSHA=${BUILD_SHA} -X github.com/flowmatters/openwater-core/util.BuildTime=${BUILD_TIME}"

cd pre/ow-specgen
go build .
cd ../..
./pre/ow-specgen/ow-specgen models/**/*
for item in `ls cmd`
do
  echo $CMD_PATH/$item
  go build -ldflags "$LDFLAGS" $CMD_PATH/$item
  go install -ldflags "$LDFLAGS" $CMD_PATH/$item
done

echo libopenwater
if [[ "$OSTYPE" == "msys"* || "$OSTYPE" == "cygwin"* || "$OSTYPE" == "win"* ]]; then
  LIB_EXT=dll
elif [[ "$OSTYPE" == "darwin"* ]]; then
  LIB_EXT=dylib
else
  LIB_EXT=so
fi
go build -ldflags "$LDFLAGS" -buildmode=c-shared -o libopenwater.$LIB_EXT ./libopenwater
mkdir -p ../bin
cp libopenwater.* ../bin

ls -lh
