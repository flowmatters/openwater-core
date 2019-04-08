#!/bin/bash

set -e

python3 -m venv .ow-test
source .ow-test/bin/activate
pip --version
curl 'https://raw.githubusercontent.com/flowmatters/openwater/master/requirements.txt' > requirements.txt
pip install -r requirements.txt
