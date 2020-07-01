#!/bin/bash

set -e
echo build.sh $PWD
export CMD_PATH=./cmd

for item in `ls cmd`
do
  echo $CMD_PATH/$item
  go build  $CMD_PATH/$item
  go install $CMD_PATH/$item
done

echo libopenwater
go build -buildmode=c-shared -o libopenwater.so ./libopenwater 

ls -lh
