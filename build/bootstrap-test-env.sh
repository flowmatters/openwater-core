#!/bin/bash

set -e

python3 -c "import sys; print(sys.path)"
# ls -lh
# python3 -m venv ow-test
# ls -lh
# echo ow-test created
# source ow-test/$VENV_DIR/activate
# echo ow-test activated
pip --version
pip install wheel pytest
curl 'https://raw.githubusercontent.com/flowmatters/openwater/master/requirements.txt' > requirements.txt
pip install -r requirements.txt
pip install https://github.com/flowmatters/openwater/archive/master.zip

