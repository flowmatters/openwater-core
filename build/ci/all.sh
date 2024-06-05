if [[ "$OSTYPE" == "darwin"* ]]; then
    ./build/ci/osx.sh
    python3 -m venv ow-test
    source ow-test/bin/activate
elif [[ "$OSTYPE" == "linux"* ]]; then
    ./build/ci/linux.sh
    python3 -m venv ow-test
    source ow-test/bin/activate
else # Windows
    source compilation_vars.txt
fi
export VENV_DIR=bin
python3 --version
./build/bootstrap-test-env.sh
OW_TEST_PATH=$PWD/test/files go test -v ./...
./build/build.sh
./build/run_test.sh
ls -al