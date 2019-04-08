#!/bin/bash

set -e

export CMD_PATH=github.com/flowmatters/openwater-core/cmd

for item in `ls ${PREFIX}cmd`
do
  echo $CMD_PATH/$item
  go build  $CMD_PATH/$item
  go install $CMD_PATH/$item
done

echo libopenwater
go build -buildmode=c-shared -o libopenwater.so github.com/flowmatters/openwater-core/libopenwater 

