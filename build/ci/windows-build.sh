#!/bin/bash

source compilation_vars.txt
export VENV_DIR=bin
python3 --version
./build/bootstrap-test-env.sh
OW_TEST_PATH=$PWD/test/files go test -v ./...
./build/build.sh
./build/run_test.sh
ls -al